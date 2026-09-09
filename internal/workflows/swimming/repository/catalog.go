// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"encoding/json"
	"errors"

	"rectablet/internal/workflows/swimming/curriculum"
)

func meta(q queryer, key string) (string, error) {
	r, e := one(q, "SELECT value FROM meta WHERE key=?", key)
	if errors.Is(e, ErrNotFound) {
		return "", nil
	}
	return r.String("value"), e
}
func (s *Store) Meta(key string) (string, error) { return meta(s.DB, key) }
func (s *Store) catalog(q queryer, version string) (curriculum.Catalog, error) {
	if c, ok := s.catalogs.Load(version); ok {
		return c.(curriculum.Catalog), nil
	}
	r, e := one(q, "SELECT body FROM curricula WHERE version=?", version)
	if e != nil {
		return curriculum.Catalog{}, e
	}
	var c curriculum.Catalog
	e = json.Unmarshal([]byte(r.String("body")), &c)
	if e == nil {
		s.catalogs.Store(version, c)
	}
	return c, e
}
func (s *Store) CatalogFor(version string) (curriculum.Catalog, error) {
	return s.catalog(s.DB, version)
}
func (s *Store) skill(q queryer, version, id string) (curriculum.Skill, error) {
	c, e := s.catalog(q, version)
	if e != nil {
		return curriculum.Skill{}, e
	}
	return c.Skill(id)
}
func (s *Store) Skill(version, id string) (curriculum.Skill, error) {
	return s.skill(s.DB, version, id)
}
