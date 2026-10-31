// Package llm spricht die OpenAI-kompatible API (Ollama, vLLM, Mistral) an.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON als String
	} `json:"function"`
}

type Request struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Stream         bool      `json:"stream,omitempty"`
	Tools          []any     `json:"tools,omitempty"`
	ResponseFormat any       `json:"response_format,omitempty"`
}

type Client struct {
	BaseURL, Key, Model string
	HTTP                *http.Client
}

// FromEnv liest LLM_BASE_URL, LLM_API_KEY und LLM_MODEL (Defaults: Ollama).
func FromEnv() *Client {
	return &Client{
		BaseURL: env("LLM_BASE_URL", "http://127.0.0.1:11434/v1"),
		Key:     os.Getenv("LLM_API_KEY"),
		Model:   env("LLM_MODEL", "ministral-3:3b"),
		HTTP:    &http.Client{Timeout: 5 * time.Minute},
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// post schickt JSON und liefert die Antwort nur bei Status 200.
func (c *Client) post(ctx context.Context, path string, body any) (
	*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%s: %s: %s", path, resp.Status,
			strings.TrimSpace(string(msg)))
	}
	return resp, nil
}
