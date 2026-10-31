// logsum fasst ein Zugriffslog zusammen: Status, Fehlerquote, Top-Pfade.
// Tag 13: schneller Pfad für das Combined-Format, Profiling per Flag.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"

	"example.com/logsum/internal/logs"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("logsum", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "combined", "Logformat: combined oder json")
	top := fs.Int("top", 3, "Anzahl der Top-Einträge")
	slow := fs.Bool("slow", false, "Parser aus Tag 2 statt schnellem Pfad")
	cpuprof := fs.String("cpuprofile", "", "CPU-Profil in diese Datei schreiben")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	parsers := map[string]logs.Parser{"combined": logs.Combined{}, "json": logs.JSON{}}
	p, ok := parsers[*format]
	if !ok {
		fmt.Fprintf(stderr, "logsum: unbekanntes Format %q\n", *format)
		return 2
	}
	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			fmt.Fprintln(stderr, "logsum:", err)
			return 1
		}
		defer f.Close() //nolint:errcheck // Profil ist best effort
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(stderr, "logsum:", err)
			return 1
		}
		defer pprof.StopCPUProfile() // läuft vor f.Close (LIFO)
	}
	data, err := load(fs.Arg(0), stdin)
	if err != nil {
		fmt.Fprintln(stderr, "logsum:", err)
		return 1
	}
	var s *logs.Summary
	if *slow || *format != "combined" {
		s, err = logs.Read(bytes.NewReader(data), p)
	} else {
		s = logs.ReadFast(data)
	}
	if err == nil {
		err = s.Report(stdout, *top)
	}
	if err != nil {
		fmt.Fprintln(stderr, "logsum:", err)
		return 1
	}
	fmt.Fprint(stdout, "Top Status:")
	for _, c := range logs.TopN(s.ByStatus, *top) { // TopN mit int-Schlüsseln
		fmt.Fprintf(stdout, " %d=%d", c.Key, c.N)
	}
	fmt.Fprintln(stdout)
	return 0
}

// load liest die Datei oder stdin komplett: schnell, kostet aber RAM.
func load(path string, stdin io.Reader) ([]byte, error) {
	if path == "" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}
