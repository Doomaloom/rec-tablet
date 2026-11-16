package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"rectablet/src/platform/database"
	"rectablet/src/workflows/swimming/transfer"
	"time"
)

type syncRecord struct {
	SyncOperation
	Revision int  `json:"revision"`
	Deleted  bool `json:"deleted"`
	Success  bool `json:"success"`
}
type attemptPayload struct {
	SkillID     string `json:"skillId"`
	CreatedAt   string `json:"createdAt"`
	SubmittedAt string `json:"submittedAt"`
	Results     []struct {
		SubskillID string `json:"subskillId"`
		Status     string `json:"status"`
	} `json:"results"`
	Timed *struct {
		LapsRequired  int `json:"lapsRequired"`
		LapsCompleted int `json:"lapsCompleted"`
		ElapsedMs     int `json:"elapsedMs"`
	} `json:"timed,omitempty"`
}

// ReadAccepted runs only for the active redemption. Unsent local records and
// draft answers survive the canonical refresh, including corrections/deletions.
func (s *Store) ReadAccepted(ctx context.Context, c *transfer.SyncCredentials) error {
	if c == nil {
		return nil
	}
	var response struct {
		SchemaVersion int          `json:"schemaVersion"`
		Records       []syncRecord `json:"records"`
		Conflicts     []struct {
			Operation SyncOperation `json:"operation"`
			Resolved  bool          `json:"resolved"`
		} `json:"conflicts"`
	}
	if e := syncRequest(ctx, c.Server, "/api/device-results", c.ReadToken, nil, &response); e != nil {
		return e
	}
	if response.SchemaVersion != 1 {
		return errors.New("unsupported results schema")
	}
	return database.Transaction(s.DB, func(tx *sql.Tx) error {
		for _, v := range response.Conflicts {
			if v.Resolved {
				if _, e := tx.Exec("DELETE FROM sync_review WHERE operation_id=?", v.Operation.OperationID); e != nil {
					return e
				}
			}
		}
		for _, r := range response.Records {
			if e := s.applySyncRecord(tx, c, r); e != nil {
				return e
			}
		}
		// Normalize state from the exact database representation without producing
		// echo operations. Tombstone state remains present for future local edits.
		g, e := grant(tx)
		if e != nil {
			return e
		}
		current, e := s.syncCurrent(tx, g)
		if e != nil {
			return e
		}
		for id, o := range current {
			dirty, e := one(tx, "SELECT COUNT(*) count FROM sync_outbox WHERE record_id=?", id)
			if e != nil {
				return e
			}
			if dirty.Int("count") == 0 {
				if _, e = tx.Exec("UPDATE sync_state SET body=? WHERE record_id=?", stateBody(o), id); e != nil {
					return e
				}
			}
		}
		return nil
	})
}
func (s *Store) applySyncRecord(tx *sql.Tx, c *transfer.SyncCredentials, r syncRecord) error {
	if r.SessionID != c.SessionID || r.DatasetID != c.DatasetID || r.Revision < 1 || r.RecordID == "" || transfer.Date(r.LessonDate) != nil || (r.Action != "upsert" && r.Action != "delete") {
		return errors.New("invalid result scope")
	}
	m, e := one(tx, `SELECT m.local_id,e.course_id,c.curriculum_version FROM sync_mapping m JOIN enrollments e ON e.id=m.local_id JOIN courses c ON c.id=e.course_id WHERE m.remote_id=? AND m.class_id=?`, r.EnrollmentID, r.ClassID)
	if e != nil {
		return errors.New("result identity outside local roster")
	}
	lesson, e := one(tx, "SELECT id FROM lessons WHERE course_id=? AND date=?", m.String("course_id"), r.LessonDate)
	if e != nil {
		return errors.New("result lesson outside local roster")
	}
	state, e := one(tx, "SELECT revision FROM sync_state WHERE record_id=?", r.RecordID)
	if e != nil && !errors.Is(e, ErrNotFound) {
		return e
	}
	if e == nil && state.Int("revision") > r.Revision {
		return nil
	}
	dirty, e := one(tx, "SELECT COUNT(*) count FROM sync_outbox WHERE record_id=?", r.RecordID)
	if e != nil {
		return e
	}
	if dirty.Int("count") > 0 {
		return nil
	}
	if r.CurriculumVersion != m.String("curriculum_version") {
		return errors.New("result curriculum outside local course")
	}
	catalog, err := s.catalog(tx, m.String("curriculum_version"))
	if err != nil {
		return errors.New("result curriculum unavailable")
	}
	deleted := r.Deleted || r.Action == "delete"
	switch r.Kind {
	case "attendance":
		if r.RecordID != "attendance:"+r.EnrollmentID+":"+r.LessonDate {
			return errors.New("invalid attendance identity")
		}
		if deleted {
			_, e = tx.Exec("DELETE FROM attendance WHERE lesson_id=? AND enrollment_id=?", lesson.String("id"), m.String("local_id"))
		} else {
			var p struct {
				Status string `json:"status"`
			}
			if json.Unmarshal(r.Payload, &p) != nil || (p.Status != "present" && p.Status != "absent" && p.Status != "unmarked") {
				return errors.New("invalid attendance")
			}
			_, e = tx.Exec("INSERT OR REPLACE INTO attendance VALUES(?,?,?,?)", lesson.String("id"), m.String("local_id"), p.Status, Now())
		}
		if e != nil {
			return e
		}
	case "attempt":
		existing, e := one(tx, "SELECT * FROM attempts WHERE id=?", r.RecordID)
		if e != nil && !errors.Is(e, ErrNotFound) {
			return e
		}
		if e == nil && (existing.String("enrollment_id") != m.String("local_id") || existing.String("lesson_id") != lesson.String("id")) {
			return errors.New("attempt identity changed")
		}
		draft, e := one(tx, "SELECT COUNT(*) count FROM drafts WHERE attempt_id=?", r.RecordID)
		if e != nil {
			return e
		}
		if draft.Int("count") > 0 {
			return nil
		}
		if deleted {
			for _, table := range []string{"results", "revisions", "timed_attempts", "attempts"} {
				key := "attempt_id"
				if table == "attempts" {
					key = "id"
				}
				if _, e = tx.Exec("DELETE FROM "+table+" WHERE "+key+"=?", r.RecordID); e != nil {
					return e
				}
			}
		} else {
			var p attemptPayload
			if json.Unmarshal(r.Payload, &p) != nil {
				return errors.New("invalid attempt payload")
			}
			if _, e = time.Parse(time.RFC3339Nano, p.CreatedAt); e != nil {
				return e
			}
			if _, e = time.Parse(time.RFC3339Nano, p.SubmittedAt); e != nil {
				return e
			}
			skill, e := catalog.Skill(p.SkillID)
			if e != nil {
				return e
			}
			valid := map[string]bool{}
			for _, v := range skill.Subskills {
				valid[v.ID] = true
			}
			seen := map[string]bool{}
			for _, v := range p.Results {
				if !valid[v.SubskillID] || seen[v.SubskillID] || (v.Status != "successful" && v.Status != "not_yet") {
					return errors.New("invalid submitted criterion")
				}
				seen[v.SubskillID] = true
			}
			if len(seen) != len(valid) {
				return errors.New("incomplete submitted criteria")
			}
			if _, e = tx.Exec(`INSERT INTO attempts VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET skill_id=excluded.skill_id,curriculum_version=excluded.curriculum_version,created_at=excluded.created_at,submitted_at=excluded.submitted_at,success=excluded.success`, r.RecordID, m.String("local_id"), lesson.String("id"), p.SkillID, r.CurriculumVersion, p.CreatedAt, p.SubmittedAt, r.Success); e != nil {
				return e
			}
			if _, e = tx.Exec("DELETE FROM results WHERE attempt_id=?", r.RecordID); e != nil {
				return e
			}
			for _, v := range p.Results {
				if _, e = tx.Exec("INSERT INTO results VALUES(?,?,?)", r.RecordID, v.SubskillID, v.Status); e != nil {
					return e
				}
			}
			if _, e = tx.Exec("DELETE FROM timed_attempts WHERE attempt_id=?", r.RecordID); e != nil {
				return e
			}
			if v := p.Timed; v != nil {
				if v.LapsRequired < 1 || v.LapsCompleted < 0 || v.ElapsedMs < 0 {
					return errors.New("invalid timed result")
				}
				if _, e = tx.Exec("INSERT INTO timed_attempts VALUES(?,?,?,?)", r.RecordID, v.LapsRequired, v.LapsCompleted, v.ElapsedMs); e != nil {
					return e
				}
			}
		}
	default:
		return errors.New("invalid result kind")
	}
	o := r.SyncOperation
	if deleted {
		o.Action = "delete"
		o.Payload = json.RawMessage(`{}`)
	}
	_, e = tx.Exec("INSERT INTO sync_state VALUES(?,?,?) ON CONFLICT(record_id) DO UPDATE SET body=excluded.body,revision=excluded.revision", r.RecordID, stateBody(o), r.Revision)
	return e
}
