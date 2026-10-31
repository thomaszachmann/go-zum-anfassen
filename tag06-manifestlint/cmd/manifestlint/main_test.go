package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "Golden Files neu schreiben")

func TestJSONGolden(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{"-o", "json", "testdata/web.yaml"}, &out, io.Discard)
	if code != 1 {
		t.Errorf("Exit-Code %d, want 1", code)
	}
	golden := "testdata/web.golden.json"
	if *update {
		if err := os.WriteFile(golden, out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Errorf("Ausgabe weicht ab:\n%s", out.String())
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"ohne Pfad", nil, 2},
		{"fehlt", []string{"gibt-es-nicht.yaml"}, 2},
		{"Format", []string{"-o", "xml", "testdata/web.yaml"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(tt.args, io.Discard, io.Discard); got != tt.want {
				t.Errorf("Exit-Code %d, want %d", got, tt.want)
			}
		})
	}
}
