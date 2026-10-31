package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"example.com/ask/internal/llm"
	"example.com/ask/internal/tools"
)

var toolDefs = []any{json.RawMessage(`{"type": "function", "function": {
  "name": "disk_free",
  "description": "Freier Speicherplatz eines Pfads auf diesem Rechner",
  "parameters": {"type": "object",
    "properties": {"path": {"type": "string"}}, "required": ["path"]}}}`)}

// callTool führt nur bekannte Tools aus; Argumente kommen vom Modell.
func callTool(name, args string) string {
	if name != "disk_free" {
		return `{"error":"unbekanntes tool"}`
	}
	var in struct{ Path string }
	if err := json.Unmarshal([]byte(args), &in); err != nil || in.Path == "" {
		return `{"error":"argument path fehlt"}`
	}
	d, err := tools.DiskFree(in.Path)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	b, _ := json.Marshal(d)
	return string(b)
}

func runTools(ctx context.Context, c *llm.Client, q string) error {
	msgs := []llm.Message{
		{Role: "system", Content: "Nutze Tools für Fakten über diesen Rechner."},
		{Role: "user", Content: q},
	}
	for range 5 { // Obergrenze gegen Endlosschleifen
		m, err := c.Chat(ctx, llm.Request{Messages: msgs, Tools: toolDefs})
		if err != nil {
			return err
		}
		if len(m.ToolCalls) == 0 {
			fmt.Println(m.Content)
			return nil
		}
		msgs = append(msgs, m)
		for _, tc := range m.ToolCalls {
			out := callTool(tc.Function.Name, tc.Function.Arguments)
			fmt.Fprintf(os.Stderr, "-> %s(%s) = %s\n",
				tc.Function.Name, tc.Function.Arguments, out)
			msgs = append(msgs, llm.Message{Role: "tool",
				ToolCallID: tc.ID, Content: out})
		}
	}
	return errors.New("zu viele tool-runden")
}
