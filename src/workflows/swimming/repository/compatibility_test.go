package repository

import (
	"database/sql"
	"os"
)

func OpenRawFixture(path string) (*sql.DB, error) {
	db, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	data, e := os.ReadFile("testdata/python-v1.sql")
	if e == nil {
		_, e = db.Exec(string(data))
	}
	if e != nil {
		db.Close()
		return nil, e
	}
	return db, nil
}
