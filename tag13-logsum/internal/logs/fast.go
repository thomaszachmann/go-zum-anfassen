package logs

import (
	"bytes"
	"iter"
	"time"
)

// Lines liefert die Zeilen von data ohne Kopie; ein \r am Zeilenende fällt
// weg – wie bei bufio.ScanLines.
func Lines(data []byte) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for len(data) > 0 {
			line, rest, _ := bytes.Cut(data, []byte{'\n'})
			if !yield(bytes.TrimSuffix(line, []byte{'\r'})) {
				return
			}
			data = rest
		}
	}
}

// ReadFast liefert dasselbe Ergebnis wie Read(r, Combined{}) – schneller.
func ReadFast(data []byte) *Summary {
	s := NewSummary()
	f := fastParser{paths: map[string]string{}}
	for line := range Lines(data) {
		s.Lines++
		if f.add(s, line) {
			continue
		}
		e, err := Combined{}.Parse(string(line)) // Fallback: exakt wie Tag 2
		if err != nil {
			s.Bad++
			continue
		}
		s.Add(e)
	}
	return s
}

// fastParser hält die Caches, die das Tempo bringen.
type fastParser struct {
	tm    timeMemo
	paths map[string]string // jeder Pfad nur einmal als String
}

// intern liefert path als String und alloziert nur beim ersten Auftreten.
func (f *fastParser) intern(path []byte) string {
	if p, ok := f.paths[string(path)]; ok { // Lookup ohne Allokation
		return p
	}
	p := string(path)
	f.paths[p] = p
	return p
}

// timeMemo merkt sich den letzten Zeitstempel: Viele Zeilen teilen ihn.
type timeMemo struct {
	key []byte
	t   time.Time
	err error
}

func (m *timeMemo) parse(b []byte) (time.Time, error) {
	if m.key == nil || !bytes.Equal(b, m.key) {
		m.key = append(m.key[:0], b...)
		m.t, m.err = time.Parse(combinedTime, string(b))
	}
	return m.t, m.err
}
