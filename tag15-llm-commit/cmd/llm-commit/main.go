// llm-commit schreibt eine Commit-Message aus dem gestagten Diff per LLM:
// git add -p && llm-commit | git commit -F -
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"example.com/llm-commit/internal/commit"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, getenv func(string) string,
	stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("llm-commit", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fromStdin := fl.Bool("stdin", false, "Diff von stdin statt git diff --cached")
	limit := fl.Int("max", 12000, "höchstens so viele Bytes Diff ans Modell")
	timeout := fl.Duration("timeout", 2*time.Minute, "Gesamtbudget")
	if err := fl.Parse(args); err != nil || fl.NArg() != 0 {
		fmt.Fprintln(stderr, "Aufruf: llm-commit [-stdin] [-max 12000]")
		return 2
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	var diff []byte
	var err error
	if *fromStdin {
		diff, err = io.ReadAll(stdin)
	} else {
		diff, err = exec.CommandContext(ctx, "git", "diff", "--cached",
			"--no-color").Output()
	}
	if err != nil {
		fmt.Fprintln(stderr, "llm-commit: diff lesen:", err)
		return 1
	}
	if strings.TrimSpace(string(diff)) == "" {
		fmt.Fprintln(stderr, "llm-commit: nichts gestaged (git add vergessen?)")
		return 1
	}
	d, cut := commit.Truncate(string(diff), *limit)
	if cut {
		fmt.Fprintf(stderr, "llm-commit: Diff auf %d Bytes gekürzt\n", len(d))
	}

	c := &commit.Client{
		BaseURL: cmp.Or(getenv("LLM_BASE_URL"), "http://127.0.0.1:11434/v1"),
		Model:   cmp.Or(getenv("LLM_MODEL"), "ministral-3:3b"),
		Key:     getenv("LLM_API_KEY"),
		HTTP:    &http.Client{},
	}
	msg, err := c.Message(ctx, d)
	if err != nil {
		fmt.Fprintln(stderr, "llm-commit:", err)
		return 1
	}
	fmt.Fprintln(stdout, msg)
	return 0
}
