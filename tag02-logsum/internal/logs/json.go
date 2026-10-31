package logs

import (
	"encoding/json"
	"fmt"
	"time"
)

// JSON liest ein JSON-Objekt pro Zeile (z. B. nginx mit escape=json).
type JSON struct{}

var _ Parser = JSON{}

// jsonLine beschreibt das Format; die Tags ordnen JSON-Keys den Feldern zu.
type jsonLine struct {
	Time   time.Time `json:"ts"`
	IP     string    `json:"ip"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Status int       `json:"status"`
	Bytes  int64     `json:"bytes"`
}

// Parse implementiert Parser.
func (JSON) Parse(line string) (Entry, error) {
	var j jsonLine
	if err := json.Unmarshal([]byte(line), &j); err != nil {
		return Entry{}, err
	}
	if j.Path == "" || j.Status == 0 {
		return Entry{}, fmt.Errorf("path oder status fehlt: %q", line)
	}
	return Entry{
		IP: j.IP, Time: j.Time, Method: j.Method, Path: j.Path,
		Status: j.Status, Bytes: j.Bytes,
	}, nil
}
