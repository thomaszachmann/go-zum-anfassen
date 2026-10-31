package semver

import (
	"cmp"
	"slices"
)

// Compare liefert -1, 0 oder +1 – passend für slices.SortFunc.
func Compare(a, b Version) int {
	if c := cmp.Or(
		cmp.Compare(a.Major, b.Major),
		cmp.Compare(a.Minor, b.Minor),
		cmp.Compare(a.Patch, b.Patch),
	); c != 0 {
		return c
	}
	// Ohne Prerelease ist größer: 1.0.0-rc.1 < 1.0.0
	switch {
	case len(a.Pre) == 0 && len(b.Pre) == 0:
		return 0
	case len(a.Pre) == 0:
		return 1
	case len(b.Pre) == 0:
		return -1
	}
	for i := range min(len(a.Pre), len(b.Pre)) {
		if c := compareID(a.Pre[i], b.Pre[i]); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(a.Pre), len(b.Pre)) // rc < rc.1
}

// compareID: Zahlen numerisch, Zahlen < Text, Text nach ASCII.
func compareID(a, b string) int {
	an, bn := isNum(a), isNum(b)
	switch {
	case an && bn: // ohne führende Nullen: kürzer = kleiner, kein Überlauf
		return cmp.Or(cmp.Compare(len(a), len(b)), cmp.Compare(a, b))
	case an:
		return -1
	case bn:
		return 1
	}
	return cmp.Compare(a, b)
}

// Sort sortiert vs aufsteigend.
func Sort(vs []Version) { slices.SortStableFunc(vs, Compare) }
