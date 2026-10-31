package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// get holt path und dekodiert JSON nach v – mit Retries und Backoff.
func (c *Client) get(ctx context.Context, path string, v any) error {
	wait := c.Backoff
	for attempt := 0; ; attempt++ {
		retryAfter, err := c.do(ctx, path, v)
		if err == nil || attempt >= c.Retries || ctx.Err() != nil || !retryable(err) {
			return err
		}
		if retryAfter > 0 {
			wait = retryAfter // Server-Vorgabe schlägt eigene Schätzung
		}
		select {
		case <-ctx.Done(): // Gesamtbudget aufgebraucht oder Strg+C
			return fmt.Errorf("%w (letzter Fehler: %v)", ctx.Err(), err)
		case <-time.After(wait):
		}
		wait *= 2
	}
}

// do macht genau einen Versuch.
func (c *Client) do(ctx context.Context, path string, v any) (time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "ghrel-lab") // ohne User-Agent lehnt GitHub ab
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var body struct{ Message string }
		_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&body)
		secs, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return time.Duration(secs) * time.Second,
			&APIError{Status: resp.StatusCode, Message: body.Message}
	}
	return 0, json.NewDecoder(resp.Body).Decode(v)
}

// retryable: 429, 5xx und Netzfehler ja, übrige 4xx nein.
func retryable(err error) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.Status == http.StatusTooManyRequests || ae.Status >= 500
	}
	return true // Netzfehler: Verbindung abgelehnt, Reset, Timeout pro Versuch
}
