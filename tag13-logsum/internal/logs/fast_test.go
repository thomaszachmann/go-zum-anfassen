package logs

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// diff vergleicht ReadFast mit der Tag-2-Version als Orakel.
func diff(t *testing.T, in string) {
	t.Helper()
	want, err := Read(strings.NewReader(in), Combined{})
	if err != nil {
		return // z. B. Zeile länger als der Scanner-Puffer
	}
	if got := ReadFast([]byte(in)); !reflect.DeepEqual(got, want) {
		t.Errorf("%q:\ngot  %+v\nwant %+v", in, got, want)
	}
}

func TestFastAgrees(t *testing.T) {
	b, err := os.ReadFile("../../testdata/access.log")
	if err != nil {
		t.Fatal(err)
	}
	diff(t, string(b))
	diff(t, "kaputt\n\r\n1.2.3.4 - - [x] \"GET / HTTP/1.1\" 200 1\n")
}

func FuzzFastAgrees(f *testing.F) {
	f.Add(`10.0.0.1 - - [10/Oct/2026:08:00:01 +0200] "GET / HTTP/1.1" 200 512`)
	f.Add(`10.0.0.1 - - [10/Oct/2026:08:00:01 +0200] "GET /a"b" 200 -`)
	f.Fuzz(diff)
}

func BenchmarkReadFast(b *testing.B) {
	data := load(b)
	for b.Loop() {
		ReadFast(data)
	}
}
