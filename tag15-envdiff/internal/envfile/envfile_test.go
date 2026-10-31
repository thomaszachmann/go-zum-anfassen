package envfile

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	in := `# Kommentar
APP_ENV=prod

export DB_HOST = db1.lab
DB_PASS="lab pw # kein Kommentar"
TOKEN='sk-lab-1234'
PORT=8080 # Kommentar am Ende
LEER=
URL=https://x.lab/?a=b
`
	got, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"APP_ENV": "prod", "DB_HOST": "db1.lab", "DB_PASS": "lab pw # kein Kommentar",
		"TOKEN": "sk-lab-1234", "PORT": "8080", "LEER": "", "URL": "https://x.lab/?a=b",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got %q\nwant %q", got, want)
	}
}

func TestParseFehler(t *testing.T) {
	for _, in := range []string{"A=1\nB-C=2\n", "A=1\nnur text\n", "A=1\n1X=2\n"} {
		_, err := Parse(strings.NewReader(in))
		if err == nil || !strings.Contains(err.Error(), "zeile 2") {
			t.Errorf("Parse(%q) err = %v, will Fehler mit zeile 2", in, err)
		}
	}
}

func TestDiff(t *testing.T) {
	a := map[string]string{"A": "1", "B": "2", "C": "3"}
	b := map[string]string{"B": "2", "C": "33", "D": "4"}
	want := []Change{
		{Key: "A", Kind: Removed, Old: "1"},
		{Key: "C", Kind: Changed, Old: "3", New: "33"},
		{Key: "D", Kind: Added, New: "4"},
	}
	if got := Diff(a, b); !reflect.DeepEqual(got, want) {
		t.Errorf("\n got %+v\nwant %+v", got, want)
	}
	if got := Diff(a, a); len(got) != 0 {
		t.Errorf("Diff(a, a) = %+v, will leer", got)
	}
}

func TestMask(t *testing.T) {
	tests := map[string]string{
		"": `""`, "geheim": "<6 Zeichen>", "Grüße": "<5 Zeichen>",
	}
	for in, want := range tests {
		if got := Mask(in); got != want {
			t.Errorf("Mask(%q) = %q, will %q", in, got, want)
		}
	}
}
