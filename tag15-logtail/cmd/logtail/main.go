// logtail zeigt die letzten Zeilen einer Datei und folgt ihr wie tail -F:
// logtail [-n 10] [-i 250ms] DATEI
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/logtail/internal/tail"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("logtail", flag.ContinueOnError)
	fl.SetOutput(stderr)
	n := fl.Int("n", 10, "so viele Zeilen zuerst zeigen")
	every := fl.Duration("i", 250*time.Millisecond, "Prüfintervall")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() != 1 || *n < 0 || *every <= 0 {
		fmt.Fprintln(stderr, "Aufruf: logtail [-n 10] [-i 250ms] DATEI")
		return 2
	}
	path := fl.Arg(0)
	lines, off, err := tail.Last(path, *n)
	if err != nil {
		fmt.Fprintln(stderr, "logtail:", err)
		return 1
	}
	for _, l := range lines {
		fmt.Fprintln(stdout, l)
	}
	if err := tail.Follow(ctx, path, off, *every, stdout); err != nil {
		fmt.Fprintln(stderr, "logtail:", err)
		return 1
	}
	return 0
}
