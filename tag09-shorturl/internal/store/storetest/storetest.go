// Package storetest prüft jede Store-Implementierung gegen denselben Vertrag.
// Tag 10 ruft Run mit SQLite auf.
package storetest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/shorturl/internal/store"
)

// Run führt alle Vertragstests aus; newStore liefert pro Test einen leeren Store.
func Run(t *testing.T, newStore func(t *testing.T) store.Store) {
	ctx := context.Background()
	link := store.Link{Code: "go", URL: "https://go.dev",
		Created: time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)}

	t.Run("CreateGet", func(t *testing.T) {
		s := newStore(t)
		if err := s.Create(ctx, link); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(ctx, "go")
		if err != nil || got.URL != link.URL || !got.Created.Equal(link.Created) {
			t.Errorf("Get = %+v, %v", got, err)
		}
		if err := s.Create(ctx, link); !errors.Is(err, store.ErrExists) {
			t.Errorf("doppeltes Create: err = %v, want ErrExists", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		s := newStore(t)
		if _, err := s.Get(ctx, "nix"); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("Get: %v", err)
		}
		if _, err := s.Hit(ctx, "nix"); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("Hit: %v", err)
		}
		if err := s.Delete(ctx, "nix"); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("Delete: %v", err)
		}
	})

	t.Run("HitsParallel", func(t *testing.T) {
		s := newStore(t)
		if err := s.Create(ctx, link); err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() {
				if _, err := s.Hit(ctx, "go"); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if got, _ := s.Get(ctx, "go"); got.Hits != 50 {
			t.Errorf("Hits = %d, want 50", got.Hits)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		s := newStore(t)
		if err := s.Create(ctx, link); err != nil {
			t.Fatal(err)
		}
		if err := s.Delete(ctx, "go"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Get(ctx, "go"); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("nach Delete: %v", err)
		}
	})
}
