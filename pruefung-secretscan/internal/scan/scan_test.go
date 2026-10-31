package scan

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Lab-Werte, zusammengesetzt, damit dieser Test selbst keinen Treffer liefert.
var (
	awsKey = "AKIA" + "LABEXAMPLE000000"
	ghTok  = "ghp_" + strings.Repeat("a", 36)
	pemKey = "-----BEGIN " + "RSA PRIVATE KEY-----"
	skKey  = "sk-" + "lab-master-1234567890"
)

func TestLine(t *testing.T) {
	tests := []struct {
		line string
		want []string // Regeln
	}{
		{"aws_key = " + awsKey, []string{"aws-access-key"}},
		{"token: " + ghTok, []string{"github-token"}},
		{pemKey, []string{"private-key"}},
		{`OPENAI_API_KEY="` + skKey + `"`, []string{"api-key"}},
		{awsKey + " " + ghTok, []string{"aws-access-key", "github-token"}},
		{"aws_key = " + awsKey + " // secretscan:ignore", nil},
		{"AKIA zu kurz, ghp_kurz, sk-x", nil},
	}
	for _, tt := range tests {
		var got []string
		for _, f := range Line(tt.line) {
			got = append(got, f.Rule)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Line(%q) = %v, will %v", tt.line, got, tt.want)
		}
	}
}

func TestMask(t *testing.T) {
	if got := Mask(awsKey); got != "AKIA…(20)" {
		t.Errorf("Mask = %q", got)
	}
	if got := Mask("abc"); got != "…(3)" {
		t.Errorf("Mask(kurz) = %q", got)
	}
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDir(t *testing.T) {
	root := t.TempDir()
	write(t, root, "app/config.yaml", "user: lab\ntoken: "+ghTok+"\n")
	write(t, root, "deploy/.env", "A=1\nAWS_KEY="+awsKey+"\nB=2\n")
	write(t, root, "keys/id_rsa", pemKey+"\nMIIE...\n")
	write(t, root, ".git/config", "token: "+ghTok+"\n") // übersprungen
	write(t, root, "bin/tool", "\x00\x01"+awsKey)       // binär
	write(t, root, "ok.txt", "nichts\n"+awsKey+" secretscan:ignore\n")
	for _, workers := range []int{1, 8} {
		got, err := Dir(context.Background(), root, workers)
		if err != nil {
			t.Fatal(err)
		}
		want := []Finding{
			{File: "app/config.yaml", Line: 2, Rule: "github-token",
				Match: "ghp_…(40)"},
			{File: "deploy/.env", Line: 2, Rule: "aws-access-key",
				Match: "AKIA…(20)"},
			{File: "keys/id_rsa", Line: 1, Rule: "private-key",
				Match: "----…(31)"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("workers=%d:\n got %+v\nwant %+v", workers, got, want)
		}
	}
}

func TestDirFehler(t *testing.T) {
	_, err := Dir(context.Background(), "/gibt/es/nicht", 2)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Dir(ctx, t.TempDir(), 2); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, will context.Canceled", err)
	}
}
