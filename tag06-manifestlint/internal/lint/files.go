package lint

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Files sammelt Manifeste: genannte Dateien immer, in Verzeichnissen nur
// *.yaml/*.yml. Versteckte Verzeichnisse (.git) werden übersprungen.
func Files(paths []string) ([]string, error) {
	var out []string
	for _, root := range paths {
		if root == "-" { // stdin
			out = append(out, root)
			continue
		}
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case d.IsDir() && p != root && strings.HasPrefix(d.Name(), "."):
				return filepath.SkipDir
			case !d.IsDir() && (p == root || isYAML(p)):
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func isYAML(p string) bool {
	ext := filepath.Ext(p)
	return ext == ".yaml" || ext == ".yml"
}

// LintFile liest eine Datei (oder stdin) und prüft alle Dokumente darin.
func LintFile(path string, labels []string) ([]Finding, error) {
	var r io.Reader = os.Stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	ms, err := Decode(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var out []Finding
	for _, m := range ms {
		for _, f := range Check(m, labels) {
			f.File = path
			out = append(out, f)
		}
	}
	return out, nil
}
