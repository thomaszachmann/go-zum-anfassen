// Package check prüft Werkzeuge, Umgebungsvariablen und Ports.
package check

import (
	"errors"
	"fmt"
)

// Sentinel-Fehler: feste Werte, vergleichbar mit errors.Is.
var (
	ErrNotFound    = errors.New("nicht im PATH")
	ErrNotSet      = errors.New("nicht gesetzt")
	ErrUnreachable = errors.New("nicht erreichbar")
)

// Error ist ein fehlgeschlagener Check. errors.As liefert dir die Felder.
type Error struct {
	Kind string // "tool", "env" oder "port"
	Name string // z. B. "kubectl", "HOME", "127.0.0.1:8099"
	Err  error  // die Ursache, meist ein Sentinel
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Kind, e.Name, e.Err)
}

// Unwrap macht die Ursache für errors.Is und errors.As sichtbar.
func (e *Error) Unwrap() error { return e.Err }
