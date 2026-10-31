package search

import (
	"math"
	"testing"
)

func TestTopK(t *testing.T) {
	docs := []Doc{{Source: "a", Vec: []float64{1, 0}},
		{Source: "b", Vec: []float64{0.7, 0.7}}, {Source: "c", Vec: []float64{0, 1}}}
	got := TopK(docs, []float64{0, 2}, 2)
	if len(got) != 2 || got[0].Source != "c" || got[1].Source != "b" {
		t.Fatalf("got %+v", got)
	}
	if math.Abs(got[0].Score-1) > 1e-9 {
		t.Errorf("gleiche Richtung: Score %v, want 1", got[0].Score)
	}
	if Cosine([]float64{0, 0}, []float64{1, 1}) != 0 {
		t.Error("Nullvektor muss 0 liefern")
	}
}
