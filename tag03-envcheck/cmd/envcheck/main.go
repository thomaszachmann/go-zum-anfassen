// envcheck prüft, ob Werkzeuge, Umgebungsvariablen und Ports vorhanden sind.
// Tag 3 aus "Go zum Anfassen".
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"example.com/envcheck/internal/check"
	"example.com/envcheck/internal/config"
)

// Exit-Codes: 0 = alles da, 1 = Checks fehlgeschlagen, 2 = Aufruf/Config kaputt.
func main() {
	path := flag.String("config", "envcheck.yaml", "Pfad zur YAML-Konfiguration")
	flag.Parse()

	err := run(*path)
	if err == nil {
		return
	}
	var ce *check.Error
	if errors.As(err, &ce) { // mindestens ein Check ist fehlgeschlagen
		if errors.Is(err, check.ErrNotSet) {
			fmt.Fprintln(os.Stderr, "Tipp: fehlende Variablen mit export setzen")
		}
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "envcheck:", err)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "Tipp: envcheck.yaml anlegen oder -config nutzen")
	}
	os.Exit(2)
}

// run lädt die Config, führt alle Checks aus und druckt eine Tabelle.
// Fehlgeschlagene Checks kommen gesammelt per errors.Join zurück.
func run(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	var errs []error
	for _, name := range cfg.Tools {
		where, err := check.Tool(name)
		errs = append(errs, line("tool", name, where, err))
	}
	for _, name := range cfg.Env {
		errs = append(errs, line("env", name, "gesetzt", check.Env(name)))
	}
	for _, addr := range cfg.Ports {
		err := check.Port(addr, cfg.Timeout)
		errs = append(errs, line("port", addr, "offen", err))
	}
	joined := errors.Join(errs...) // nil-Einträge ignoriert Join
	if joined == nil {
		fmt.Printf("alle %d Checks bestanden\n", len(errs))
		return nil
	}
	failed := 0
	for _, err := range errs {
		if err != nil {
			failed++
		}
	}
	fmt.Printf("%d von %d Checks fehlgeschlagen\n", failed, len(errs))
	return joined
}

// line schreibt eine Ergebniszeile und reicht err unverändert durch.
func line(kind, name, ok string, err error) error {
	mark, detail := "✔", ok
	// Go 1.26: errors.AsType ist die generische Variante von errors.As.
	if ce, isCheck := errors.AsType[*check.Error](err); isCheck {
		mark, detail = "✘", ce.Err.Error()
	}
	fmt.Printf("%s %-5s %-16s %s\n", mark, kind, name, detail)
	return err
}
