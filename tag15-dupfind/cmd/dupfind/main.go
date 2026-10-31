// dupfind findet doppelte Dateien: dupfind [-w 8] [-json] ORDNER
// Tag 15 aus "Go zum Anfassen".
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

	"example.com/dupfind/internal/dup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("dupfind", flag.ContinueOnError)
	fl.SetOutput(stderr)
	workers := fl.Int("w", runtime.NumCPU(), "parallele Hasher")
	asJSON := fl.Bool("json", false, "JSON-Ausgabe")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() != 1 {
		fmt.Fprintln(stderr, "Aufruf: dupfind [-w N] [-json] ORDNER")
		return 2
	}
	groups, err := dup.Find(ctx, fl.Arg(0), *workers)
	if err != nil {
		fmt.Fprintln(stderr, "dupfind:", err)
		return 1
	}
	if *asJSON {
		_ = json.NewEncoder(stdout).Encode(groups)
		return 0
	}
	var wasted int64
	for _, g := range groups {
		fmt.Fprintf(stdout, "%d × %d B  %s\n", len(g.Paths), g.Size, g.Hash[:12])
		for _, p := range g.Paths {
			fmt.Fprintln(stdout, "  "+p)
		}
		wasted += g.Wasted()
	}
	fmt.Fprintf(stdout, "%d Gruppen, %d B verschwendet\n", len(groups), wasted)
	return 0
}
