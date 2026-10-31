// urlcheck prüft URLs parallel: urlcheck [-c 8] [-rps 0] [-json] [DATEI...]
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"example.com/urlcheck/internal/check"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer,
) int {
	fl := flag.NewFlagSet("urlcheck", flag.ContinueOnError)
	fl.SetOutput(stderr)
	workers := fl.Int("c", 8, "gleichzeitige Anfragen")
	rps := fl.Float64("rps", 0, "max. Anfragen pro Sekunde (0 = unbegrenzt)")
	timeout := fl.Duration("timeout", 10*time.Second, "Timeout pro URL")
	asJSON := fl.Bool("json", false, "JSON-Ausgabe")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	urls, err := readURLs(fl.Args(), stdin)
	if err != nil {
		fmt.Fprintln(stderr, "urlcheck:", err)
		return 2
	}

	c := &check.Checker{
		HTTP:    &http.Client{Timeout: *timeout},
		Workers: *workers,
		RPS:     *rps,
	}
	start := time.Now()
	results := c.Run(ctx, urls)

	failed := 0
	for _, r := range results {
		if !r.OK() {
			failed++
		}
	}
	if *asJSON {
		_ = json.NewEncoder(stdout).Encode(results)
	} else {
		for _, r := range results {
			mark, status := "✔", fmt.Sprint(r.Status)
			if !r.OK() {
				mark = "✘"
			}
			if r.Err != "" {
				status = "ERR"
			}
			line := fmt.Sprintf("%s %-3s %5dms %s", mark, status,
				r.Duration.Milliseconds(), r.URL)
			if r.Err != "" {
				line += "  " + r.Err
			}
			fmt.Fprintln(stdout, line)
		}
		fmt.Fprintf(stdout, "%d URLs, %d fehlerhaft, %d Worker, %v\n", len(urls),
			failed, *workers, time.Since(start).Round(time.Millisecond))
	}
	if failed > 0 {
		return 1
	}
	return 0
}

// readURLs liest eine URL pro Zeile aus Dateien oder stdin; # = Kommentar.
func readURLs(files []string, stdin io.Reader) ([]string, error) {
	readers := []io.Reader{stdin}
	if len(files) > 0 {
		readers = nil
		for _, name := range files {
			f, err := os.Open(name)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			readers = append(readers, f)
		}
	}
	var urls []string
	sc := bufio.NewScanner(io.MultiReader(readers...))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}
	return urls, sc.Err()
}
