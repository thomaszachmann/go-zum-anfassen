package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	clean, dirty := t.TempDir(), t.TempDir()
	key := "AKIA" + "LABEXAMPLE000000" // Lab-Wert
	err := os.WriteFile(filepath.Join(dirty, "x.env"), []byte("K="+key+"\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
		code int
		out  string
	}{
		{"sauber", []string{clean}, 0, "0 Treffer\n"},
		{"Treffer", []string{dirty}, 1,
			"x.env:1: aws-access-key AKIA…(20)\n1 Treffer\n"},
		{"ohne Ordner", nil, 2, ""},
		{"fehlt", []string{"/gibt/es/nicht"}, 2, ""},
		{"falsches Flag", []string{"-x", clean}, 2, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(context.Background(), tt.args, &out, &errOut)
			if code != tt.code || out.String() != tt.out {
				t.Errorf("Exit %d, Ausgabe %q; will %d, %q (stderr %q)",
					code, out.String(), tt.code, tt.out, errOut.String())
			}
		})
	}
}

func TestRunJSON(t *testing.T) {
	dir := t.TempDir()
	pem := "-----BEGIN " + "EC PRIVATE KEY-----\n"
	err := os.WriteFile(filepath.Join(dir, "k.pem"), []byte(pem), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := run(context.Background(), []string{"-json", dir}, &out, &out)
	if code != 1 {
		t.Fatalf("Exit %d", code)
	}
	var got []map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, out.String())
	}
	if len(got) != 1 || got[0]["rule"] != "private-key" || got[0]["line"] != 1.0 ||
		strings.Contains(out.String(), "PRIVATE KEY") {
		t.Errorf("JSON = %s", out.String())
	}
}
