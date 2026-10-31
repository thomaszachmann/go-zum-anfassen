// docsearch: semantische Suche über Markdown-Runbooks mit Embeddings.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/ask/internal/llm"
	"example.com/ask/internal/search"
)

func main() {
	dir := flag.String("dir", "docs", "Verzeichnis mit *.md")
	model := flag.String("model", "all-minilm", "Embedding-Modell")
	k := flag.Int("k", 2, "Anzahl Treffer")
	flag.Parse()
	q := strings.Join(flag.Args(), " ")
	if q == "" {
		fmt.Fprintln(os.Stderr, "usage: docsearch [-dir docs] [-k 2] <frage>")
		os.Exit(2)
	}
	ctx, c := context.Background(), llm.FromEnv()
	hits, err := find(ctx, c, *dir, *model, q, *k)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docsearch:", err)
		os.Exit(1)
	}
	for _, h := range hits {
		title, _, _ := strings.Cut(h.Text, "\n")
		fmt.Printf("%.3f  %-16s %s\n", h.Score, h.Source, title)
	}
}

// find bettet Frage und Dokumente ein und liefert die k besten Treffer.
func find(ctx context.Context, c *llm.Client, dir, model, q string, k int) (
	[]search.Hit, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("keine *.md in %s", dir)
	}
	var docs []search.Doc
	texts := []string{q} // Frage zuerst: ein einziger Embedding-Aufruf
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		docs = append(docs, search.Doc{Source: filepath.Base(f), Text: string(b)})
		texts = append(texts, string(b))
	}
	vecs, err := c.Embed(ctx, model, texts)
	if err != nil {
		return nil, err
	}
	for i := range docs {
		docs[i].Vec = vecs[i+1]
	}
	return search.TopK(docs, vecs[0], k), nil
}
