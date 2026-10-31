package store

import (
	"context"
	"sync"
)

// Memory hält Links in einer Map. Nach einem Neustart ist alles weg (bis Tag 10).
type Memory struct {
	mu    sync.RWMutex // viele Leser (Redirects), wenige Schreiber
	links map[string]Link
}

var _ Store = (*Memory)(nil) // Compiler prüft: Memory erfüllt Store

// NewMemory liefert einen leeren Speicher.
func NewMemory() *Memory {
	return &Memory{links: make(map[string]Link)}
}

func (m *Memory) Create(_ context.Context, l Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[l.Code]; ok {
		return ErrExists
	}
	m.links[l.Code] = l
	return nil
}

func (m *Memory) Get(_ context.Context, code string) (Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return l, nil
}

func (m *Memory) Hit(_ context.Context, code string) (Link, error) {
	m.mu.Lock() // schreibt: Hits erhöhen
	defer m.mu.Unlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	l.Hits++
	m.links[code] = l
	return l, nil
}

func (m *Memory) Delete(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[code]; !ok {
		return ErrNotFound
	}
	delete(m.links, code)
	return nil
}
