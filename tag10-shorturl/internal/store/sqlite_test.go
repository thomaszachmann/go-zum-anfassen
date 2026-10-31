package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"example.com/shorturl/internal/store"
	"example.com/shorturl/internal/store/storetest"
)

func TestSQLite(t *testing.T) {
	storetest.Run(t, func(t *testing.T) store.Store {
		s, err := store.OpenSQLite(context.Background(),
			filepath.Join(t.TempDir(), "test.db"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { s.Close() })
		return s
	})
}
