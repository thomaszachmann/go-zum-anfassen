package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestClean(t *testing.T) {
	tests := []struct{ in, want string }{
		{"feat: neu", "feat: neu"},
		{"```\nfix(api): Timeout setzen\n```", "fix(api): Timeout setzen"},
		{"\"docs: README\"", "docs: README"},
		{"feat: x\nBody direkt", "feat: x\n\nBody direkt"},
		{"  chore: y  \n\n\n\nmehr  ", "chore: y\n\nmehr"},
	}
	for _, tt := range tests {
		if got := Clean(tt.in); got != tt.want {
			t.Errorf("Clean(%q) = %q, will %q", tt.in, got, tt.want)
		}
	}
}

func TestValid(t *testing.T) {
	ok := []string{"feat: neu", "fix(api)!: Bruch", "ci(go): Lint in CI"}
	bad := []string{"Added stuff", "feat:ohne Leerzeichen", "feat: " +
		strings.Repeat("x", 70), "", "feature: falscher Typ"}
	for _, s := range ok {
		if err := Valid(s); err != nil {
			t.Errorf("Valid(%q) = %v", s, err)
		}
	}
	for _, s := range bad {
		if Valid(s) == nil {
			t.Errorf("Valid(%q) = nil, will Fehler", s)
		}
	}
}

func TestTruncate(t *testing.T) {
	diff := strings.Repeat("+zeile\n", 100) // 700 Bytes
	got, cut := Truncate(diff, 50)
	if !cut || len(got) > 50 || !strings.HasSuffix(got, "+zeile\n") {
		t.Errorf("Truncate = %q, %v", got, cut)
	}
	if got, cut := Truncate("kurz\n", 50); cut || got != "kurz\n" {
		t.Errorf("kurzer Diff verändert: %q, %v", got, cut)
	}
}

// fakeLLM antwortet nacheinander mit answers und prüft jede Anfrage.
func fakeLLM(t *testing.T, answers ...string) (*Client, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			n := calls.Add(1)
			if r.URL.Path != "/chat/completions" ||
				r.Header.Get("Authorization") != "Bearer sk-lab-1234" {
				t.Errorf("Anfrage: %s %q", r.URL.Path, r.Header.Get("Authorization"))
			}
			var req request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Error(err)
			}
			last := req.Messages[len(req.Messages)-1].Content
			if req.Model != "lab-model" || (n == 1 && !strings.Contains(last, "+x")) {
				t.Errorf("Request %d: model=%q, zuletzt %q", n, req.Model, last)
			}
			a, _ := json.Marshal(answers[min(int(n), len(answers))-1])
			fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant",`+
				`"content":%s}}]}`, a)
		}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL, Model: "lab-model", Key: "sk-lab-1234",
		HTTP: srv.Client()}, &calls
}

func TestMessage(t *testing.T) {
	c, calls := fakeLLM(t, "```\nfeat(api): /healthz ergänzen\n```")
	msg, err := c.Message(context.Background(), "+x\n")
	if err != nil || msg != "feat(api): /healthz ergänzen" || calls.Load() != 1 {
		t.Errorf("Message = %q, %v (%d Aufrufe)", msg, err, calls.Load())
	}
}

func TestMessageReparatur(t *testing.T) {
	c, calls := fakeLLM(t, "Added a health endpoint", "feat: Health-Endpunkt")
	msg, err := c.Message(context.Background(), "+x\n")
	if err != nil || msg != "feat: Health-Endpunkt" || calls.Load() != 2 {
		t.Errorf("Message = %q, %v (%d Aufrufe)", msg, err, calls.Load())
	}
}

func TestMessageGibtAuf(t *testing.T) {
	c, calls := fakeLLM(t, "Quatsch")
	if _, err := c.Message(context.Background(), "+x\n"); err == nil ||
		calls.Load() != attempts {
		t.Errorf("err = %v nach %d Aufrufen, will Fehler nach %d", err,
			calls.Load(), attempts)
	}
}
