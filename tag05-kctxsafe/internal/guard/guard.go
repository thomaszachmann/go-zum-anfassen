// Package guard schützt Produktionskontexte vor versehentlichem Wechsel.
package guard

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Confirm fragt nach dem Kontextnamen. Nur exakt derselbe Name bestätigt.
func Confirm(in io.Reader, out io.Writer, name string) (bool, error) {
	fmt.Fprintf(out, "%s ist prod. Namen zum Bestätigen tippen: ", name)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	return strings.TrimSpace(line) == name, nil
}
