// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"database/sql"
)

type record map[string]any

func (r record) String(key string) string { s, _ := r[key].(string); return s }
func (r record) Int(key string) int {
	switch v := r[key].(type) {
	case int64:
		return int(v)
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

type queryer interface {
	Query(string, ...any) (*sql.Rows, error)
	Exec(string, ...any) (sql.Result, error)
}

func rows(q queryer, query string, args ...any) ([]record, error) {
	rs, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	names, err := rs.Columns()
	if err != nil {
		return nil, err
	}
	out := []record{}
	for rs.Next() {
		values := make([]any, len(names))
		ptrs := make([]any, len(names))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err = rs.Scan(ptrs...); err != nil {
			return nil, err
		}
		r := record{}
		for i, n := range names {
			if b, ok := values[i].([]byte); ok {
				r[n] = string(b)
			} else {
				r[n] = values[i]
			}
		}
		out = append(out, r)
	}
	return out, rs.Err()
}
func one(q queryer, query string, args ...any) (record, error) {
	rs, e := rows(q, query, args...)
	if e != nil {
		return nil, e
	}
	if len(rs) == 0 {
		return nil, ErrNotFound
	}
	return rs[0], nil
}
func (s *Store) queryRows(query string, args ...any) ([]record, error) {
	return rows(s.DB, query, args...)
}
func (s *Store) queryOne(query string, args ...any) (record, error) { return one(s.DB, query, args...) }
