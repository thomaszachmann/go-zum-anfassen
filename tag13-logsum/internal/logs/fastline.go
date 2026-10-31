package logs

import "bytes"

// add versteht nur „brave" ASCII-Zeilen und zählt sie direkt in s.
// false heißt: nichts verändert, der Aufrufer nimmt den Regex-Fallback.
func (f *fastParser) add(s *Summary, line []byte) bool {
	rest, ok := line, false
	for range 3 { // IP, ident, user
		if _, rest, ok = word(rest, ' '); !ok {
			return false
		}
	}
	if !bytes.HasPrefix(rest, []byte("[")) {
		return false
	}
	end := bytes.IndexByte(rest, ']')
	if end < 2 || !bytes.HasPrefix(rest[end:], []byte(`] "`)) {
		return false
	}
	ts := rest[1:end]
	if _, rest, ok = word(rest[end+3:], ' '); !ok { // Methode
		return false
	}
	path, rest, ok := word(rest, ' ', '"')
	q := bytes.IndexByte(rest, '"')
	if !ok || q < 0 {
		return false
	}
	rest = rest[q+1:] // jetzt: ` 200 512 ...`
	if len(rest) < 6 || rest[0] != ' ' || rest[4] != ' ' {
		return false
	}
	status, size := digits(rest[1:4]), int64(0)
	if n := digitRun(rest[5:]); n > 0 && n <= 18 {
		size = int64(digits(rest[5 : 5+n]))
	} else if n != 0 || rest[5] != '-' {
		return false
	}
	t, err := f.tm.parse(ts)
	if status < 0 || err != nil {
		return false
	}
	s.Bytes += size
	s.ByStatus[status]++
	s.ByPath[f.intern(path)]++
	if s.First.IsZero() || t.Before(s.First) {
		s.First = t
	}
	if t.After(s.Last) {
		s.Last = t
	}
	return true
}
