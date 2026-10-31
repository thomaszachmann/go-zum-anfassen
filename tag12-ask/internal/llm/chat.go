package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Chat schickt eine Anfrage ohne Streaming und liefert die Antwort.
func (c *Client) Chat(ctx context.Context, r Request) (Message, error) {
	r.Model, r.Stream = c.Model, false
	resp, err := c.post(ctx, "/chat/completions", r)
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()
	var out struct {
		Choices []struct{ Message Message } `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Message{}, fmt.Errorf("antwort lesen: %w", err)
	}
	if len(out.Choices) == 0 {
		return Message{}, errors.New("antwort ohne choices")
	}
	return out.Choices[0].Message, nil
}

// Stream ruft onDelta für jedes Textstück aus dem SSE-Stream auf.
func (c *Client) Stream(ctx context.Context, msgs []Message,
	onDelta func(string)) error {
	resp, err := c.post(ctx, "/chat/completions",
		Request{Model: c.Model, Messages: msgs, Stream: true})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data: ")
		if !ok {
			continue // Leerzeilen und Kommentare
		}
		if data == "[DONE]" {
			return nil
		}
		var chunk struct {
			Choices []struct {
				Delta struct{ Content string } `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("sse-chunk: %w", err)
		}
		if len(chunk.Choices) > 0 {
			onDelta(chunk.Choices[0].Delta.Content)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return errors.New("stream ohne [DONE] beendet")
}
