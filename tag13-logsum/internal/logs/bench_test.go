package logs

import (
	"bytes"
	"os"
	"testing"
)

// load liest das große Log aus Drill 1 einmal pro Benchmark.
func load(b *testing.B) []byte {
	data, err := os.ReadFile("../../access.log")
	if err != nil {
		b.Skip("erst access.log erzeugen (Drill 1)")
	}
	b.SetBytes(int64(len(data))) // zeigt zusätzlich MB/s
	return data
}

func BenchmarkRead(b *testing.B) {
	data := load(b)
	for b.Loop() {
		if _, err := Read(bytes.NewReader(data), Combined{}); err != nil {
			b.Fatal(err)
		}
	}
}
