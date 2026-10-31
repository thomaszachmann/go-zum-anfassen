package logs

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// combinedRe ist klein geschrieben, also nur in Paket logs sichtbar.
var combinedRe = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+)[^"]*" (\d{3}) (\d+|-)`)

const combinedTime = "02/Jan/2006:15:04:05 -0700"

// Combined liest das Combined Log Format von nginx und Apache.
type Combined struct{}

// Compiler prüft: Combined erfüllt Parser.
var _ Parser = Combined{}

// Parse implementiert Parser.
func (Combined) Parse(line string) (Entry, error) {
	m := combinedRe.FindStringSubmatch(line)
	if m == nil {
		return Entry{}, fmt.Errorf("kein Combined-Format: %q", line)
	}
	t, err := time.Parse(combinedTime, m[2])
	if err != nil {
		return Entry{}, err
	}
	status, _ := strconv.Atoi(m[5]) // Regex garantiert drei Ziffern
	var size int64
	if m[6] != "-" {
		size, _ = strconv.ParseInt(m[6], 10, 64)
	}
	return Entry{
		IP: m[1], Time: t, Method: m[3], Path: m[4],
		Status: status, Bytes: size,
	}, nil
}
