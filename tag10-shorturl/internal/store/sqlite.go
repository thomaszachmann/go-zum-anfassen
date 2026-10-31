package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// SQLite speichert Links dauerhaft in einer Datei – ohne CGO.
type SQLite struct {
	db *sql.DB
}

var _ Store = (*SQLite)(nil)

// OpenSQLite öffnet (oder erzeugt) die Datenbank und migriert sie.
func OpenSQLite(ctx context.Context, file string) (*SQLite, error) {
	dsn := "file:" + file + "?_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := migrate(ctx, db, migrations); err != nil {
		db.Close() //nolint:errcheck // erster Fehler zählt
		return nil, err
	}
	return &SQLite{db: db}, nil
}

// Close schließt alle Verbindungen.
func (s *SQLite) Close() error { return s.db.Close() }

func (s *SQLite) Create(ctx context.Context, l Link) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO links (code, url, hits, created) VALUES (?, ?, ?, ?)",
		l.Code, l.URL, l.Hits, l.Created.UnixNano())
	if e, ok := errors.AsType[*sqlite.Error](err); ok &&
		e.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
		return ErrExists
	}
	return err
}

func (s *SQLite) Get(ctx context.Context, code string) (Link, error) {
	return scan(s.db.QueryRowContext(ctx,
		"SELECT code, url, hits, created FROM links WHERE code = ?", code))
}

// Hit zählt hoch und liest in einem Statement – atomar ohne Transaktion.
func (s *SQLite) Hit(ctx context.Context, code string) (Link, error) {
	return scan(s.db.QueryRowContext(ctx, `UPDATE links SET hits = hits + 1
		WHERE code = ? RETURNING code, url, hits, created`, code))
}

func (s *SQLite) Delete(ctx context.Context, code string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM links WHERE code = ?", code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// scan liest eine Zeile; keine Zeile heißt ErrNotFound.
func scan(row *sql.Row) (Link, error) {
	var l Link
	var created int64
	err := row.Scan(&l.Code, &l.URL, &l.Hits, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return Link{}, ErrNotFound
	}
	l.Created = time.Unix(0, created).UTC()
	return l, err
}
