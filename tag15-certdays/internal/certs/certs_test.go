package certs

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

// selfSigned erzeugt ein Lab-Zertifikat als PEM, gültig bis notAfter.
func selfSigned(t *testing.T, cn string, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    now.AddDate(0, -1, 0),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestDays(t *testing.T) {
	tests := []struct {
		notAfter time.Time
		want     int
	}{
		{now.Add(10*24*time.Hour + time.Hour), 10},
		{now.Add(23 * time.Hour), 0},
		{now.Add(-time.Hour), -1}, // abgelaufen: negativ
	}
	for _, tt := range tests {
		if got := Days(tt.notAfter, now); got != tt.want {
			t.Errorf("Days(%v) = %d, will %d", tt.notAfter, got, tt.want)
		}
	}
}

func TestFromPEM(t *testing.T) {
	data := append(selfSigned(t, "a.lab", now.AddDate(0, 0, 30)),
		selfSigned(t, "b.lab", now.Add(-48*time.Hour))...)
	infos, err := FromPEM("bundle.pem", data, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 2 {
		t.Fatalf("%d Zertifikate, will 2", len(infos))
	}
	if infos[0].Subject != "CN=a.lab" || infos[0].Days != 30 ||
		infos[1].Days != -2 || infos[1].Source != "bundle.pem" {
		t.Errorf("infos = %+v", infos)
	}
}

func TestFromPEMOhneZertifikat(t *testing.T) {
	_, err := FromPEM("x.pem", []byte("kein PEM"), now)
	if !errors.Is(err, ErrNoCert) {
		t.Errorf("err = %v, will ErrNoCert", err)
	}
}

func TestFromAddr(t *testing.T) {
	srv := httptest.NewTLSServer(nil) // Testzertifikat von Go, lange gültig
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "https://")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	infos, err := FromAddr(ctx, addr, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 || infos[0].Days < 365 || infos[0].Source != addr {
		t.Errorf("infos = %+v", infos)
	}
}
