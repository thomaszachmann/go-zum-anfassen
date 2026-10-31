package guard

import (
	"io"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name, input string
		want        bool
	}{
		{"exakter Name", "prod-eu\n", true},
		{"ohne Zeilenende", "prod-eu", true},
		{"nur y", "y\n", false},
		{"leer", "", false},
		{"anderer Kontext", "prod-us\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Confirm(strings.NewReader(tt.input), io.Discard, "prod-eu")
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Confirm(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
