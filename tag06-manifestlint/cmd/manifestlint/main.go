// manifestlint prüft Kubernetes-YAMLs. Exit 0 = sauber, 1 = Befunde, 2 = Fehler.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/manifestlint/internal/lint"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("manifestlint", flag.ContinueOnError)
	fl.SetOutput(stderr)
	format := fl.String("o", "text", "Ausgabe: text oder json")
	var labels []string
	fl.Func("label", "Pflicht-Label (mehrfach; Default app.kubernetes.io/name)",
		func(s string) error { labels = append(labels, s); return nil })
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if len(labels) == 0 {
		labels = []string{"app.kubernetes.io/name"}
	}
	if fl.NArg() == 0 {
		fmt.Fprintln(stderr, "Aufruf: manifestlint [-o json] [-label k] PFAD...")
		return 2
	}

	files, err := lint.Files(fl.Args())
	if err != nil {
		fmt.Fprintln(stderr, "manifestlint:", err)
		return 2
	}
	findings := []lint.Finding{} // leer statt nil: JSON wird [] statt null
	for _, f := range files {
		fs, err := lint.LintFile(f, labels)
		if err != nil {
			fmt.Fprintln(stderr, "manifestlint:", err)
			return 2
		}
		findings = append(findings, fs...)
	}

	switch *format {
	case "json":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(findings); err != nil {
			fmt.Fprintln(stderr, "manifestlint:", err)
			return 2
		}
	case "text":
		for _, f := range findings {
			fmt.Fprintf(stdout, "%s %s/%s %s: %s\n",
				f.File, f.Kind, f.Name, f.Rule, f.Message)
		}
		fmt.Fprintf(stdout, "%d Datei(en), %d Befund(e)\n", len(files), len(findings))
	default:
		fmt.Fprintf(stderr, "manifestlint: unbekanntes Format %q\n", *format)
		return 2
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}
