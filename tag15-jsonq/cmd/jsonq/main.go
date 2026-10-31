// jsonq ist ein Mini-jq: jsonq [-r] AUSDRUCK [DATEI]
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/jsonq/internal/query"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("jsonq", flag.ContinueOnError)
	fl.SetOutput(stderr)
	raw := fl.Bool("r", false, "Strings ohne Anführungszeichen ausgeben")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() < 1 || fl.NArg() > 2 {
		fmt.Fprintln(stderr, "Aufruf: jsonq [-r] AUSDRUCK [DATEI]")
		return 2
	}
	in := stdin
	if fl.NArg() == 2 {
		f, err := os.Open(fl.Arg(1))
		if err != nil {
			fmt.Fprintln(stderr, "jsonq:", err)
			return 1
		}
		defer f.Close()
		in = f
	}

	dec := json.NewDecoder(in)
	dec.UseNumber() // 1000000 bleibt 1000000 statt 1e+06
	enc := json.NewEncoder(stdout)
	enc.SetEscapeHTML(false)
	for { // mehrere JSON-Werte hintereinander (JSON Lines) sind erlaubt
		var v any
		err := dec.Decode(&v)
		if errors.Is(err, io.EOF) {
			return 0
		}
		if err != nil {
			fmt.Fprintln(stderr, "jsonq: Eingabe:", err)
			return 1
		}
		out, err := query.Eval(v, fl.Arg(0))
		if errors.Is(err, query.ErrSyntax) {
			fmt.Fprintln(stderr, "jsonq:", err)
			return 2
		}
		if err != nil {
			fmt.Fprintln(stderr, "jsonq:", err)
			return 1
		}
		for _, x := range out {
			if s, ok := x.(string); ok && *raw {
				fmt.Fprintln(stdout, s)
				continue
			}
			_ = enc.Encode(x)
		}
	}
}
