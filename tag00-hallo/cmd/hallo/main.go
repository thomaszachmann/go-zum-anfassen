// hallo begrüßt dich und verrät seine Version.
// Tag 0 aus "Go zum Anfassen".
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

// version setzt der Build per -ldflags "-X main.version=...".
var version = "dev"

func main() {
	name := flag.String("name", "Welt", "wen hallo begrüßt")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	flag.Parse()

	if *showVersion {
		fmt.Printf("hallo %s (%s %s/%s)\n",
			version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return
	}
	if *name == "" {
		fmt.Fprintln(os.Stderr, "hallo: -name darf nicht leer sein")
		os.Exit(2)
	}
	fmt.Printf("Hallo, %s!\n", *name)
}
