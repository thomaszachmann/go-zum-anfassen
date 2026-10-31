package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient startet einen Fake-GitHub; winziger Backoff hält Tests schnell.
func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Token: "lab-token", HTTP: srv.Client(),
		Retries: 3, Backoff: time.Millisecond}
}

func TestReleasesHeaderUndJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/tool/releases" ||
			r.URL.Query().Get("per_page") != "2" {
			t.Errorf("falsche URL: %s", r.URL)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer lab-token" {
			t.Errorf("Authorization = %q", got)
		}
		fmt.Fprint(w, `[{"tag_name":"v1.2.0","prerelease":false,
			"published_at":"2026-10-01T12:00:00Z"},{"tag_name":"v1.1.0"}]`)
	})
	rs, err := c.Releases(context.Background(), "acme/tool", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 2 || rs[0].TagName != "v1.2.0" || rs[0].PublishedAt.Day() != 1 {
		t.Errorf("Releases = %+v", rs)
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name      string
		failWith  int // Status der ersten beiden Antworten
		wantCalls int32
		wantErr   bool
	}{
		{"503 wird wiederholt", http.StatusServiceUnavailable, 3, false},
		{"429 wird wiederholt", http.StatusTooManyRequests, 3, false},
		{"404 sofort Fehler", http.StatusNotFound, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32 // Handler laufen in eigenen Goroutines
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) <= 2 {
					w.Header().Set("Retry-After", "0")
					http.Error(w, `{"message":"kaputt"}`, tt.failWith)
					return
				}
				fmt.Fprint(w, `{"tag_name":"v1.0.0"}`)
			})
			_, err := c.Latest(context.Background(), "acme/tool")
			if (err != nil) != tt.wantErr || calls.Load() != tt.wantCalls {
				t.Errorf("err=%v calls=%d, want err=%v calls=%d",
					err, calls.Load(), tt.wantErr, tt.wantCalls)
			}
		})
	}
}

func TestAPIErrorPerErrorsAs(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	})
	_, err := c.Latest(context.Background(), "acme/fehlt")
	var ae *APIError
	if !errors.As(err, &ae) || ae.Status != 404 || ae.Message != "Not Found" {
		t.Errorf("err = %v", err)
	}
}

func TestGesamtTimeout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select { // hängt, bis der Client aufgibt
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := c.Latest(ctx, "acme/tool")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("hat %v gedauert – Context ignoriert?", d)
	}
}
