// ghrel zeigt GitHub-Releases eines Repos: ghrel [-n 5] [-latest] owner/repo
package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"text/tabwriter"
	"time"

	"example.com/ghrel/internal/github"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, getenv func(string) string,
	stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("ghrel", flag.ContinueOnError)
	fl.SetOutput(stderr)
	n := fl.Int("n", 5, "Anzahl Releases")
	latest := fl.Bool("latest", false, "nur das neueste stabile Release")
	asJSON := fl.Bool("json", false, "JSON statt Tabelle")
	timeout := fl.Duration("timeout", 15*time.Second, "Gesamtbudget inkl. Retries")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	repo := fl.Arg(0)
	if fl.NArg() != 1 || strings.Count(repo, "/") != 1 {
		fmt.Fprintln(stderr, "Aufruf: ghrel [-n 5] [-latest] [-json] owner/repo")
		return 2
	}

	c := &github.Client{
		BaseURL: cmp.Or(getenv("GHREL_API"), "https://api.github.com"),
		Token:   getenv("GITHUB_TOKEN"),
		HTTP:    &http.Client{Timeout: 5 * time.Second}, // pro Versuch
		Retries: 3,
		Backoff: 500 * time.Millisecond,
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	var rs []github.Release
	var err error
	if *latest {
		var r github.Release
		r, err = c.Latest(ctx, repo)
		rs = []github.Release{r}
	} else {
		rs, err = c.Releases(ctx, repo, *n)
	}
	var ae *github.APIError
	switch {
	case errors.As(err, &ae) && ae.Status == http.StatusNotFound:
		fmt.Fprintf(stderr, "ghrel: %s nicht gefunden (oder keine Releases)\n", repo)
		return 1
	case err != nil:
		fmt.Fprintln(stderr, "ghrel:", err)
		return 1
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rs)
		return 0
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "TAG\tDATUM\tNAME")
	for _, r := range rs {
		name := r.Name
		if r.Prerelease {
			name += " (pre)"
		}
		date := r.PublishedAt.Format(time.DateOnly)
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.TagName, date, name)
	}
	_ = tw.Flush()
	return 0
}
