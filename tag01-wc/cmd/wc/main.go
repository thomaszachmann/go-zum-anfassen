// wc zählt Zeilen, Wörter, Bytes und Zeichen – wie das Unix-Original.
// Tag 1 aus "Go zum Anfassen".
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	lines := flag.Bool("l", false, "Zeilen zählen")
	words := flag.Bool("w", false, "Wörter zählen")
	bytes := flag.Bool("c", false, "Bytes zählen")
	chars := flag.Bool("m", false, "Zeichen (Runen) zählen")
	top := flag.Int("top", 0, "die `N` häufigsten Wörter ausgeben")
	flag.Parse()

	// Ohne Auswahl wie das Original: Zeilen, Wörter, Bytes.
	if !*lines && !*words && !*bytes && !*chars {
		*lines, *words, *bytes = true, true, true
	}
	show := func(c counts, name string) {
		if *lines {
			fmt.Printf("%8d", c.lines)
		}
		if *words {
			fmt.Printf("%8d", c.words)
		}
		if *chars {
			fmt.Printf("%8d", c.chars)
		}
		if *bytes {
			fmt.Printf("%8d", c.bytes)
		}
		if name != "-" {
			fmt.Print(" ", name)
		}
		fmt.Println()
	}

	var freq map[string]int // nil-Map: lesen ok, schreiben verboten
	if *top > 0 {
		freq = make(map[string]int)
	}
	files := flag.Args()
	if len(files) == 0 {
		files = []string{"-"} // "-" steht für stdin
	}

	var total counts
	exitCode := 0
	for _, name := range files {
		c, err := countFile(name, freq)
		if err != nil {
			fmt.Fprintln(os.Stderr, "wc:", err)
			exitCode = 1
			continue
		}
		show(c, name)
		total.lines += c.lines
		total.words += c.words
		total.bytes += c.bytes
		total.chars += c.chars
	}
	if len(files) > 1 {
		show(total, "total")
	}
	for i, wc := range topWords(freq, *top) {
		fmt.Printf("%3d. %-20s %d\n", i+1, wc.word, wc.n)
	}
	os.Exit(exitCode)
}

// countFile öffnet name (oder stdin bei "-") und zählt.
func countFile(name string, freq map[string]int) (counts, error) {
	if name == "-" {
		return count(os.Stdin, freq)
	}
	f, err := os.Open(name)
	if err != nil {
		return counts{}, err
	}
	defer f.Close() //nolint:errcheck // nur gelesen, Close-Fehler egal
	return count(f, freq)
}
