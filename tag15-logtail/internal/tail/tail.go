// Package tail folgt einer Logdatei wie tail -F – per Polling, ohne fsnotify.
package tail

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"time"
)

// Last liefert die letzten n Zeilen von path und das Byte-Offset danach.
func Last(path string, n int) ([]string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	var ring []string // höchstens n Zeilen im Speicher, auch bei GB-Logs
	sc := bufio.NewScanner(io.LimitReader(f, info.Size()))
	for sc.Scan() {
		ring = append(ring, sc.Text())
		if len(ring) > n {
			ring = ring[1:]
		}
	}
	return ring, info.Size(), sc.Err()
}

// follower merkt sich offene Datei, Position und eine angefangene Zeile.
type follower struct {
	path    string
	f       *os.File
	info    os.FileInfo
	off     int64
	partial []byte
	out     io.Writer
}

// Follow schreibt ab Offset off jede neue, vollständige Zeile nach out,
// bis ctx endet. Erkennt Truncate (Datei kleiner) und Rotation (neue Datei).
func Follow(ctx context.Context, path string, off int64, every time.Duration,
	out io.Writer) error {
	fw := &follower{path: path, off: off, out: out}
	if err := fw.open(); err != nil {
		return err
	}
	defer func() { fw.f.Close() }()
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		if err := fw.read(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil // Strg+C ist kein Fehler
		case <-t.C:
		}
		cur, err := os.Stat(path)
		switch {
		case err != nil: // gerade rotiert, neue Datei fehlt noch
		case !os.SameFile(fw.info, cur): // Rotation: Rest lesen, neu öffnen
			if err := fw.read(); err != nil {
				return err
			}
			fw.f.Close()
			fw.off, fw.partial = 0, nil
			if err := fw.open(); err != nil {
				return err
			}
		case cur.Size() < fw.off: // copytruncate: von vorn
			fw.off, fw.partial = 0, nil
		}
	}
}

func (fw *follower) open() error {
	f, err := os.Open(fw.path)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	fw.f, fw.info = f, info
	return nil
}

// read liest alles ab fw.off und gibt nur Zeilen mit \n aus.
func (fw *follower) read() error {
	data, err := io.ReadAll(io.NewSectionReader(fw.f, fw.off, 1<<62))
	if err != nil {
		return err
	}
	fw.off += int64(len(data))
	data = append(fw.partial, data...)
	end := bytes.LastIndexByte(data, '\n') + 1
	if _, err := fw.out.Write(data[:end]); err != nil {
		return err
	}
	fw.partial = append([]byte(nil), data[end:]...)
	return nil
}
