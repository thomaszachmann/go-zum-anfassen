package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/semver/internal/semver"
)

var update = flag.Bool("update", false, "Golden Files neu schreiben")

func TestSortGolden(t *testing.T) {
	in, err := os.Open(filepath.Join("testdata", "sort.input"))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close() //nolint:errcheck // nur gelesen

	got, err := run([]string{"sort"}, in)
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "sort.golden")
	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("Ausgabe weicht ab:\n%s\nwill (%s):\n%s", got, golden, want)
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{"compare", []string{"compare", "1.2.3", "1.10.0"}, "1.2.3 < 1.10.0\n", nil},
		{"bump", []string{"bump", "minor", "v1.2.3"}, "v1.3.0\n", nil},
		{"leer", nil, "", errUsage},
		{"zu wenig", []string{"compare", "1.0.0"}, "", errUsage},
		{"kaputt", []string{"bump", "patch", "1.x.0"}, "", semver.ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := run(tt.args, strings.NewReader(""))
			if !errors.Is(err, tt.wantErr) { // errors.Is(nil, nil) ist true
				t.Fatalf("err = %v, will %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Ausgabe = %q, will %q", got, tt.want)
			}
		})
	}
}
