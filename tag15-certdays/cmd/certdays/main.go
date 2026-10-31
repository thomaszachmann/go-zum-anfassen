// certdays zeigt, wie lange Zertifikate noch gültig sind:
// certdays [-warn 30] DATEI.pem|host:port...
// Exit 1, wenn eins unter -warn Tagen liegt oder nicht lesbar ist.
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"text/tabwriter"
	"time"

	"example.com/certdays/internal/certs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], time.Now(), os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, now time.Time,
	stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("certdays", flag.ContinueOnError)
	fl.SetOutput(stderr)
	warn := fl.Int("warn", 30, "Exit 1 unter so vielen Resttagen")
	timeout := fl.Duration("timeout", 5*time.Second, "Timeout pro Server")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() == 0 {
		fmt.Fprintln(stderr, "Aufruf: certdays [-warn 30] DATEI.pem|host:port...")
		return 2
	}
	code := 0
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "TAGE\tABLAUF\tSUBJECT\tQUELLE")
	for _, target := range fl.Args() {
		infos, err := load(ctx, target, now, *timeout)
		if err != nil {
			fmt.Fprintln(stderr, "certdays:", err)
			code = 1
			continue
		}
		for _, i := range infos {
			mark := ""
			if i.Days < *warn {
				mark, code = " ✘", 1
			}
			fmt.Fprintf(tw, "%d%s\t%s\t%s\t%s\n", i.Days, mark,
				i.NotAfter.Format(time.DateOnly), i.Subject, i.Source)
		}
	}
	_ = tw.Flush()
	return code
}

// load nimmt eine vorhandene Datei als PEM, alles andere als host:port.
func load(ctx context.Context, target string, now time.Time,
	timeout time.Duration) ([]certs.Info, error) {
	if data, err := os.ReadFile(target); err == nil {
		return certs.FromPEM(target, data, now)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return certs.FromAddr(ctx, target, now)
}
