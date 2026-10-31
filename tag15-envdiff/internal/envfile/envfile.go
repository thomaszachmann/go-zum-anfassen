// Package envfile liest .env-Dateien und vergleicht sie.
package envfile

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

var keyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Parse liest KEY=VALUE-Zeilen mit #-Kommentaren, "export " und Quotes.
// Doppelte Schlüssel: der letzte gewinnt (wie bei source in der Shell).
func Parse(r io.Reader) (map[string]string, error) {
	env := map[string]string{}
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		if !ok || !keyRe.MatchString(key) {
			return nil, fmt.Errorf("zeile %d: erwarte KEY=VALUE, nicht %q", n, line)
		}
		env[key] = value(val)
	}
	return env, sc.Err()
}

// value entfernt Quotes; ohne Quotes beginnt bei " #" ein Kommentar.
func value(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return v
}

// Kind ist die Art einer Änderung.
type Kind string

const (
	Added   Kind = "+"
	Removed Kind = "-"
	Changed Kind = "~"
)

// Change ist ein Unterschied zwischen zwei Dateien.
type Change struct {
	Key      string
	Kind     Kind
	Old, New string
}

// Diff liefert alle Unterschiede von a nach b, sortiert nach Schlüssel.
func Diff(a, b map[string]string) []Change {
	keys := slices.Sorted(maps.Keys(a))
	for k := range b {
		if _, ok := a[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	var out []Change
	for _, k := range keys {
		va, inA := a[k]
		vb, inB := b[k]
		switch {
		case !inB:
			out = append(out, Change{Key: k, Kind: Removed, Old: va})
		case !inA:
			out = append(out, Change{Key: k, Kind: Added, New: vb})
		case va != vb:
			out = append(out, Change{Key: k, Kind: Changed, Old: va, New: vb})
		}
	}
	return out
}

// Mask verrät nur die Länge eines Werts – Secrets bleiben unsichtbar.
func Mask(v string) string {
	if v == "" {
		return `""`
	}
	return fmt.Sprintf("<%d Zeichen>", utf8.RuneCountInString(v))
}
