// Package config lädt die envcheck-Konfiguration aus einer YAML-Datei.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

// ErrEmpty meldet eine Konfiguration ohne einen einzigen Check.
var ErrEmpty = errors.New("keine Checks konfiguriert")

// Config beschreibt, was auf dem Rechner vorhanden sein muss.
type Config struct {
	Tools   []string      `yaml:"tools"`
	Env     []string      `yaml:"env"`
	Ports   []string      `yaml:"ports"`
	Timeout time.Duration `yaml:"timeout"`
}

// Load liest path. Jeder Fehler wird mit Kontext umhüllt (%w).
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("config laden: %w", err)
	}
	var c Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // Tippfehler wie "tool:" sind Fehler
	if err := dec.Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	if len(c.Tools)+len(c.Env)+len(c.Ports) == 0 {
		return Config{}, fmt.Errorf("config %s: %w", path, ErrEmpty)
	}
	if c.Timeout <= 0 {
		c.Timeout = 2 * time.Second
	}
	return c, nil
}
