package main

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// counts hält die Zählerstände einer Eingabe (mehr zu Structs an Tag 2).
type counts struct {
	lines, words, bytes, chars int
}

// count liest r Rune für Rune. Ist freq nicht nil, zählt es jedes Wort mit.
func count(r io.Reader, freq map[string]int) (counts, error) {
	var c counts
	var word strings.Builder
	inWord := false

	flush := func() { // Wort abschließen und in die Map eintragen
		w := strings.TrimFunc(word.String(), unicode.IsPunct)
		if freq != nil && w != "" {
			freq[w]++
		}
		word.Reset()
	}

	br := bufio.NewReader(r)
	for {
		ch, size, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c, err
		}
		c.bytes += size
		c.chars++
		if ch == '\n' {
			c.lines++
		}
		if unicode.IsSpace(ch) {
			if inWord {
				flush()
			}
			inWord = false
			continue
		}
		if !inWord {
			c.words++
			inWord = true
		}
		word.WriteRune(unicode.ToLower(ch))
	}
	if inWord {
		flush()
	}
	return c, nil
}
