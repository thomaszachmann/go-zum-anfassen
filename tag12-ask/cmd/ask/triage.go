package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"example.com/ask/internal/llm"
)

type Triage struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Action   string `json:"action"`
}

// triageSchema zwingt das Modell in genau diese JSON-Form.
var triageSchema = json.RawMessage(`{"type": "json_schema", "json_schema": {
  "name": "triage", "schema": {"type": "object",
    "properties": {
      "severity": {"type": "string", "enum": ["low", "medium", "high"]},
      "summary": {"type": "string"}, "action": {"type": "string"}},
    "required": ["severity", "summary", "action"]}}}`)

func runTriage(ctx context.Context, c *llm.Client, q string) error {
	m, err := c.Chat(ctx, llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "Du bist SRE. Bewerte die Meldung knapp."},
			{Role: "user", Content: q},
		},
		ResponseFormat: triageSchema,
	})
	if err != nil {
		return err
	}
	t, err := parseTriage(m.Content)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(t)
}

// parseTriage traut dem Modell nicht: Schema ist Wunsch, Prüfung ist Pflicht.
func parseTriage(s string) (Triage, error) {
	var t Triage
	if err := json.Unmarshal([]byte(s), &t); err != nil {
		return t, fmt.Errorf("kein gültiges JSON: %w", err)
	}
	if !slices.Contains([]string{"low", "medium", "high"}, t.Severity) {
		return t, fmt.Errorf("ungültige severity %q", t.Severity)
	}
	return t, nil
}
