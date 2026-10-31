package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fake prüft den Key und streamt drei Textstücke per SSE.
func fake(t *testing.T) *Client {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer sk-lab-1234" {
				http.Error(w, "kein key", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			for _, s := range []string{"Hal", "lo", " Welt"} {
				fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":"+
					"{\"content\":%q}}]}\n\n", s)
			}
			fmt.Fprint(w, "data: [DONE]\n\n")
		}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Key: "sk-lab-1234", HTTP: srv.Client()}
}

func TestStream(t *testing.T) {
	var got strings.Builder
	err := fake(t).Stream(context.Background(), nil,
		func(s string) { got.WriteString(s) })
	if err != nil || got.String() != "Hallo Welt" {
		t.Fatalf("got %q, err %v", got.String(), err)
	}
}

func TestStatusError(t *testing.T) {
	c := fake(t)
	c.Key = "falsch"
	_, err := c.Chat(context.Background(), Request{})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("erwartet 401, got %v", err)
	}
}
