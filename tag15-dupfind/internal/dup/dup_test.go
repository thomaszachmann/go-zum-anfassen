package dup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// tree legt Dateien relativ zu einem TempDir an und liefert dessen Pfad.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestFind(t *testing.T) {
	root := tree(t, map[string]string{
		"a.txt":     "hallo",
		"sub/b.txt": "hallo",  // Duplikat von a.txt
		"c.txt":     "hallx",  // gleiche Größe, anderer Inhalt
		"d.txt":     "hallo!", // andere Größe
		"leer1":     "",       // leere Dateien zählen nicht
		"leer2":     "",
	})
	if err := os.Symlink("a.txt", filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	for _, workers := range []int{1, 8} {
		got, err := Find(context.Background(), root, workers)
		if err != nil {
			t.Fatal(err)
		}
		want := []Group{{
			Size: 5,
			Hash: "d3751d33f9cd5049c4af2b462735457e4d3baf130bcbb87f389e349fbaeb20b9",
			Paths: []string{filepath.Join(root, "a.txt"),
				filepath.Join(root, "sub", "b.txt")},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("workers=%d:\n got %+v\nwant %+v", workers, got, want)
		}
	}
}

func TestFindSortierung(t *testing.T) {
	root := tree(t, map[string]string{
		"k1": "kurz", "k2": "kurz",
		"l1": "deutlich länger", "l2": "deutlich länger", "l3": "deutlich länger",
	})
	got, err := Find(context.Background(), root, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || len(got[0].Paths) != 3 || got[0].Wasted() != 2*16 {
		t.Errorf("größte Verschwendung zuerst erwartet: %+v", got)
	}
}

func TestFindAbbruch(t *testing.T) {
	root := tree(t, map[string]string{"a": "x", "b": "x"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Find(ctx, root, 2); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, will context.Canceled", err)
	}
}

func TestFindFehlenderOrdner(t *testing.T) {
	_, err := Find(context.Background(), "/gibt/es/nicht", 2)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, will os.ErrNotExist", err)
	}
}
