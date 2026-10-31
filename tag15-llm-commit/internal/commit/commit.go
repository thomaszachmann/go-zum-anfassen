// Package commit lässt ein LLM aus einem git diff eine Commit-Message schreiben.
package commit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const system = `Du schreibst Commit-Messages im Format Conventional Commits.
Erste Zeile: typ(scope): beschreibung – höchstens 72 Zeichen, Imperativ.
Typen: feat, fix, docs, refactor, test, chore, build, ci, perf, style.
Danach optional eine Leerzeile und 1-3 Zeilen Begründung.
Antworte nur mit der Commit-Message, ohne Erklärungen und ohne Markdown.`

var subjectRe = regexp.MustCompile(
	`^(feat|fix|docs|refactor|test|chore|build|ci|perf|style)(\([\w./-]+\))?!?: \S`)

// Valid prüft die erste Zeile gegen Conventional Commits.
func Valid(msg string) error {
	subject, _, _ := strings.Cut(msg, "\n")
	if !subjectRe.MatchString(subject) {
		return fmt.Errorf("betreff %q ist kein Conventional Commit", subject)
	}
	if n := len([]rune(subject)); n > 72 {
		return fmt.Errorf("betreff hat %d Zeichen, erlaubt sind 72", n)
	}
	return nil
}

// Clean entfernt Markdown-Zäune, Anführungszeichen und überflüssige Leerzeilen
// und setzt genau eine Leerzeile zwischen Betreff und Body.
func Clean(s string) string {
	var lines []string
	for l := range strings.Lines(strings.Trim(strings.TrimSpace(s), `"`)) {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "```") {
			continue
		}
		if l == "" && (len(lines) == 0 || lines[len(lines)-1] == "") {
			continue // keine doppelten Leerzeilen
		}
		lines = append(lines, l)
	}
	if len(lines) > 1 && lines[1] != "" {
		lines = append(lines[:1], append([]string{""}, lines[1:]...)...)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// Truncate kürzt diff an einer Zeilengrenze auf höchstens limit Bytes.
func Truncate(diff string, limit int) (string, bool) {
	if len(diff) <= limit {
		return diff, false
	}
	cut := diff[:limit]
	if i := strings.LastIndexByte(cut, '\n'); i >= 0 {
		cut = cut[:i+1]
	}
	return cut, true
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

// Client spricht eine OpenAI-kompatible Chat-API an (Ollama, vLLM, Mistral).
type Client struct {
	BaseURL, Model, Key string
	HTTP                *http.Client
}

// attempts: so oft darf das Modell eine ungültige Antwort reparieren.
const attempts = 3

// Message erzeugt eine Commit-Message. Ist eine Antwort ungültig, bekommt
// das Modell den Fehler als Rückmeldung und einen neuen Versuch.
func (c *Client) Message(ctx context.Context, diff string) (string, error) {
	msgs := []message{
		{Role: "system", Content: system},
		{Role: "user", Content: "Schreibe die Commit-Message für:\n\n" + diff},
	}
	var err error
	for range attempts {
		var answer string
		answer, err = c.chat(ctx, msgs)
		if err != nil {
			return "", err
		}
		msg := Clean(answer)
		if err = Valid(msg); err == nil {
			return msg, nil
		}
		msgs = append(msgs, message{Role: "assistant", Content: answer},
			message{Role: "user", Content: "Ungültig: " + err.Error() +
				". Antworte nur mit der korrigierten Commit-Message."})
	}
	return "", err
}

func (c *Client) chat(ctx context.Context, msgs []message) (string, error) {
	body, err := json.Marshal(request{Model: c.Model, Messages: msgs,
		Temperature: 0.2})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("LLM: %s: %s", resp.Status, bytes.TrimSpace(msg))
	}
	var out struct {
		Choices []struct {
			Message message `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("LLM: Antwort ohne choices")
	}
	return out.Choices[0].Message.Content, nil
}
