// envdiff vergleicht zwei .env-Dateien, ohne Secrets zu zeigen:
// envdiff [-show] ALT NEU. Exit 0 = gleich, 1 = Unterschiede, 2 = Fehler.
// Tag 15 aus "Go zum Anfassen".
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/envdiff/internal/envfile"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("envdiff", flag.ContinueOnError)
	fl.SetOutput(stderr)
	show := fl.Bool("show", false, "Werte im Klartext zeigen (Vorsicht: Secrets!)")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if fl.NArg() != 2 {
		fmt.Fprintln(stderr, "Aufruf: envdiff [-show] ALT NEU")
		return 2
	}
	var envs [2]map[string]string
	for i, name := range fl.Args() {
		env, err := load(name)
		if err != nil {
			fmt.Fprintln(stderr, "envdiff:", err)
			return 2
		}
		envs[i] = env
	}
	val := envfile.Mask
	if *show {
		val = func(v string) string { return v }
	}
	changes := envfile.Diff(envs[0], envs[1])
	for _, c := range changes {
		switch c.Kind {
		case envfile.Added:
			fmt.Fprintf(stdout, "+ %s=%s\n", c.Key, val(c.New))
		case envfile.Removed:
			fmt.Fprintf(stdout, "- %s\n", c.Key)
		case envfile.Changed:
			fmt.Fprintf(stdout, "~ %s: %s -> %s\n", c.Key, val(c.Old), val(c.New))
		}
	}
	if len(changes) > 0 {
		return 1
	}
	return 0
}

func load(name string) (map[string]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	env, err := envfile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return env, nil
}
