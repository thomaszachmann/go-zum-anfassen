package store_test

import (
	"testing"

	"example.com/shorturl/internal/store"
	"example.com/shorturl/internal/store/storetest"
)

func TestMemory(t *testing.T) {
	storetest.Run(t, func(*testing.T) store.Store { return store.NewMemory() })
}
