// portwait wartet, bis TCP-Ports offen sind: portwait [-timeout 30s] host:port...
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"example.com/portwait/internal/wait"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("portwait", flag.ContinueOnError)
	fl.SetOutput(stderr)
	timeout := fl.Duration("timeout", 30*time.Second, "Gesamtbudget für alle Ports")
	maxWait := fl.Duration("max", 2*time.Second, "längste Pause zwischen Versuchen")
	verbose := fl.Bool("v", false, "jeden Fehlversuch melden")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() == 0 {
		fmt.Fprintln(stderr, "Aufruf: portwait [-timeout 30s] [-v] host:port...")
		return 2
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	o := wait.Options{Max: *maxWait}
	if *verbose {
		o.OnRetry = func(n int, err error, d time.Duration) {
			fmt.Fprintf(stderr, "  Versuch %d: %v (warte %v)\n", n, err, d)
		}
	}
	start := time.Now()
	for _, addr := range fl.Args() {
		n, err := wait.Port(ctx, addr, o)
		if err != nil {
			fmt.Fprintln(stderr, "portwait:", err)
			return 1
		}
		fmt.Fprintf(stdout, "✔ %s offen (%d Versuche, %v)\n", addr, n,
			time.Since(start).Round(time.Millisecond))
	}
	return 0
}
