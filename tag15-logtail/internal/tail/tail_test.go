package tail

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuf ist ein Writer, den Test und Follow-Goroutine gefahrlos teilen.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// waitFor wartet bis zu 2 s, bis out mit suffix endet.
func waitFor(t *testing.T, out *syncBuf, suffix string) {
	t.Helper()
	for range 200 {
		if strings.HasSuffix(out.String(), suffix) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Ausgabe endet nicht mit %q:\n%s", suffix, out.String())
}

func appendTo(t *testing.T, path, s string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}

func TestLast(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	appendTo(t, path, "a\nb\nc\nd\ne\n")
	lines, off, err := Last(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(lines, []string{"c", "d", "e"}) || off != 10 {
		t.Errorf("Last = %q, %d; will [c d e], 10", lines, off)
	}
}

func TestFollow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	appendTo(t, path, "alt\n")
	ctx, cancel := context.WithCancel(context.Background())
	out := &syncBuf{}
	done := make(chan error, 1)
	go func() { done <- Follow(ctx, path, 4, 10*time.Millisecond, out) }()

	appendTo(t, path, "eins\nzw")
	time.Sleep(50 * time.Millisecond) // halbe Zeile darf noch nicht erscheinen
	if got := out.String(); got != "eins\n" {
		t.Errorf("nach halber Zeile: %q", got)
	}
	appendTo(t, path, "ei\n")
	waitFor(t, out, "eins\nzwei\n")

	// copytruncate: Datei wird kleiner
	if err := os.WriteFile(path, []byte("neu\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, out, "zwei\nneu\n")

	// Rotation: alte Datei umbenennen, neue anlegen
	appendTo(t, path, "letzte alte\n")
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	appendTo(t, path, "rotiert\n")
	waitFor(t, out, "neu\nletzte alte\nrotiert\n")

	cancel()
	if err := <-done; err != nil {
		t.Errorf("Follow nach cancel: %v", err)
	}
}
