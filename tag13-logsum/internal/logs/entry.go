// Package logs liest Zugriffslogs und fasst sie zusammen.
package logs

import "time"

// Entry ist eine geparste Logzeile.
type Entry struct {
	IP     string
	Time   time.Time
	Method string
	Path   string
	Status int
	Bytes  int64
}

// Parser macht aus einer Zeile einen Entry. Jedes Logformat implementiert ihn.
type Parser interface {
	Parse(line string) (Entry, error)
}
