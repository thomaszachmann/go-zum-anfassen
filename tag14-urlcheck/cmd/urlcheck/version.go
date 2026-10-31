package main

import (
	"runtime"
	"runtime/debug"
)

// Setzt der Build per -ldflags "-X main.version=1.2.3 -X main.commit=abc1234".
var (
	version = "dev"
	commit  = ""
)

// versionString nimmt die ldflags-Werte, sonst die VCS-Daten, die go build
// automatisch einbettet, wenn im Git-Repo gebaut wird.
func versionString() string {
	c := commit
	if bi, ok := debug.ReadBuildInfo(); ok && c == "" {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				c = s.Value[:7]
			}
		}
	}
	if c == "" {
		c = "unbekannt"
	}
	return version + " (" + c + ", " + runtime.Version() + ")"
}
