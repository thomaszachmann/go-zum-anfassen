package logs

import (
	"cmp"
	"slices"
)

// Pair ist ein Eintrag einer Top-Liste – für jeden Schlüsseltyp.
type Pair[K cmp.Ordered] struct {
	Key K
	N   int
}

// TopN liefert die n häufigsten Schlüssel, bei Gleichstand aufsteigend.
func TopN[K cmp.Ordered](m map[K]int, n int) []Pair[K] {
	list := make([]Pair[K], 0, len(m))
	for k, v := range m {
		list = append(list, Pair[K]{k, v})
	}
	slices.SortFunc(list, func(a, b Pair[K]) int {
		return cmp.Or(cmp.Compare(b.N, a.N), cmp.Compare(a.Key, b.Key))
	})
	return list[:min(n, len(list))]
}
