// Package database owns SQLite connection policy, not any workflow schema.
package database

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	dsn := url.URL{Scheme: "file", Path: path}
	params := url.Values{}
	for _, setting := range []string{"foreign_keys=ON", "synchronous=FULL", "busy_timeout=3000"} {
		params.Add("_pragma", setting)
	}
	dsn.RawQuery = params.Encode()
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func Transaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
