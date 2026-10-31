package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	var out bytes.Buffer
	code := run(context.Background(), []string{"-version"}, nil, &out, &out)
	if code != 0 || !strings.HasPrefix(out.String(), "urlcheck dev (") {
		t.Fatalf("exit %d, Ausgabe %q", code, out.String())
	}
}
