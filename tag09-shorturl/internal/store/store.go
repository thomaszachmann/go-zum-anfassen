// Package store definiert, wie Kurzlinks gespeichert werden. Tag 10 ersetzt den
// Speicher durch SQLite – Server und Tests bleiben unverändert.
package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("store: Link nicht gefunden")
	ErrExists   = errors.New("store: Code existiert bereits")
)

// Link ist ein Kurzlink.
type Link struct {
	Code    string    `json:"code"`
	URL     string    `json:"url"`
	Hits    int64     `json:"hits"`
	Created time.Time `json:"created"`
}

// Store ist der Vertrag zwischen HTTP-Schicht und Speicher.
type Store interface {
	Create(ctx context.Context, l Link) error           // ErrExists
	Get(ctx context.Context, code string) (Link, error) // ErrNotFound
	Hit(ctx context.Context, code string) (Link, error) // Hits+1; ErrNotFound
	Delete(ctx context.Context, code string) error      // ErrNotFound
}
