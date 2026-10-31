package logs

import "slices"

// word liest druckbare ASCII-Zeichen ohne '"' bis zu einem der Trenner.
// Ist der Trenner '"', bleibt er in rest.
func word(b []byte, seps ...byte) (w, rest []byte, ok bool) {
	for i, c := range b {
		if c > ' ' && c <= '~' && c != '"' {
			continue // der häufige Fall zuerst: ein normales Zeichen
		}
		if i == 0 || !slices.Contains(seps, c) { // generisch, inline
			return nil, nil, false
		}
		if c == '"' {
			return b[:i], b[i:], true
		}
		return b[:i], b[i+1:], true
	}
	return nil, nil, false
}

// digitRun zählt führende Ziffern.
func digitRun(b []byte) int {
	n := 0
	for n < len(b) && b[n] >= '0' && b[n] <= '9' {
		n++
	}
	return n
}

// digits parst reine Ziffern; -1 sonst.
func digits(b []byte) int {
	if digitRun(b) != len(b) {
		return -1
	}
	n := 0
	for _, c := range b {
		n = n*10 + int(c-'0')
	}
	return n
}
