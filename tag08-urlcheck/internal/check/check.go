// Package check prüft URLs nebenläufig mit Worker-Pool und Rate-Limit.
package check

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Result ist das Ergebnis für eine URL.
type Result struct {
	URL      string        `json:"url"`
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration_ns"`
	Err      string        `json:"error,omitempty"`
}

// OK: Antwort erhalten und Status < 400.
func (r Result) OK() bool { return r.Err == "" && r.Status < 400 }

// Checker hält die Einstellungen; der Nullwert ist nicht nutzbar.
type Checker struct {
	HTTP    *http.Client
	Workers int     // gleichzeitige Anfragen
	RPS     float64 // Anfragen pro Sekunde gesamt, 0 = unbegrenzt
}

// One prüft genau eine URL.
func (c *Checker) One(ctx context.Context, url string) Result {
	start := time.Now()
	res := Result{URL: url}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		res.Err = err.Error()
		return res
	}
	req.Header.Set("User-Agent", "urlcheck-lab")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		res.Err = err.Error()
		res.Duration = time.Since(start)
		return res
	}
	defer resp.Body.Close()
	// Body (begrenzt) lesen, damit die Verbindung wiederverwendet werden kann.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	res.Status = resp.StatusCode
	res.Duration = time.Since(start)
	return res
}
