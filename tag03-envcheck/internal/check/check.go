package check

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// Tool sucht name im PATH und liefert den gefundenen Pfad.
func Tool(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", &Error{Kind: "tool", Name: name, Err: ErrNotFound}
	}
	return path, nil
}

// Env prüft, ob die Umgebungsvariable name gesetzt und nicht leer ist.
func Env(name string) error {
	if os.Getenv(name) == "" {
		return &Error{Kind: "env", Name: name, Err: ErrNotSet}
	}
	return nil
}

// Port prüft, ob auf addr (host:port) ein Dienst TCP-Verbindungen annimmt.
func Port(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		// errors.As holt den *net.OpError aus der Kette; sein Err ist kürzer.
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			err = opErr.Err
		}
		// Zwei %w: Sentinel UND Ursache bleiben für errors.Is erreichbar.
		cause := fmt.Errorf("%w (%w)", ErrUnreachable, err)
		return &Error{Kind: "port", Name: addr, Err: cause}
	}
	return conn.Close()
}
