// semver vergleicht, sortiert und erhöht Versionen nach semver.org.
// Tag 4 aus "Go zum Anfassen".
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/semver/internal/semver"
)

const usage = `Aufruf:
  semver compare A B         gibt A<B, A=B oder A>B aus
  semver bump PART VERSION   PART: major, minor, patch
  semver sort                sortiert Versionen von stdin`

var errUsage = errors.New("falscher Aufruf")

func main() {
	out, err := run(os.Args[1:], os.Stdin)
	fmt.Print(out)
	switch {
	case errors.Is(err, errUsage):
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	case err != nil:
		fmt.Fprintln(os.Stderr, "semver:", err)
		os.Exit(1)
	}
}

// run ist main ohne os.*: Argumente und stdin rein, Ausgabe und Fehler raus.
// So können Tests das ganze Tool aufrufen.
func run(args []string, stdin io.Reader) (string, error) {
	if len(args) == 0 {
		return "", errUsage
	}
	switch cmd := args[0]; {
	case cmd == "compare" && len(args) == 3:
		a, err := semver.Parse(args[1])
		if err != nil {
			return "", err
		}
		b, err := semver.Parse(args[2])
		if err != nil {
			return "", err
		}
		op := [...]string{"<", "=", ">"}[semver.Compare(a, b)+1]
		return fmt.Sprintf("%s %s %s\n", a, op, b), nil
	case cmd == "bump" && len(args) == 3:
		v, err := semver.Parse(args[2])
		if err != nil {
			return "", err
		}
		v, err = semver.Bump(v, args[1])
		if err != nil {
			return "", err
		}
		return v.String() + "\n", nil
	case cmd == "sort" && len(args) == 1:
		return sortVersions(stdin)
	}
	return "", errUsage
}

// sortVersions liest eine Version pro Zeile; Leerzeilen werden übersprungen.
func sortVersions(r io.Reader) (string, error) {
	var vs []semver.Version
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		v, err := semver.Parse(line)
		if err != nil {
			return "", err
		}
		vs = append(vs, v)
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	semver.Sort(vs)
	var b strings.Builder
	for _, v := range vs {
		b.WriteString(v.String() + "\n")
	}
	return b.String(), nil
}
