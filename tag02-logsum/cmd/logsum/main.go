// logsum fasst ein Zugriffslog zusammen: Status, Fehlerquote, Top-Pfade.
// Tag 2 aus "Go zum Anfassen".
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/logsum/internal/logs"
)

func main() {
	format := flag.String("format", "combined", "Logformat: combined oder json")
	top := flag.Int("top", 3, "Anzahl der Top-Pfade")
	flag.Parse()

	// Das Interface macht die Formate austauschbar.
	parsers := map[string]logs.Parser{
		"combined": logs.Combined{},
		"json":     logs.JSON{},
	}
	p, ok := parsers[*format]
	if !ok {
		fmt.Fprintf(os.Stderr, "logsum: unbekanntes Format %q\n", *format)
		os.Exit(2)
	}

	var in io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "logsum:", err)
			os.Exit(1)
		}
		defer f.Close() //nolint:errcheck // nur gelesen
		in = f
	}

	s, err := logs.Read(in, p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logsum:", err)
		os.Exit(1)
	}
	if err := s.Report(os.Stdout, *top); err != nil {
		fmt.Fprintln(os.Stderr, "logsum:", err)
		os.Exit(1)
	}
}
