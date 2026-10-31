package main

import (
	"cmp"
	"slices"
)

// wordCount ist ein Paar aus Wort und Häufigkeit.
type wordCount struct {
	word string
	n    int
}

// topWords liefert die n häufigsten Wörter, bei Gleichstand alphabetisch.
func topWords(freq map[string]int, n int) []wordCount {
	list := make([]wordCount, 0, len(freq))
	for w, k := range freq {
		list = append(list, wordCount{w, k})
	}
	slices.SortFunc(list, func(a, b wordCount) int {
		if c := cmp.Compare(b.n, a.n); c != 0 {
			return c // absteigend nach Häufigkeit
		}
		return cmp.Compare(a.word, b.word)
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
