// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"rectablet/src/platform/database"

	"sync"
	"time"

	"rectablet/src/workflows/swimming/curriculum"
)

//go:embed schema.sql
var schema string
var ErrNotFound = errors.New("this record is no longer available; return to My Classes")

type Store struct {
	DB        *sql.DB
	Path      string
	Catalog   curriculum.Catalog
	catalogs  sync.Map
	syncReady bool
}

func Now() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000000+00:00") }
func ID() string {
	b := [16]byte{}
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
func Open(path string) (*Store, error) { return OpenWithCatalog(path, curriculum.Bundled()) }
func OpenWithCatalog(path string, c curriculum.Catalog) (*Store, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	db, err := database.Open(path)
	if err != nil {
		return nil, err
	}
	s := &Store{DB: db, Path: path, Catalog: c}
	fail := func(e error) (*Store, error) { db.Close(); return nil, e }
	var version int
	if err = db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fail(err)
	}
	if version > 2 {
		return fail(errors.New("this database was created by a newer app; open it with that version"))
	}
	if version == 0 {
		if err = s.transaction(func(q queryer) error { _, e := q.Exec(schema); return e }); err != nil {
			return fail(err)
		}
	} else {
		// Keep the v1 database contract while adding the timed-assessment table
		// for databases created by earlier builds.
		if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS timed_attempts (
		 attempt_id TEXT PRIMARY KEY REFERENCES attempts(id),
		 laps_required INTEGER NOT NULL CHECK(laps_required > 0),
		 laps_completed INTEGER NOT NULL DEFAULT 0 CHECK(laps_completed >= 0),
		 elapsed_ms INTEGER NOT NULL DEFAULT 0 CHECK(elapsed_ms >= 0))`); err != nil {
			return fail(err)
		}
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS lesson_plans (
 lesson_id TEXT PRIMARY KEY REFERENCES lessons(id), body TEXT NOT NULL)`); err != nil {
		return fail(err)
	}
	var hadScope int
	if err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sync_scope'").Scan(&hadScope); err != nil {
		return fail(err)
	}
	if err = s.transaction(func(q queryer) error {
		if _, e := q.Exec(syncSchema); e != nil {
			return e
		}
		if hadScope == 0 {
			// Existing bound installations retain their roster until the next redemption.
			_, e := q.Exec("INSERT OR IGNORE INTO sync_scope SELECT DISTINCT e.course_id FROM enrollments e JOIN sync_mapping m ON m.local_id=e.id")
			return e
		}
		return nil
	}); err != nil {
		return fail(err)
	}
	s.syncReady = true
	body, err := json.Marshal(c)
	if err != nil {
		return fail(err)
	}
	// Refresh wording edits within the same version while keeping assessment IDs
	// and their saved results intact. Structural changes still use a new version.
	if _, err = db.Exec("INSERT INTO curricula VALUES (?,?) ON CONFLICT(version) DO UPDATE SET body=excluded.body", c.Version, string(body)); err != nil {
		return fail(err)
	}
	if err = s.resetMarkingDataForCatalogChange(c.Version); err != nil {
		return fail(err)
	}
	return s, nil
}
func (s *Store) Close() error { return s.DB.Close() }
func (s *Store) transaction(fn func(queryer) error) error {
	return database.Transaction(s.DB, func(tx *sql.Tx) error {
		if err := fn(tx); err != nil {
			return err
		}
		if s.syncReady {
			return s.captureSync(tx)
		}
		return nil
	})
}

// resetMarkingDataForCatalogChange handles the intentionally non-production
// curriculum reset used when the bundled catalog changes shape. Courses keep
// their roster and attendance data, while assessment rows are discarded and
// all courses are pointed at the newly bundled curriculum version.
func (s *Store) resetMarkingDataForCatalogChange(version string) error {
	// Bound history and drafts remain authoritative local data across upgrades.
	// Never turn a curriculum update into server deletion operations.
	if s.syncReady {
		if _, err := grant(s.DB); err == nil {
			return nil
		} else if !errors.Is(err, ErrNotFound) {
			return err
		}
	}

	return s.transaction(func(q queryer) error {
		oldCourses, err := one(q, "SELECT COUNT(*) AS count FROM courses WHERE curriculum_version<>?", version)
		if err != nil {
			return err
		}
		oldAttempts, err := one(q, "SELECT COUNT(*) AS count FROM attempts WHERE curriculum_version<>?", version)
		if err != nil {
			return err
		}
		if oldCourses.Int("count") == 0 && oldAttempts.Int("count") == 0 {
			return nil
		}
		for _, table := range []string{"revisions", "results", "drafts", "timed_attempts", "attempts"} {
			if _, err = q.Exec("DELETE FROM " + table); err != nil {
				return err
			}
		}
		_, err = q.Exec("UPDATE courses SET curriculum_version=?", version)
		return err
	})
}
