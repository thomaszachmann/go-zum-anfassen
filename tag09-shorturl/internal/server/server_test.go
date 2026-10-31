package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/shorturl/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(store.NewMemory(), log, "http://short.test", prometheus.NewRegistry())
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// noRedirect: Der Test will die 302 sehen, nicht dem Link folgen.
var noRedirect = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func do(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

const godev = `{"url":"https://go.dev","code":"godev"}`

func TestLebenszyklus(t *testing.T) {
	srv := newTestServer(t)
	resp := do(t, "POST", srv.URL+"/api/links", godev)
	var created createResp
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil ||
		resp.StatusCode != 201 || created.Short != "http://short.test/godev" {
		t.Fatalf("create: %d %+v %v", resp.StatusCode, created, err)
	}

	resp = do(t, "GET", srv.URL+"/godev", "")
	if resp.StatusCode != 302 || resp.Header.Get("Location") != "https://go.dev" {
		t.Errorf("redirect: %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	var link store.Link
	resp = do(t, "GET", srv.URL+"/api/links/godev", "")
	if err := json.NewDecoder(resp.Body).Decode(&link); err != nil || link.Hits != 1 {
		t.Errorf("stats: %+v %v", link, err)
	}

	if resp = do(t, "DELETE", srv.URL+"/api/links/godev", ""); resp.StatusCode != 204 {
		t.Errorf("delete: %d", resp.StatusCode)
	}
	if resp = do(t, "GET", srv.URL+"/godev", ""); resp.StatusCode != 404 {
		t.Errorf("nach delete: %d", resp.StatusCode)
	}
}

func TestFehlerfaelle(t *testing.T) {
	srv := newTestServer(t)
	do(t, "POST", srv.URL+"/api/links", godev)
	tests := []struct {
		name, method, path, body string
		want                     int
	}{
		{"kein JSON", "POST", "/api/links", `url=x`, 400},
		{"unbekanntes Feld", "POST", "/api/links", `{"url":"https://a.b","x":1}`, 400},
		{"ftp-URL", "POST", "/api/links", `{"url":"ftp://a.b"}`, 400},
		{"reserviert", "POST", "/api/links", `{"url":"https://a.b","code":"api"}`, 400},
		{"Code belegt", "POST", "/api/links", godev, 409},
		{"falsche Methode", "PUT", "/api/links/godev", ``, 405},
		{"unbekannt", "GET", "/api/links/nix", ``, 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, tt.method, srv.URL+tt.path, tt.body)
			if resp.StatusCode != tt.want {
				t.Errorf("Status %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestZufallscodeUndMetriken(t *testing.T) {
	srv := newTestServer(t)
	resp := do(t, "POST", srv.URL+"/api/links", `{"url":"https://kubernetes.io"}`)
	var created createResp
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil ||
		!validCode.MatchString(created.Code) {
		t.Fatalf("Code %q, %v", created.Code, err)
	}
	do(t, "GET", srv.URL+"/"+created.Code, "")

	body, _ := io.ReadAll(do(t, "GET", srv.URL+"/metrics", "").Body)
	want := `shorturl_http_requests_total{code="302",route="GET /{code}"} 1`
	if !strings.Contains(string(body), want) {
		t.Errorf("Metrik fehlt: %s", want)
	}
}
