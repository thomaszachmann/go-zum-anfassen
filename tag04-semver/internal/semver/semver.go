// Package semver parst, vergleicht und erhöht Versionen nach semver.org 2.0.0.
package semver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalid meldet eine ungültige Version.
var ErrInvalid = errors.New("ungültige Version")

// Version ist eine semantische Version wie v1.4.2-rc.1+build.7.
type Version struct {
	Major, Minor, Patch int
	Pre                 []string // Prerelease, z. B. ["rc", "1"]
	Build               string   // Build-Metadaten, zählen beim Vergleich nicht
	V                   bool     // mit führendem "v" geschrieben
}

// Parse zerlegt s. Das führende "v" ist optional.
func Parse(s string) (Version, error) {
	var v Version
	rest, hasV := strings.CutPrefix(s, "v")
	v.V = hasV
	rest, build, hasBuild := strings.Cut(rest, "+")
	core, pre, hasPre := strings.Cut(rest, "-")

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("%w %q: brauche MAJOR.MINOR.PATCH", ErrInvalid, s)
	}
	nums := [3]*int{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		n, err := number(p)
		if err != nil {
			return Version{}, fmt.Errorf("%w %q: %v", ErrInvalid, s, err)
		}
		*nums[i] = n
	}
	if hasPre {
		v.Pre = strings.Split(pre, ".")
		for _, id := range v.Pre {
			if err := identifier(id); err != nil {
				return Version{}, fmt.Errorf("%w %q: %v", ErrInvalid, s, err)
			}
		}
	}
	if hasBuild {
		for id := range strings.SplitSeq(build, ".") {
			if id == "" || strings.Trim(id, idChars) != "" {
				return Version{}, fmt.Errorf("%w %q: Build %q", ErrInvalid, s, id)
			}
		}
		v.Build = build
	}
	return v, nil
}

const idChars = "0123456789" +
	"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"

// number prüft eine Zahl ohne führende Nullen.
func number(s string) (int, error) {
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, fmt.Errorf("%q ist keine Zahl", s)
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("%q hat führende Null", s)
	}
	return strconv.Atoi(s)
}

// identifier prüft einen Prerelease-Teil wie "rc" oder "1".
func identifier(id string) error {
	if id == "" || strings.Trim(id, idChars) != "" {
		return fmt.Errorf("prerelease %q ungültig", id)
	}
	if isNum(id) && len(id) > 1 && id[0] == '0' {
		return fmt.Errorf("prerelease %q hat führende Null", id)
	}
	return nil
}

func isNum(s string) bool { return strings.Trim(s, "0123456789") == "" }

// String setzt die Version wieder zusammen.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.V {
		s = "v" + s
	}
	if len(v.Pre) > 0 {
		s += "-" + strings.Join(v.Pre, ".")
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}
