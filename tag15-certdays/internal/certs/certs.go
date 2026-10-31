// Package certs liest Zertifikate aus PEM-Dateien oder von TLS-Servern
// und berechnet die Resttage.
package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"net"
	"time"
)

// ErrNoCert meldet eine Datei ohne CERTIFICATE-Block.
var ErrNoCert = errors.New("kein Zertifikat gefunden")

// Info beschreibt ein Zertifikat.
type Info struct {
	Source   string // Datei oder host:port
	Subject  string
	NotAfter time.Time
	Days     int // volle Tage bis Ablauf, negativ = abgelaufen
}

// Days liefert die vollen Tage von now bis notAfter (abgerundet).
func Days(notAfter, now time.Time) int {
	return int(math.Floor(notAfter.Sub(now).Hours() / 24))
}

func info(src string, c *x509.Certificate, now time.Time) Info {
	return Info{Source: src, Subject: c.Subject.String(), NotAfter: c.NotAfter,
		Days: Days(c.NotAfter, now)}
}

// FromPEM liest alle Zertifikate aus data (z. B. eine Chain oder ein Bundle).
func FromPEM(name string, data []byte, now time.Time) ([]Info, error) {
	var out []Info
	for {
		var b *pem.Block
		b, data = pem.Decode(data)
		if b == nil {
			break
		}
		if b.Type != "CERTIFICATE" {
			continue // z. B. PRIVATE KEY im selben File
		}
		c, err := x509.ParseCertificate(b.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out = append(out, info(name, c, now))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: %w", name, ErrNoCert)
	}
	return out, nil
}

// FromAddr holt das Server-Zertifikat von addr (host:port).
func FromAddr(ctx context.Context, addr string, now time.Time) ([]Info, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	d := tls.Dialer{Config: &tls.Config{
		ServerName: host,
		// Nur lesen: Auch abgelaufene oder selbst signierte Zertifikate
		// sollen gemeldet werden. Es werden keine Daten gesendet.
		InsecureSkipVerify: true,
	}}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	state := conn.(*tls.Conn).ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("%s: %w", addr, ErrNoCert)
	}
	return []Info{info(addr, state.PeerCertificates[0], now)}, nil
}
