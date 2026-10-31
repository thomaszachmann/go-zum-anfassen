// Package dup findet doppelte Dateien per SHA-256 – mit Worker-Pool.
package dup

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// Group sind Dateien mit identischem Inhalt.
type Group struct {
	Size  int64
	Hash  string   // SHA-256, hex
	Paths []string // sortiert
}

// Wasted ist der Platz, den die Kopien zusätzlich belegen.
func (g Group) Wasted() int64 { return g.Size * int64(len(g.Paths)-1) }

// Find durchsucht root. Gehasht werden nur Dateien, deren Größe mehrfach vorkommt.
func Find(ctx context.Context, root string, workers int) ([]Group, error) {
	bySize := map[int64][]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !d.Type().IsRegular() { // Ordner, Symlinks, Sockets
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > 0 {
			bySize[info.Size()] = append(bySize[info.Size()], p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, ps := range bySize {
		if len(ps) > 1 {
			paths = append(paths, ps...)
		}
	}
	sums, err := hashAll(ctx, paths, workers)
	if err != nil {
		return nil, err
	}

	byHash := map[string]*Group{}
	for i, p := range paths {
		g := byHash[sums[i]]
		if g == nil {
			g = &Group{Hash: sums[i]}
			byHash[sums[i]] = g
		}
		g.Paths = append(g.Paths, p)
	}
	var groups []Group
	for _, g := range byHash {
		if len(g.Paths) < 2 {
			continue
		}
		slices.Sort(g.Paths)
		info, err := os.Stat(g.Paths[0])
		if err != nil {
			return nil, err
		}
		g.Size = info.Size()
		groups = append(groups, *g)
	}
	slices.SortFunc(groups, func(a, b Group) int {
		return cmp.Or(cmp.Compare(b.Wasted(), a.Wasted()),
			cmp.Compare(a.Paths[0], b.Paths[0]))
	})
	return groups, nil
}

// hashAll hasht paths mit workers Goroutines; jeder schreibt nur seinen Index.
func hashAll(ctx context.Context, paths []string, workers int) ([]string, error) {
	sums := make([]string, len(paths))
	errs := make([]error, len(paths))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range max(workers, 1) {
		wg.Go(func() {
			for i := range jobs {
				sums[i], errs[i] = hashFile(paths[i])
			}
		})
	}
feed:
	for i := range paths {
		select {
		case jobs <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return sums, errors.Join(errs...)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
