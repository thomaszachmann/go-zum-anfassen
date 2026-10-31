package logs

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
)

// Summary sammelt Kennzahlen über viele Entries.
type Summary struct {
	Lines    int // gelesene Zeilen
	Bad      int // nicht lesbare Zeilen
	Bytes    int64
	First    time.Time
	Last     time.Time
	ByStatus map[int]int    // 200 -> 17
	ByPath   map[string]int // "/api" -> 4
}

// PathCount ist ein Eintrag der Top-Liste.
type PathCount struct {
	Path string
	N    int
}

// NewSummary liefert einen Summary mit initialisierten Maps.
func NewSummary() *Summary {
	return &Summary{ByStatus: map[int]int{}, ByPath: map[string]int{}}
}

// Read parst jede Zeile aus r mit p und zählt sie in einen neuen Summary.
func Read(r io.Reader, p Parser) (*Summary, error) {
	s := NewSummary()
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		s.Lines++
		e, err := p.Parse(sc.Text())
		if err != nil {
			s.Bad++
			continue
		}
		s.Add(e)
	}
	return s, sc.Err()
}

// Add zählt e. Pointer-Receiver, weil Add den Summary verändert.
func (s *Summary) Add(e Entry) {
	s.Bytes += e.Bytes
	s.ByStatus[e.Status]++
	s.ByPath[e.Path]++
	if s.First.IsZero() || e.Time.Before(s.First) {
		s.First = e.Time
	}
	if e.Time.After(s.Last) {
		s.Last = e.Time
	}
}

// Requests ist die Zahl der lesbaren Zeilen.
func (s *Summary) Requests() int { return s.Lines - s.Bad }

// ErrorRate liefert den Anteil der 5xx-Antworten in Prozent.
func (s *Summary) ErrorRate() float64 {
	if s.Requests() == 0 {
		return 0
	}
	n := 0
	for code, k := range s.ByStatus {
		if code >= 500 {
			n += k
		}
	}
	return 100 * float64(n) / float64(s.Requests())
}

// TopPaths liefert die n häufigsten Pfade.
func (s *Summary) TopPaths(n int) []PathCount {
	list := make([]PathCount, 0, len(s.ByPath))
	for p, k := range s.ByPath {
		list = append(list, PathCount{p, k})
	}
	slices.SortFunc(list, func(a, b PathCount) int {
		return cmp.Or(cmp.Compare(b.N, a.N), cmp.Compare(a.Path, b.Path))
	})
	return list[:min(n, len(list))]
}

// Report schreibt eine Übersicht nach w – Terminal, Datei oder Puffer.
func (s *Summary) Report(w io.Writer, top int) error {
	var b strings.Builder // sammeln, dann einmal schreiben
	fmt.Fprintf(&b, "Requests:   %d (%d Zeilen unlesbar)\n", s.Requests(), s.Bad)
	if s.Requests() > 0 {
		const layout = "2006-01-02 15:04:05"
		fmt.Fprintf(&b, "Zeitraum:   %s bis %s\n",
			s.First.Format(layout), s.Last.Format(layout))
		fmt.Fprintf(&b, "Bytes:      %d\n", s.Bytes)
		b.WriteString("Status:    ")
		for _, class := range []int{1, 2, 3, 4, 5} {
			n := 0
			for code, k := range s.ByStatus {
				if code/100 == class {
					n += k
				}
			}
			if n > 0 {
				fmt.Fprintf(&b, " %dxx=%d", class, n)
			}
		}
		fmt.Fprintf(&b, "\n5xx-Quote:  %.1f %%\n", s.ErrorRate())
		fmt.Fprintf(&b, "Top %d Pfade:\n", top)
		for _, pc := range s.TopPaths(top) {
			fmt.Fprintf(&b, "%8d  %s\n", pc.N, pc.Path)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
