package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunTabelle(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"tag_name":"v2.0.0-rc.1","name":"RC","prerelease":true,
			"published_at":"2026-10-02T08:00:00Z"}]`)
	}
	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()
	env := map[string]string{"GHREL_API": srv.URL}

	var out bytes.Buffer
	code := run(context.Background(), []string{"acme/tool"},
		func(k string) string { return env[k] }, &out, &out)
	want := "TAG          DATUM       NAME\nv2.0.0-rc.1  2026-10-02  RC (pre)\n"
	if code != 0 || out.String() != want {
		t.Errorf("Exit %d, Ausgabe:\n%s\nwant:\n%s", code, out.String(), want)
	}
}

func TestRunFalscherAufruf(t *testing.T) {
	var out bytes.Buffer
	getenv := func(string) string { return "" }
	code := run(context.Background(), []string{"nurname"}, getenv, &out, &out)
	if code != 2 {
		t.Errorf("Exit %d, want 2", code)
	}
}
