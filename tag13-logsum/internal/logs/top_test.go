package logs

import (
	"slices"
	"testing"
)

func TestTopN(t *testing.T) {
	paths := TopN(map[string]int{"/b": 2, "/a": 2, "/c": 5, "/d": 1}, 3)
	want := []Pair[string]{{"/c", 5}, {"/a", 2}, {"/b", 2}}
	if !slices.Equal(paths, want) {
		t.Errorf("got %v, want %v", paths, want)
	}
	codes := TopN(map[int]int{200: 7, 500: 9, 404: 1}, 5) // K = int
	if !slices.Equal(codes, []Pair[int]{{500, 9}, {200, 7}, {404, 1}}) {
		t.Errorf("got %v", codes)
	}
}
