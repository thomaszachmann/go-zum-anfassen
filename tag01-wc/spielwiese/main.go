// spielwiese: Go-Syntax im Schnelldurchlauf.
package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {
	// Variablen: var mit Typ (Nullwert), := mit Typinferenz, const
	var n int
	gruss := "Grüße"
	const limit = 3
	fmt.Println("1)", n, gruss, limit)

	// Strings sind Bytes, range liefert Runen (Unicode-Codepoints)
	fmt.Println("2)", len(gruss), utf8.RuneCountInString(gruss))
	for i, r := range gruss {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()

	// Slices: append wächst bei Bedarf, s[a:b] ist eine Sicht
	tools := []string{"kubectl", "helm"}
	tools = append(tools, "bao")
	fmt.Println("3)", tools, len(tools), tools[1:])

	// Maps: Zugriff mit comma-ok, fehlender Key liefert den Nullwert
	ports := map[string]int{"http": 80, "https": 443}
	p, ok := ports["ssh"]
	fmt.Println("4)", p, ok, len(ports))

	// Schleifen: es gibt nur for – klassisch, range, while-artig
	for i := range limit {
		fmt.Print(i, " ")
	}
	words := strings.Fields("  go   ist  schnell ")
	for len(words) > 0 {
		fmt.Print(strings.ToUpper(words[0]), " ")
		words = words[1:]
	}
	fmt.Println()
}
