package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func version(t *testing.T, db *sql.DB) int {
	t.Helper()
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMigrateRollback(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fsys := fstest.MapFS{
		"migrations/001_a.sql": {Data: []byte("CREATE TABLE a (x INTEGER);")},
		"migrations/002_b.sql": {Data: []byte(
			"CREATE TABLE b (x INTEGER); KAPUTT;")}, // zweite Anweisung scheitert
	}
	if err := migrate(ctx, db, fsys); err == nil {
		t.Fatal("Fehler erwartet")
	}
	if v := version(t, db); v != 1 {
		t.Errorf("user_version = %d, want 1", v)
	}
	var n int
	err = db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name = 'b'").Scan(&n)
	if err != nil || n != 0 {
		t.Error("Tabelle b existiert: Rollback hat nicht gegriffen")
	}
	delete(fsys, "migrations/002_b.sql") // repariert: zweiter Lauf ist ein No-op
	if err := migrate(ctx, db, fsys); err != nil || version(t, db) != 1 {
		t.Errorf("zweiter Lauf: err=%v, version=%d", err, version(t, db))
	}
}
