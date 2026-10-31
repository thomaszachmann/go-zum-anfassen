// secretscan sucht Secrets in einem Ordner: secretscan [-w 8] [-json] ORDNER
// Exit 0 = nichts gefunden, 1 = Treffer, 2 = Aufruf- oder Lesefehler.
// Abschlussprüfung aus "Go zum Anfassen".
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"

	"example.com/secretscan/internal/scan"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("secretscan", flag.ContinueOnError)
	fl.SetOutput(stderr)
	workers := fl.Int("w", runtime.NumCPU(), "parallele Leser")
	asJSON := fl.Bool("json", false, "JSON-Ausgabe")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() != 1 {
		fmt.Fprintln(stderr, "Aufruf: secretscan [-w N] [-json] ORDNER")
		return 2
	}
	found, err := scan.Dir(ctx, fl.Arg(0), *workers)
	if err != nil {
		fmt.Fprintln(stderr, "secretscan:", err)
		return 2
	}
	if *asJSON {
		if found == nil {
			found = []scan.Finding{} // [] statt null
		}
		enc := json.NewEncoder(stdout)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(found)
	} else {
		for _, f := range found {
			fmt.Fprintf(stdout, "%s:%d: %s %s\n", f.File, f.Line, f.Rule, f.Match)
		}
		fmt.Fprintf(stdout, "%d Treffer\n", len(found))
	}
	if len(found) > 0 {
		return 1
	}
	return 0
}
