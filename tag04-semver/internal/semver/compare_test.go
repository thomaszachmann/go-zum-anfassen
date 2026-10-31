package semver

import "testing"

// mustParse ist ein Test-Helfer: t.Helper() lässt Fehler beim Aufrufer zeigen.
func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return v
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.10.0", -1},
		{"2.0.0", "1.99.99", 1},
		{"1.0.0-rc.1", "1.0.0", -1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-1", "1.0.0-alpha", -1},
		{"1.0.0+build.1", "1.0.0+build.2", 0},
		{"v1.0.0", "1.0.0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a, b := mustParse(t, tt.a), mustParse(t, tt.b)
			if got := Compare(a, b); got != tt.want {
				t.Errorf("Compare = %d, will %d", got, tt.want)
			}
			if got := Compare(b, a); got != -tt.want {
				t.Errorf("umgekehrt = %d, will %d", got, -tt.want)
			}
		})
	}
}

func TestBump(t *testing.T) {
	tests := []struct{ in, part, want string }{
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3", "minor", "1.3.0"},
		{"v1.2.3", "major", "v2.0.0"},
		{"1.2.3-rc.1", "patch", "1.2.3"},
		{"1.3.0-rc.1", "minor", "1.3.0"},
		{"1.3.1-rc.1", "minor", "1.4.0"},
		{"2.0.0-beta", "major", "2.0.0"},
		{"1.2.3+git.abc", "patch", "1.2.4"},
	}
	for _, tt := range tests {
		t.Run(tt.in+"_"+tt.part, func(t *testing.T) {
			got, err := Bump(mustParse(t, tt.in), tt.part)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tt.want {
				t.Errorf("Bump = %s, will %s", got, tt.want)
			}
		})
	}
	if _, err := Bump(Version{}, "micro"); err == nil {
		t.Error("Bump(micro) ohne Fehler")
	}
}
