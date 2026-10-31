package query

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

const doc = `{"name":"shop","items":[{"name":"a","tags":["x"]},
	{"name":"b","price":2}],"meta":null}`

func TestEval(t *testing.T) {
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		expr string
		want []any
	}{
		{".", []any{v}},
		{".name", []any{"shop"}},
		{".items[0].name", []any{"a"}},
		{".items[].name", []any{"a", "b"}},
		{".items[-1].price", []any{2.0}},
		{".items[5]", []any{nil}},
		{".fehlt.x", []any{nil}},
		{".meta.x", []any{nil}},
		{".items[].tags[0]", []any{"x", nil}},
		{".items[1][]", []any{"b", 2.0}}, // Objekt: Werte nach Schlüssel sortiert
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := Eval(v, tt.expr)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, will %#v", got, tt.want)
			}
		})
	}
}

func TestEvalFehler(t *testing.T) {
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		expr string
		want error
	}{
		{"name", ErrSyntax},
		{".a..b", ErrSyntax},
		{".a[x]", ErrSyntax},
		{".a[1", ErrSyntax},
		{".name[0]", ErrType},
		{".name.x", ErrType},
		{".name[]", ErrType},
	}
	for _, tt := range tests {
		if _, err := Eval(v, tt.expr); !errors.Is(err, tt.want) {
			t.Errorf("Eval(%q) err = %v, will %v", tt.expr, err, tt.want)
		}
	}
}

func FuzzEval(f *testing.F) {
	f.Add(".items[].name")
	f.Add(".a[-1][]")
	var v any
	_ = json.Unmarshal([]byte(doc), &v)
	f.Fuzz(func(t *testing.T, expr string) {
		_, _ = Eval(v, expr) // darf Fehler liefern, aber nie paniken
	})
}
