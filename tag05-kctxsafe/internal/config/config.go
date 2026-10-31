// Package config lädt Einstellungen: Default < Datei < Env (< Flag in main).
package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Config hält die Einstellungen von kctxsafe.
type Config struct {
	ProdPattern string // regulärer Ausdruck für Produktionskontexte
}

// Load liest optional eine key=value-Datei und danach die Umgebung.
func Load(path string, getenv func(string) string) (Config, error) {
	c := Config{ProdPattern: "prod"}
	if path != "" {
		if err := c.readFile(path); err != nil {
			return c, err
		}
	}
	if v := getenv("KCTXSAFE_PROD_PATTERN"); v != "" {
		c.ProdPattern = v
	}
	return c, nil
}

func (c *Config) readFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: erwartet key=value", path, n)
		}
		switch strings.TrimSpace(key) {
		case "prod_pattern":
			c.ProdPattern = strings.TrimSpace(val)
		default:
			return fmt.Errorf("%s:%d: unbekannter Schlüssel %q", path, n, key)
		}
	}
	return sc.Err()
}

// ProdRegexp kompiliert das Muster; Fehler meldet die Config, nicht der Wechsel.
func (c Config) ProdRegexp() (*regexp.Regexp, error) {
	re, err := regexp.Compile(c.ProdPattern)
	if err != nil {
		return nil, fmt.Errorf("prod_pattern: %w", err)
	}
	return re, nil
}
