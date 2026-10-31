// Package scan sucht Secrets in Textdateien.
package scan

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

// Finding ist ein Treffer. Match ist maskiert – nie das ganze Secret.
type Finding struct {
	File  string `json:"file"`
	Line  int    `json:"line"`
	Rule  string `json:"rule"`
	Match string `json:"match"`
}

type rule struct {
	name string
	re   *regexp.Regexp
}

var rules = []rule{
	{"aws-access-key", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"github-token", regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36}\b`)},
	{"private-key", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"api-key", regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{20,}`)},
}

const (
	ignoreMarker = "secretscan:ignore"
	maxSize      = 1 << 20 // größere Dateien sind selten Konfiguration
)

var skipDirs = []string{".git", "node_modules", "vendor"}

// Mask zeigt die ersten vier Zeichen und die Länge.
func Mask(s string) string {
	n := utf8.RuneCountInString(s)
	if n <= 4 {
		return fmt.Sprintf("…(%d)", n)
	}
	return fmt.Sprintf("%s…(%d)", string([]rune(s)[:4]), n)
}

// Line prüft eine Zeile gegen alle Regeln, Treffer nach Position sortiert.
func Line(line string) []Finding {
	if strings.Contains(line, ignoreMarker) {
		return nil
	}
	type hit struct {
		pos int
		f   Finding
	}
	var hits []hit
	for _, r := range rules {
		for _, loc := range r.re.FindAllStringIndex(line, -1) {
			m := line[loc[0]:loc[1]]
			hits = append(hits, hit{loc[0], Finding{Rule: r.name, Match: Mask(m)}})
		}
	}
	slices.SortFunc(hits, func(a, b hit) int { return cmp.Compare(a.pos, b.pos) })
	var out []Finding
	for _, h := range hits {
		out = append(out, h.f)
	}
	return out
}

// File prüft eine Datei. Binärdateien (NUL-Byte vorn) liefern nichts.
func File(path string) ([]Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(data[:min(len(data), 512)], 0) >= 0 {
		return nil, nil
	}
	var out []Finding
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(nil, maxSize)
	for n := 1; sc.Scan(); n++ {
		for _, f := range Line(sc.Text()) {
			f.Line = n
			out = append(out, f)
		}
	}
	return out, sc.Err()
}

// Dir durchsucht root mit workers Goroutines; Dateinamen relativ zu root.
func Dir(ctx context.Context, root string, workers int) ([]Finding, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() && slices.Contains(skipDirs, d.Name()) {
			return filepath.SkipDir
		}
		if !d.Type().IsRegular() {
			return nil
		}
		// zu groß: überspringen (err ist dann nil)
		if info, err := d.Info(); err != nil || info.Size() > maxSize {
			return err
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return nil, err
	}

	results := make([][]Finding, len(files))
	errs := make([]error, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range max(workers, 1) {
		wg.Go(func() {
			for i := range jobs {
				results[i], errs[i] = File(files[i])
			}
		})
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	var out []Finding
	for i, found := range results {
		rel, err := filepath.Rel(root, files[i])
		if err != nil {
			return nil, err
		}
		for _, f := range found {
			f.File = filepath.ToSlash(rel)
			out = append(out, f)
		}
	}
	slices.SortStableFunc(out, func(a, b Finding) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
	return out, nil
}
