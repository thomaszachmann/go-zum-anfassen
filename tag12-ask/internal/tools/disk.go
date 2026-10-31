// Package tools enthält lokale Funktionen, die ein Modell aufrufen darf.
package tools

import (
	"fmt"
	"math"
	"syscall"
)

// Disk ist das Ergebnis von DiskFree.
type Disk struct {
	Path    string  `json:"path"`
	FreeGiB float64 `json:"free_gib"`
	FreePct float64 `json:"free_pct"`
}

// DiskFree misst den freien Platz des Dateisystems unter path (Unix).
func DiskFree(path string) (Disk, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return Disk{}, fmt.Errorf("statfs %s: %w", path, err)
	}
	free := float64(st.Bavail) * float64(st.Bsize)
	total := float64(st.Blocks) * float64(st.Bsize)
	return Disk{path, round(free / (1 << 30)), round(100 * free / total)}, nil
}

func round(f float64) float64 { return math.Round(f*10) / 10 }
