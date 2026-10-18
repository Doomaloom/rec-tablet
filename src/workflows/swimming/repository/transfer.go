package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"rectablet/src/platform/files"
	"rectablet/src/workflows/swimming/transfer"

	"time"
	"unicode/utf8"
)

func (s *Store) ImportFile(path string) (int, error) {
	p, e := transfer.ReadPackage(path)
	if e != nil {
		return 0, e
	}
	return s.Import(p)
}
func (s *Store) Import(p transfer.ClassPackage) (int, error) {
	if p.SchemaVersion != 1 {
		return 0, errors.New("expected a class package with schemaVersion 1")
	}
	if p.Kind != "rec-tablet-classes" {
		return 0, errors.New("choose a rec-tablet-classes JSON package, not an export or raw roster")
	}
	var e error
	if p.DatasetID, e = transfer.Text(p.DatasetID, "datasetId"); e != nil {
		return 0, e
	}
	if p.Instructor, e = transfer.Text(p.Instructor, "instructor"); e != nil {
		return 0, e
	}
	if len(p.Courses) < 1 || len(p.Courses) > 1000 {
		return 0, errors.New("courses must contain between 1 and 1000 courses")
	}
	// Copy before normalizing so the caller's package stays unchanged.
	copyBytes, _ := json.Marshal(p)
	var cloned transfer.ClassPackage
	_ = json.Unmarshal(copyBytes, &cloned)
	p = cloned
	seenCourses, seenNames := map[string]bool{}, map[string]string{}
	for i := range p.Courses {
		c := &p.Courses[i]
		if c.SourceID, e = transfer.Text(c.SourceID, "sourceId"); e != nil {
			return 0, e
		}
		if seenCourses[c.SourceID] {
			return 0, fmt.Errorf("duplicate course sourceId: %s", c.SourceID)
		}
		seenCourses[c.SourceID] = true
		if c.Name, e = transfer.Text(c.Name, "name"); e != nil {
			return 0, e
		}
		if c.Level, e = transfer.Text(c.Level, "level"); e != nil {
			return 0, e
		}
		c.Level = s.Catalog.NormalizeLevel(c.Level)
		if c.StartTime, e = transfer.Text(c.StartTime, "startTime"); e != nil {
			return 0, e
		}
		t, err := time.Parse("15:04", c.StartTime)
		if err != nil || t.Format("15:04") != c.StartTime {
			return 0, errors.New("startTime must use 24-hour HH:MM")
		}
		if utf8.RuneCountInString(c.Location) > 1000 {
			return 0, errors.New("location must be text (maximum 1000 characters)")
		}
		if len(c.LessonDates) < 1 || len(c.LessonDates) > 366 {
			return 0, errors.New("each course needs 1–366 lessonDates")
		}
		seenDates := map[string]bool{}
		for _, d := range c.LessonDates {
			if e = transfer.Date(d); e != nil {
				return 0, e
			}
			if seenDates[d] {
				return 0, errors.New("duplicate lesson date in a course")
			}
			seenDates[d] = true
		}
		if c.Swimmers == nil || len(c.Swimmers) > 500 {
			return 0, errors.New("swimmers must be a list with at most 500 entries")
		}
		seenIDs := map[string]bool{}
		for j := range c.Swimmers {
			pupil := &c.Swimmers[j]
			if pupil.SourceID, e = transfer.Text(pupil.SourceID, "sourceId"); e != nil {
				return 0, e
			}
			if pupil.Name, e = transfer.Text(pupil.Name, "name"); e != nil {
				return 0, e
			}
			level := c.Level
			if pupil.Level != nil {
				if strings.TrimSpace(*pupil.Level) == "" && c.Level == "SplashPrivate" {
					level = ""
				} else {
					if level, e = transfer.Text(*pupil.Level, "level"); e != nil {
						return 0, e
					}
					level = s.Catalog.NormalizeLevel(level)
				}
			} else if c.Level == "SplashPrivate" {
				level = ""
			}
			pupil.Level = &level
			old, ok := seenNames[pupil.SourceID]
			if seenIDs[pupil.SourceID] || (ok && old != pupil.Name) {
				return 0, fmt.Errorf("ambiguous swimmer sourceId %s; use unique durable identifiers", pupil.SourceID)
			}
			if level == "SplashPrivate" {
				if c.Level == "SplashPrivate" {
					level = ""
					pupil.Level = &level
				} else {
					return 0, fmt.Errorf("assign %s an actual swimmer level; SplashPrivate is a multi-level reference", pupil.Name)
				}
			}
			seenIDs[pupil.SourceID] = true
			seenNames[pupil.SourceID] = pupil.Name
		}
	}
	e = s.transaction(func(q queryer) error {
		dataset, e := meta(q, "dataset_id")
		if e != nil {
			return e
		}
		if dataset != "" && dataset != p.DatasetID {
			return errors.New("this device already has another dataset; use the original datasetId or a separate database")
		}
		for id, name := range seenNames {
			r, e := one(q, "SELECT name FROM swimmers WHERE dataset_id=? AND source_id=?", p.DatasetID, id)
			if e != nil && !errors.Is(e, ErrNotFound) {
				return e
			}
			if e == nil && r.String("name") != name {
				return fmt.Errorf("source ID %s now names a different swimmer; resolve the identity before importing", id)
			}
		}
		for key, value := range map[string]string{"dataset_id": p.DatasetID, "instructor": p.Instructor} {
			if _, e = q.Exec("INSERT OR REPLACE INTO meta VALUES (?,?)", key, value); e != nil {
				return e
			}
		}
		for _, c := range p.Courses {
			if _, e = q.Exec(`INSERT INTO courses VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(dataset_id,source_id) DO UPDATE SET name=excluded.name,level=excluded.level,time=excluded.time,location=excluded.location`, ID(), p.DatasetID, c.SourceID, c.Name, c.Level, c.StartTime, c.Location, s.Catalog.Version); e != nil {
				return e
			}
			course, e := one(q, "SELECT * FROM courses WHERE dataset_id=? AND source_id=?", p.DatasetID, c.SourceID)
			if e != nil {
				return e
			}
			if _, e = q.Exec("UPDATE enrollments SET active=0 WHERE course_id=?", course.String("id")); e != nil {
				return e
			}
			for _, pupil := range c.Swimmers {
				if _, e = q.Exec("INSERT OR IGNORE INTO swimmers VALUES (?,?,?,?)", ID(), p.DatasetID, pupil.SourceID, pupil.Name); e != nil {
					return e
				}
				person, e := one(q, "SELECT id FROM swimmers WHERE dataset_id=? AND source_id=?", p.DatasetID, pupil.SourceID)
				if e != nil {
					return e
				}
				if _, e = q.Exec(`INSERT INTO enrollments VALUES (?,?,?,?,1) ON CONFLICT(course_id,swimmer_id) DO UPDATE SET level=excluded.level,active=1`, ID(), course.String("id"), person.String("id"), *pupil.Level); e != nil {
					return e
				}
			}
			for _, d := range c.LessonDates {
				if _, e = q.Exec("INSERT OR IGNORE INTO lessons VALUES (?,?,?)", ID(), course.String("id"), d); e != nil {
					return e
				}
			}
		}
		_, e = q.Exec("INSERT OR REPLACE INTO meta VALUES ('last_import',?)", Now())
		return e
	})
	if e != nil {
		return 0, e
	}
	return len(p.Courses), nil
}

