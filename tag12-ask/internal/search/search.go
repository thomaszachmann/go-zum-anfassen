// Package search findet die ähnlichsten Texte per Kosinus-Ähnlichkeit.
package search

import (
	"math"
	"sort"
)

type Doc struct {
	Source, Text string
	Vec          []float64
}

type Hit struct {
	Doc
	Score float64
}

// Cosine liefert 1 für gleiche Richtung, 0 für orthogonale Vektoren.
func Cosine(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// TopK liefert die k Dokumente mit der höchsten Ähnlichkeit zu q.
func TopK(docs []Doc, q []float64, k int) []Hit {
	hits := make([]Hit, 0, len(docs))
	for _, d := range docs {
		hits = append(hits, Hit{d, Cosine(d.Vec, q)})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	return hits[:min(k, len(hits))]
}
