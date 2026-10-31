package semver

import (
	"strings"
	"testing"

	xsemver "golang.org/x/mod/semver"
)

// FuzzParse: Was Parse akzeptiert, muss String() exakt zurückgeben.
func FuzzParse(f *testing.F) {
	for _, s := range []string{"1.2.3", "v1.0.0-rc.1+b.7", "0.0.0-0.a-b"} {
		f.Add(s) // Seed-Korpus
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := Parse(s)
		if err != nil {
			return // ungültige Eingaben sind ok, nur kein Panic
		}
		if got := v.String(); got != s {
			t.Errorf("Parse(%q).String() = %q", s, got)
		}
		if Compare(v, v) != 0 {
			t.Errorf("Compare(%q, %q) != 0", s, s)
		}
	})
}

// FuzzCompare vergleicht mit golang.org/x/mod/semver (differentielles Fuzzing).
func FuzzCompare(f *testing.F) {
	f.Add("1.0.0-alpha.1", "1.0.0-alpha.beta")
	f.Add("1.0.0-beta.11", "1.0.0-beta.2")
	f.Fuzz(func(t *testing.T, a, b string) {
		va, errA := Parse(a)
		vb, errB := Parse(b)
		if errA != nil || errB != nil {
			return
		}
		ref := xsemver.Compare(withV(a), withV(b))
		if got := Compare(va, vb); got != ref {
			t.Errorf("Compare(%q, %q) = %d, x/mod sagt %d", a, b, got, ref)
		}
	})
}

func withV(s string) string { return "v" + strings.TrimPrefix(s, "v") }
