// ask: Fragen an ein lokales LLM – gestreamt, als JSON oder mit Tools.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"example.com/ask/internal/llm"
)

func main() {
	triage := flag.Bool("triage", false, "Antwort als JSON (severity, summary, action)")
	agent := flag.Bool("tools", false, "Modell darf lokale Tools aufrufen")
	flag.Parse()
	q := strings.Join(flag.Args(), " ")
	if q == "" {
		fmt.Fprintln(os.Stderr, "usage: ask [-triage|-tools] <frage>")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	c := llm.FromEnv()

	var err error
	switch {
	case *triage:
		err = runTriage(ctx, c, q)
	case *agent:
		err = runTools(ctx, c, q)
	default:
		msgs := []llm.Message{{Role: "user", Content: q}}
		err = c.Stream(ctx, msgs, func(s string) { fmt.Print(s) })
		fmt.Println() // Zeilenende nach dem Stream
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ask:", err)
		os.Exit(1)
	}
}
