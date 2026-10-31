package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// slowServer antwortet nach delay; peak() liefert die maximale Parallelität.
func slowServer(t *testing.T, delay time.Duration) (*httptest.Server, func() int) {
	t.Helper()
	var mu sync.Mutex
	cur, peak := 0, 0
	h := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cur++
		peak = max(peak, cur)
		mu.Unlock()
		time.Sleep(delay)
		mu.Lock()
		cur--
		mu.Unlock()
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	t.Cleanup(srv.Close)
	return srv, func() int { mu.Lock(); defer mu.Unlock(); return peak }
}

func urls(base string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = base + "/ok"
	}
	return out
}

func TestRunReihenfolge(t *testing.T) {
	srv, _ := slowServer(t, 0)
	c := &Checker{HTTP: srv.Client(), Workers: 3}
	in := []string{srv.URL + "/ok", srv.URL + "/fail", "http://127.0.0.1:9/zu"}
	res := c.Run(context.Background(), in)
	for i, want := range []bool{true, false, false} {
		if res[i].URL != in[i] || res[i].OK() != want {
			t.Errorf("res[%d] = %+v, want OK=%v", i, res[i], want)
		}
	}
}

func TestRunWorkerLimit(t *testing.T) {
	srv, peak := slowServer(t, 50*time.Millisecond)
	c := &Checker{HTTP: srv.Client(), Workers: 5}
	start := time.Now()
	c.Run(context.Background(), urls(srv.URL, 20))
	if p := peak(); p != 5 {
		t.Errorf("max. parallel = %d, want 5", p)
	}
	if d := time.Since(start); d > 500*time.Millisecond { // seriell: 1 s
		t.Errorf("Dauer %v – läuft das parallel?", d)
	}
}

func TestRunRateLimit(t *testing.T) {
	srv, _ := slowServer(t, 0)
	c := &Checker{HTTP: srv.Client(), Workers: 10, RPS: 50} // alle 20 ms
	start := time.Now()
	c.Run(context.Background(), urls(srv.URL, 10))
	if d := time.Since(start); d < 180*time.Millisecond {
		t.Errorf("10 Anfragen in %v – Rate-Limit greift nicht", d)
	}
}

func TestRunAbbruch(t *testing.T) {
	srv, _ := slowServer(t, time.Second)
	c := &Checker{HTTP: srv.Client(), Workers: 2, RPS: 1}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	for _, r := range c.Run(ctx, urls(srv.URL, 6)) {
		if r.OK() {
			t.Errorf("%s OK trotz Abbruch", r.URL)
		}
	}
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Errorf("Abbruch dauerte %v", d)
	}
}
