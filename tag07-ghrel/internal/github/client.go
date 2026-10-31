// Package github ist ein minimaler Client für die GitHub-Releases-API.
package github

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Release enthält die Felder, die ghrel anzeigt.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// APIError ist eine Antwort mit Status >= 400.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("GitHub-API: %d %s", e.Status, e.Message)
}

// Client spricht mit api.github.com (oder einem Testserver).
type Client struct {
	BaseURL string        // z. B. https://api.github.com
	Token   string        // optional; erhöht das Rate-Limit von 60 auf 5000/h
	HTTP    *http.Client  // mit Timeout pro Versuch
	Retries int           // zusätzliche Versuche bei 429/5xx/Netzfehlern
	Backoff time.Duration // Wartezeit vor dem 1. Retry, danach verdoppelt
}

// Releases liefert die neuesten n Releases von owner/repo.
func (c *Client) Releases(ctx context.Context, repo string, n int) ([]Release, error) {
	var rs []Release
	err := c.get(ctx, fmt.Sprintf("/repos/%s/releases?per_page=%d", repo, n), &rs)
	return rs, err
}

// Latest liefert das neueste stabile Release (keine Drafts/Prereleases).
func (c *Client) Latest(ctx context.Context, repo string) (Release, error) {
	var r Release
	err := c.get(ctx, "/repos/"+repo+"/releases/latest", &r)
	return r, err
}
