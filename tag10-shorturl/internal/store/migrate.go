package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrations embed.FS

// migrate spielt alle Dateien NNN_name.sql ein, deren Nummer größer als
// PRAGMA user_version ist – jede in einer eigenen Transaktion.
func migrate(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	var current int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").
		Scan(&current); err != nil {
		return err
	}
	files, err := fs.Glob(fsys, "migrations/*.sql") // sortiert
	if err != nil {
		return err
	}
	for _, f := range files {
		num, _, _ := strings.Cut(path.Base(f), "_")
		v, err := strconv.Atoi(num)
		if err != nil {
			return fmt.Errorf("migration %s: Name muss mit Nummer beginnen", f)
		}
		if v <= current {
			continue
		}
		if err := apply(ctx, db, fsys, f, v); err != nil {
			return fmt.Errorf("migration %s: %w", f, err)
		}
	}
	return nil
}

func apply(ctx context.Context, db *sql.DB, fsys fs.FS, f string, v int) error {
	stmt, err := fs.ReadFile(fsys, f)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // nach Commit ein No-op
	if _, err := tx.ExecContext(ctx, string(stmt)); err != nil {
		return err
	}
	// PRAGMA kennt keine Platzhalter; v ist eine geprüfte Zahl.
	q := fmt.Sprintf("PRAGMA user_version = %d", v)
	if _, err := tx.ExecContext(ctx, q); err != nil {
		return err
	}
	return tx.Commit()
}
