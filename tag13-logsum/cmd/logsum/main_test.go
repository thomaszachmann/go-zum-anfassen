package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunFastEqualsSlow(t *testing.T) {
	var out [2]bytes.Buffer
	for i, mode := range []string{"-slow=true", "-slow=false"} {
		code := run([]string{mode, "../../testdata/access.log"}, nil, &out[i], &out[i])
		if code != 0 {
			t.Fatalf("%s: exit %d: %s", mode, code, out[i].String())
		}
	}
	if out[0].String() != out[1].String() {
		t.Errorf("Ausgaben weichen ab:\n%s\n---\n%s", &out[0], &out[1])
	}
	if !strings.Contains(out[1].String(), "Top Status:") {
		t.Errorf("Top Status fehlt:\n%s", &out[1])
	}
}

func TestRunBadFormat(t *testing.T) {
	var buf bytes.Buffer
	if code := run([]string{"-format", "xml"}, nil, &buf, &buf); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
}