func (s *Store) Snapshot() (transfer.Snapshot, error) {
	snap := transfer.Snapshot{SchemaVersion: 1, Kind: "rec-tablet-results", ExportedAt: Now(), DatabaseVersion: 1, Tables: map[string][]map[string]any{}}
	e := s.transaction(func(q queryer) error {
		for _, table := range []string{"meta", "curricula", "courses", "swimmers", "enrollments", "lessons", "attendance", "attempts", "results", "drafts", "revisions"} {
			rs, e := rows(q, "SELECT * FROM "+table)
			if e != nil {
				return e
			}
			snap.Tables[table] = make([]map[string]any, 0, len(rs))
			for _, r := range rs {
				snap.Tables[table] = append(snap.Tables[table], map[string]any(r))
			}
		}
		// Older v1 snapshots do not contain timed_attempts. Keep that shape when
		// there is no timed data, while exporting the table once it has rows.
		timed, e := rows(q, "SELECT * FROM timed_attempts")
		if e != nil {
			return e
		}
		if len(timed) > 0 {
			snap.Tables["timed_attempts"] = make([]map[string]any, 0, len(timed))
			for _, r := range timed {
				snap.Tables["timed_attempts"] = append(snap.Tables["timed_attempts"], map[string]any(r))
			}
		}
		return nil
	})
	return snap, e
}
func (s *Store) Export(path string) error {
	for _, protected := range []string{s.Path, s.Path + "-wal", s.Path + "-shm", s.Path + ".lock", s.Path + ".go.lock"} {
		if files.SamePath(path, protected) {
			return errors.New("choose a JSON file separate from the app database")
		}
	}
	snap, e := s.Snapshot()
	if e != nil {
		return e
	}
	return files.WriteJSON(path, snap)
}
