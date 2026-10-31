package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestReadURLs(t *testing.T) {
	in := "# Kommentar\nhttps://a.example\n\n  https://b.example  \n"
	got, err := readURLs(nil, strings.NewReader(in))
	want := []string{"https://a.example", "https://b.example"}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("readURLs = %v, %v; want %v", got, err, want)
	}
}

func TestRunExitCode(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	var out bytes.Buffer
	code := run(context.Background(), []string{"-json"},
		strings.NewReader(srv.URL+"\n"), &out, &out)
	if code != 1 || !strings.Contains(out.String(), `"status":404`) {
		t.Errorf("Exit %d, Ausgabe %s", code, out.String())
	}
}
