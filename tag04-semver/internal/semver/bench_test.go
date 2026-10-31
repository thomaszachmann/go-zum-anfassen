package semver

import "testing"

func BenchmarkParse(b *testing.B) {
	for b.Loop() { // seit Go 1.24: misst nur den Schleifenrumpf
		if _, err := Parse("v1.42.7-rc.3+build.2026"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompare(b *testing.B) {
	x, _ := Parse("1.0.0-beta.11")
	y, _ := Parse("1.0.0-beta.2")
	for b.Loop() {
		Compare(x, y)
	}
}
