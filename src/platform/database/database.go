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
	// SQLite creates WAL/SHM with the database's permissions. Restrict the file
	// before opening any connection, including upgrades of existing installations.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = file.Chmod(0600); err != nil {
		file.Close()
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if err = os.Chmod(path+suffix, 0600); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
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
