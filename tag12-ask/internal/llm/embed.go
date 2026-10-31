package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// Embed liefert pro Eingabetext einen Vektor (Reihenfolge wie input).
func (c *Client) Embed(ctx context.Context, model string, input []string) (
	[][]float64, error) {
	resp, err := c.post(ctx, "/embeddings",
		map[string]any{"model": model, "input": input})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("embeddings lesen: %w", err)
	}
	if len(out.Data) != len(input) {
		return nil, fmt.Errorf("%d vektoren für %d texte", len(out.Data), len(input))
	}
	vecs := make([][]float64, len(input))
	for _, d := range out.Data {
		vecs[d.Index] = d.Embedding
	}
	return vecs, nil
}
