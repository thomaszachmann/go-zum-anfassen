// Package query wertet einfache jq-Pfade aus: . .a .a.b .a[0] .a[-1] .a[]
package query

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Sentinels: Syntaxfehler im Ausdruck bzw. falscher Typ im Dokument.
var (
	ErrSyntax = errors.New("syntaxfehler")
	ErrType   = errors.New("typfehler")
)

type kind int

const (
	field kind = iota // .name
	index             // [N]
	iter              // []
)

type step struct {
	kind kind
	key  string
	n    int
}

// parse zerlegt ".items[0].name" in Schritte.
func parse(expr string) ([]step, error) {
	rest, ok := strings.CutPrefix(expr, ".")
	if !ok {
		return nil, fmt.Errorf("%w: %q muss mit . beginnen", ErrSyntax, expr)
	}
	var steps []step
	for rest != "" {
		switch rest[0] {
		case '[':
			inner, after, ok := strings.Cut(rest[1:], "]")
			if !ok {
				return nil, fmt.Errorf("%w: ] fehlt in %q", ErrSyntax, expr)
			}
			if inner == "" {
				steps = append(steps, step{kind: iter})
			} else {
				n, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("%w: Index %q", ErrSyntax, inner)
				}
				steps = append(steps, step{kind: index, n: n})
			}
			rest = after
		case '.':
			rest = rest[1:]
			if rest == "" || rest[0] == '.' {
				return nil, fmt.Errorf("%w: leerer Schlüssel in %q", ErrSyntax, expr)
			}
		default:
			end := strings.IndexAny(rest, ".[")
			if end < 0 {
				end = len(rest)
			}
			steps = append(steps, step{kind: field, key: rest[:end]})
			rest = rest[end:]
		}
	}
	return steps, nil
}

// Eval wendet expr auf v an (Werte wie von encoding/json: map, slice, ...).
func Eval(v any, expr string) ([]any, error) {
	steps, err := parse(expr)
	if err != nil {
		return nil, err
	}
	cur := []any{v}
	for _, s := range steps {
		var next []any
		for _, x := range cur {
			out, err := apply(x, s)
			if err != nil {
				return nil, err
			}
			next = append(next, out...)
		}
		cur = next
	}
	return cur, nil
}

// apply führt einen Schritt aus. Wie jq: .x auf null oder Fehlendes ist null.
func apply(x any, s step) ([]any, error) {
	switch s.kind {
	case field:
		if x == nil {
			return []any{nil}, nil
		}
		m, ok := x.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: .%s auf %T", ErrType, s.key, x)
		}
		return []any{m[s.key]}, nil
	case index:
		if x == nil {
			return []any{nil}, nil
		}
		a, ok := x.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: [%d] auf %T", ErrType, s.n, x)
		}
		i := s.n
		if i < 0 {
			i += len(a) // [-1] ist das letzte Element
		}
		if i < 0 || i >= len(a) {
			return []any{nil}, nil
		}
		return []any{a[i]}, nil
	}
	switch t := x.(type) { // iter
	case []any:
		return t, nil
	case map[string]any:
		out := make([]any, 0, len(t))
		for _, k := range slices.Sorted(maps.Keys(t)) {
			out = append(out, t[k])
		}
		return out, nil
	}
	return nil, fmt.Errorf("%w: [] auf %T", ErrType, x)
}
