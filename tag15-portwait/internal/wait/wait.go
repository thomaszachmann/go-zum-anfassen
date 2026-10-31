// Package wait wartet auf TCP-Ports – mit context und exponentiellem Backoff.
package wait

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Backoff liefert die Pause nach dem Fehlversuch n (ab 0): initial·2^n,
// höchstens limit. Die Schleife verdoppelt nur bis limit – kein Überlauf.
func Backoff(initial, limit time.Duration, n int) time.Duration {
	d := initial
	for range n {
		if d >= limit/2 {
			return limit
		}
		d *= 2
	}
	return min(d, limit)
}

// Options steuert Port. Nullwerte bekommen Defaults.
type Options struct {
	Initial time.Duration // erste Pause, Default 100ms
	Max     time.Duration // längste Pause, Default 2s
	Dial    func(ctx context.Context, network, addr string) (net.Conn, error)
	OnRetry func(attempt int, err error, wait time.Duration) // optional
}

// Port versucht addr zu erreichen, bis es klappt oder ctx endet,
// und liefert die Zahl der Versuche.
func Port(ctx context.Context, addr string, o Options) (int, error) {
	if o.Initial <= 0 {
		o.Initial = 100 * time.Millisecond
	}
	if o.Max <= 0 {
		o.Max = 2 * time.Second
	}
	if o.Dial == nil {
		var d net.Dialer
		o.Dial = d.DialContext
	}
	for attempt := 1; ; attempt++ {
		dctx, cancel := context.WithTimeout(ctx, time.Second) // pro Versuch
		conn, err := o.Dial(dctx, "tcp", addr)
		cancel()
		if err == nil {
			return attempt, conn.Close()
		}
		wait := Backoff(o.Initial, o.Max, attempt-1)
		if o.OnRetry != nil {
			o.OnRetry(attempt, err, wait)
		}
		select {
		case <-ctx.Done():
			return attempt, fmt.Errorf("%s: %w (letzter Fehler: %v)",
				addr, ctx.Err(), err)
		case <-time.After(wait):
		}
	}
}
