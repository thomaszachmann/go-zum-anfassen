package semver

import "fmt"

// Bump erhöht part ("major", "minor" oder "patch") wie npm und Cargo:
// Eine Prerelease wird zuerst "fertig", 1.3.0-rc.1 +minor = 1.3.0.
func Bump(v Version, part string) (Version, error) {
	pre := len(v.Pre) > 0
	switch part {
	case "major":
		if !pre || v.Minor != 0 || v.Patch != 0 {
			v.Major, v.Minor, v.Patch = v.Major+1, 0, 0
		}
	case "minor":
		if !pre || v.Patch != 0 {
			v.Minor, v.Patch = v.Minor+1, 0
		}
	case "patch":
		if !pre {
			v.Patch++
		}
	default:
		return v, fmt.Errorf("unbekannter Teil %q (major, minor, patch)", part)
	}
	v.Pre, v.Build = nil, ""
	return v, nil
}
