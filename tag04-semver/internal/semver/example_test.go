package semver_test

import (
	"fmt"

	"example.com/semver/internal/semver"
)

// Beispiele sind Doku UND Test: go test vergleicht die Output-Zeilen.
func ExampleBump() {
	v, _ := semver.Parse("v1.4.2")
	next, _ := semver.Bump(v, "minor")
	fmt.Println(next)
	// Output: v1.5.0
}

func ExampleSort() {
	var vs []semver.Version
	for _, s := range []string{"1.10.0", "1.2.0", "1.2.0-rc.1"} {
		v, _ := semver.Parse(s)
		vs = append(vs, v)
	}
	semver.Sort(vs)
	fmt.Println(vs)
	// Output: [1.2.0-rc.1 1.2.0 1.10.0]
}
