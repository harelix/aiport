// Package store opens aiport's local SQLite database (<config>.db, 0600).
//
// The driver is modernc.org/sqlite: pure Go, compiled into the executable, so
// neither builders nor users need a C toolchain or a system SQLite library.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// Path is the database file kept next to the config (aiport.yaml -> aiport.yaml.db).
func Path(configPath string) string { return configPath + ".db" }

// ErrNewerSchema means the database was written by a newer aiport. Callers must
// not write to it (telemetry runs disabled instead of risking corruption).
var ErrNewerSchema = errors.New("database schema is newer than this aiport; upgrade aiport")

// Open creates or opens the database with private permissions, WAL journaling
// and a bounded busy timeout, then applies pending migrations.
func Open(path string) (*sql.DB, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, errors.New("cannot create database file " + path)
	}
	st, err := f.Stat()
	_ = f.Close()
	if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0077 != 0 {
		return nil, errors.New("database file must be private (chmod 600 " + path + ")")
	}
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)&_txlock=immediate", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, errors.New("cannot open database " + path)
	}
	db.SetMaxOpenConns(4)
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
