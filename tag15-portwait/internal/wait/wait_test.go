package wait

import (
	"context"
	"errors"
	"net"
	"slices"
	"syscall"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	ms := time.Millisecond
	tests := []struct {
		n    int
		want time.Duration
	}{
		{0, 100 * ms}, {1, 200 * ms}, {3, 800 * ms}, {4, 1600 * ms},
		{5, 2 * time.Second}, {100, 2 * time.Second}, // 100: kein Überlauf!
	}
	for _, tt := range tests {
		if got := Backoff(100*ms, 2*time.Second, tt.n); got != tt.want {
			t.Errorf("Backoff(n=%d) = %v, will %v", tt.n, got, tt.want)
		}
	}
}

func TestPortOffen(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	n, err := Port(context.Background(), ln.Addr().String(), Options{})
	if err != nil || n != 1 {
		t.Errorf("Port = %d, %v; will 1, nil", n, err)
	}
}

func TestPortRetryMitBackoff(t *testing.T) {
	calls := 0
	dial := func(ctx context.Context, _, _ string) (net.Conn, error) {
		calls++
		if calls <= 3 {
			return nil, syscall.ECONNREFUSED
		}
		c1, c2 := net.Pipe()
		c2.Close()
		return c1, nil
	}
	var waits []time.Duration
	o := Options{Initial: time.Millisecond, Max: 3 * time.Millisecond, Dial: dial,
		OnRetry: func(_ int, err error, d time.Duration) {
			if !errors.Is(err, syscall.ECONNREFUSED) {
				t.Errorf("OnRetry err = %v", err)
			}
			waits = append(waits, d)
		}}
	n, err := Port(context.Background(), "lab:1", o)
	if err != nil || n != 4 {
		t.Fatalf("Port = %d, %v; will 4, nil", n, err)
	}
	ms := time.Millisecond
	want := []time.Duration{ms, 2 * ms, 3 * ms} // verdoppelt, dann gedeckelt
	if !slices.Equal(waits, want) {
		t.Errorf("Wartezeiten = %v, will %v", waits, want)
	}
}

func TestPortDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // Port ist jetzt zu
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = Port(ctx, addr, Options{Initial: 10 * time.Millisecond})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, will DeadlineExceeded", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Abbruch nach %v – context ignoriert?", d)
	}
}
