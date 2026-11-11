package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rectablet/src/platform/database"
	"rectablet/src/workflows/swimming/transfer"
	"sort"
	"strings"
	"sync"
)

// These tables are intentionally absent from v1 manual snapshots: capabilities,
// transport history, and canonical bindings belong only to this installation.
const syncSchema = `
CREATE TABLE IF NOT EXISTS sync_config(id INTEGER PRIMARY KEY CHECK(id=1),body TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS sync_mapping(local_id TEXT PRIMARY KEY,class_id TEXT NOT NULL,remote_id TEXT UNIQUE NOT NULL);
CREATE TABLE IF NOT EXISTS sync_state(record_id TEXT PRIMARY KEY,body TEXT NOT NULL,revision INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS sync_outbox(seq INTEGER PRIMARY KEY AUTOINCREMENT,operation_id TEXT UNIQUE NOT NULL,record_id TEXT NOT NULL,body TEXT NOT NULL,frozen INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS sync_review(operation_id TEXT PRIMARY KEY,record_id TEXT NOT NULL,status TEXT NOT NULL,body TEXT NOT NULL);
PRAGMA user_version=2;`

type SyncOperation struct {
	OperationID       string          `json:"operationId"`
	SessionID         string          `json:"sessionId"`
	DatasetID         string          `json:"datasetId"`
	ClassID           string          `json:"classId"`
	EnrollmentID      string          `json:"enrollmentId"`
	RecordID          string          `json:"recordId"`
	Kind              string          `json:"kind"`
	Action            string          `json:"action"`
	BaseRevision      int             `json:"baseRevision"`
	LessonDate        string          `json:"lessonDate"`
	CurriculumVersion string          `json:"curriculumVersion"`
	Payload           json.RawMessage `json:"payload"`
}
type uploadGrant struct{ Server, UploadToken, SessionID, DatasetID, InstructorID, ExpiresAt string }

var syncNetwork sync.Mutex

// ConfigureSync never persists the read token. Identity bindings are immutable.
func (s *Store) ConfigureSync(p transfer.ClassPackage) error {
	c := p.Credentials
	if c == nil {
		return nil
	}
	identity := p.Sync
	if identity == nil || identity.SchemaVersion != 1 || c.SchemaVersion != 1 || identity.DatasetID != p.DatasetID || c.DatasetID != p.DatasetID || c.SessionID != identity.SessionID || c.InstructorID != identity.InstructorID || c.UploadToken == "" || c.Server == "" {
		return errors.New("invalid sync identity")
	}
	return database.Transaction(s.DB, func(tx *sql.Tx) error {
		old, e := one(tx, "SELECT body FROM sync_config WHERE id=1")
		if e != nil && !errors.Is(e, ErrNotFound) {
			return e
		}
		if e == nil {
			var g uploadGrant
			if json.Unmarshal([]byte(old.String("body")), &g) != nil || g.SessionID != c.SessionID || g.DatasetID != c.DatasetID || g.InstructorID != c.InstructorID {
				return errors.New("sync identity cannot change")
			}
		}
		if len(identity.Courses) != len(p.Courses) {
			return errors.New("incomplete canonical roster")
		}
		courses := map[string]transfer.ClassInput{}
		for _, v := range p.Courses {
			courses[v.SourceID] = v
		}
		seen := map[string]bool{}
		for _, course := range identity.Courses {
			source, ok := courses[course.SourceID]
			if !ok || seen[course.SourceID] || course.ClassID == "" || len(source.Swimmers) != len(course.Swimmers) {
				return errors.New("canonical course mismatch")
			}
			seen[course.SourceID] = true
			people := map[string]bool{}
			for _, v := range source.Swimmers {
				people[v.SourceID] = true
			}
			for _, sw := range course.Swimmers {
				if !people[sw.SourceID] || sw.EnrollmentID == "" {
					return errors.New("canonical swimmer mismatch")
				}
				delete(people, sw.SourceID)
				local, e := one(tx, `SELECT e.id FROM enrollments e JOIN courses c ON c.id=e.course_id JOIN swimmers w ON w.id=e.swimmer_id WHERE c.dataset_id=? AND c.source_id=? AND w.source_id=?`, p.DatasetID, course.SourceID, sw.SourceID)
				if e != nil {
					return e
				}
				prior, e := one(tx, "SELECT * FROM sync_mapping WHERE local_id=?", local.String("id"))
				if e != nil && !errors.Is(e, ErrNotFound) {
					return e
				}
				if e == nil && (prior.String("class_id") != course.ClassID || prior.String("remote_id") != sw.EnrollmentID) {
					return errors.New("canonical identity cannot change")
				}
				if _, e = tx.Exec("INSERT OR IGNORE INTO sync_mapping VALUES(?,?,?)", local.String("id"), course.ClassID, sw.EnrollmentID); e != nil {
					return e
				}
			}
		}
		b, _ := json.Marshal(uploadGrant{c.Server, c.UploadToken, c.SessionID, c.DatasetID, c.InstructorID, c.ExpiresAt})
		if _, e = tx.Exec("INSERT OR REPLACE INTO sync_config VALUES(1,?)", string(b)); e != nil {
			return e
		}
		return s.captureSync(tx)
	})
}
func grant(q queryer) (uploadGrant, error) {
	var g uploadGrant
	r, e := one(q, "SELECT body FROM sync_config WHERE id=1")
	if e != nil {
		return g, e
	}
	e = json.Unmarshal([]byte(r.String("body")), &g)
	return g, e
}
func stateBody(o SyncOperation) string {
	o.OperationID = ""
	o.BaseRevision = 0
	var payload any
	if json.Unmarshal(o.Payload, &payload) == nil {
		o.Payload, _ = json.Marshal(payload)
	}
	b, _ := json.Marshal(o)
	return string(b)
}

// The state comparison runs inside the domain transaction. Thus auto absence,
// corrections, undo and bulk resets cannot commit without their outbound edits.
func (s *Store) captureSync(q queryer) error {
	g, e := grant(q)
	if errors.Is(e, ErrNotFound) {
		return nil
	}
	if e != nil {
		return e
	}
	current, e := s.syncCurrent(q, g)
	if e != nil {
		return e
	}
	previous, e := rows(q, "SELECT * FROM sync_state")
	if e != nil {
		return e
	}
	old := map[string]record{}
	for _, r := range previous {
		id := r.String("record_id")
		old[id] = r
		if _, ok := current[id]; !ok {
			var o SyncOperation
			if e = json.Unmarshal([]byte(r.String("body")), &o); e != nil {
				return e
			}
			if o.Action != "delete" {
				o.Action = "delete"
				o.Payload = json.RawMessage(`{}`)
				current[id] = o
			}
		}
	}
	ids := []string{}
	for id := range current {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		o := current[id]
		body := stateBody(o)
		prior := old[id]
		if prior.String("body") == body {
			continue
		}
		o.OperationID = ID()
		o.BaseRevision = prior.Int("revision")
		b, _ := json.Marshal(o)
		if _, e = q.Exec("INSERT INTO sync_outbox(operation_id,record_id,body) VALUES(?,?,?)", o.OperationID, id, string(b)); e != nil {
			return e
		}
		if _, e = q.Exec("INSERT INTO sync_state VALUES(?,?,?) ON CONFLICT(record_id) DO UPDATE SET body=excluded.body", id, body, prior.Int("revision")); e != nil {
			return e
		}
	}
	return nil
}
func (s *Store) syncCurrent(q queryer, g uploadGrant) (map[string]SyncOperation, error) {
	out := map[string]SyncOperation{}
	attendance, e := rows(q, `SELECT a.status,l.date,c.curriculum_version,m.class_id,m.remote_id FROM attendance a JOIN lessons l ON l.id=a.lesson_id JOIN courses c ON c.id=l.course_id JOIN sync_mapping m ON m.local_id=a.enrollment_id`)
	if e != nil {
		return nil, e
	}
	for _, r := range attendance {
		p, _ := json.Marshal(map[string]string{"status": r.String("status")})
		o := SyncOperation{SessionID: g.SessionID, DatasetID: g.DatasetID, ClassID: r.String("class_id"), EnrollmentID: r.String("remote_id"), RecordID: "attendance:" + r.String("remote_id") + ":" + r.String("date"), Kind: "attendance", Action: "upsert", LessonDate: r.String("date"), CurriculumVersion: r.String("curriculum_version"), Payload: p}
		out[o.RecordID] = o
	}
	attempts, e := rows(q, `SELECT a.*,l.date,m.class_id,m.remote_id FROM attempts a JOIN lessons l ON l.id=a.lesson_id JOIN sync_mapping m ON m.local_id=a.enrollment_id WHERE a.submitted_at IS NOT NULL`)
	if e != nil {
		return nil, e
	}
	for _, r := range attempts {
		criteria, e := rows(q, "SELECT subskill_id,status FROM results WHERE attempt_id=? ORDER BY subskill_id", r.String("id"))
		if e != nil {
			return nil, e
		}
		results := []map[string]string{}
		for _, v := range criteria {
			results = append(results, map[string]string{"subskillId": v.String("subskill_id"), "status": v.String("status")})
		}
		p := map[string]any{"skillId": r.String("skill_id"), "createdAt": r.String("created_at"), "submittedAt": r.String("submitted_at"), "results": results}
		timed, e := one(q, "SELECT * FROM timed_attempts WHERE attempt_id=?", r.String("id"))
		if e == nil {
			p["timed"] = map[string]int{"lapsRequired": timed.Int("laps_required"), "lapsCompleted": timed.Int("laps_completed"), "elapsedMs": timed.Int("elapsed_ms")}
		} else if !errors.Is(e, ErrNotFound) {
			return nil, e
		}
		body, _ := json.Marshal(p)
		o := SyncOperation{SessionID: g.SessionID, DatasetID: g.DatasetID, ClassID: r.String("class_id"), EnrollmentID: r.String("remote_id"), RecordID: r.String("id"), Kind: "attempt", Action: "upsert", LessonDate: r.String("date"), CurriculumVersion: r.String("curriculum_version"), Payload: body}
		out[o.RecordID] = o
	}
	return out, nil
}
func syncRequest(ctx context.Context, server, path, token string, body, out any) error {
	var reader io.Reader
	method := "GET"
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		reader = bytes.NewReader(b)
		method = "POST"
	}
	req, e := http.NewRequestWithContext(ctx, method, strings.TrimRight(server, "/")+path, reader)
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("sync redirects refused") }}
	res, e := client.Do(req)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("results server returned %d", res.StatusCode)
	}
	data, e := io.ReadAll(io.LimitReader(res.Body, 16<<20+1))
	if e != nil {
		return e
	}
	if len(data) > 16<<20 {
		return errors.New("results response too large")
	}
	return json.Unmarshal(data, out)
}

// UploadPending sends one immutable operation per record at a time. Only the
// next unsent local edit receives a new base revision after a positive receipt.
func (s *Store) UploadPending(ctx context.Context) error {
	syncNetwork.Lock()
	defer syncNetwork.Unlock()
	g, e := grant(s.DB)
	if errors.Is(e, ErrNotFound) {
		return nil
	}
	if e != nil {
		return e
	}
	for ctx.Err() == nil {
		batch := []SyncOperation{}
		e = database.Transaction(s.DB, func(tx *sql.Tx) error {
			pending, e := rows(tx, `SELECT o.* FROM sync_outbox o WHERE o.seq=(SELECT MIN(n.seq) FROM sync_outbox n WHERE n.record_id=o.record_id) AND NOT EXISTS(SELECT 1 FROM sync_review r WHERE r.record_id=o.record_id) ORDER BY seq LIMIT 200`)
			if e != nil {
				return e
			}
			for _, r := range pending {
				var o SyncOperation
				if e = json.Unmarshal([]byte(r.String("body")), &o); e != nil {
					return e
				}
				if r.Int("frozen") == 0 {
					state, e := one(tx, "SELECT revision FROM sync_state WHERE record_id=?", o.RecordID)
					if e != nil {
						return e
					}
					o.BaseRevision = state.Int("revision")
					b, _ := json.Marshal(o)
					if _, e = tx.Exec("UPDATE sync_outbox SET body=?,frozen=1 WHERE operation_id=?", string(b), o.OperationID); e != nil {
						return e
					}
				}
				batch = append(batch, o)
			}
			return nil
		})
		if e != nil {
			return e
		}
		if len(batch) == 0 {
			return nil
		}
		var response struct {
			SchemaVersion int `json:"schemaVersion"`
			Receipts      []struct {
				OperationID string `json:"operationId"`
				Status      string `json:"status"`
				Revision    int    `json:"revision"`
				Reason      string `json:"reason"`
			} `json:"receipts"`
		}
		if e = syncRequest(ctx, g.Server, "/api/device-results/upload", g.UploadToken, map[string]any{"schemaVersion": 1, "operations": batch}, &response); e != nil {
			return e
		}
		if response.SchemaVersion != 1 {
			return errors.New("unsupported sync response")
		}
		progress := false
		e = database.Transaction(s.DB, func(tx *sql.Tx) error {
			for _, r := range response.Receipts {
				var sent *SyncOperation
				for i := range batch {
					if batch[i].OperationID == r.OperationID {
						sent = &batch[i]
						break
					}
				}
				if sent == nil {
					continue
				}
				switch r.Status {
				case "accepted", "already_accepted":
					if r.Revision < 1 {
						return errors.New("invalid receipt revision")
					}
					if _, e := tx.Exec("UPDATE sync_state SET revision=? WHERE record_id=?", r.Revision, sent.RecordID); e != nil {
						return e
					}
				case "pending_review":
					b, _ := json.Marshal(sent)
					if _, e := tx.Exec("INSERT OR REPLACE INTO sync_review VALUES(?,?,?,?)", r.OperationID, sent.RecordID, r.Status, string(b)); e != nil {
						return e
					}
				case "rejected":
					continue
				default:
					return errors.New("unknown receipt status")
				}
				if _, e := tx.Exec("DELETE FROM sync_outbox WHERE operation_id=?", r.OperationID); e != nil {
					return e
				}
				progress = true
			}
			return nil
		})
		if e != nil {
			return e
		}
		if !progress {
			return errors.New("no acknowledged results; queue retained")
		}
	}
	return ctx.Err()
}
func (s *Store) SyncStatus() (pending, review int, err error) {
	r, err := one(s.DB, "SELECT (SELECT COUNT(*) FROM sync_outbox) pending,(SELECT COUNT(*) FROM sync_review) review")
	if err == nil {
		pending = r.Int("pending")
		review = r.Int("review")
	}
	return
}

// SyncInstructorID identifies a previously bound profile without disclosing capabilities.
func (s *Store) SyncInstructorID() (string, error) {
	g, e := grant(s.DB)
	if errors.Is(e, ErrNotFound) {
		return "", nil
	}
	return g.InstructorID, e
}

// ValidateSyncProfile refuses to reuse a name-based legacy database containing
// another canonical scope. It never deletes historical work to make a match.
func (s *Store) ValidateSyncProfile(p transfer.ClassPackage) error {
	if p.Credentials == nil {
		return nil
	}
	id, e := s.SyncInstructorID()
	if e != nil {
		return e
	}
	if id != "" && id != p.Credentials.InstructorID {
		return errors.New("saved profile belongs to another instructor")
	}
	expected := map[string]bool{}
	for _, c := range p.Courses {
		expected[c.SourceID] = true
	}
	existing, e := rows(s.DB, "SELECT source_id FROM courses")
	if e != nil {
		return e
	}
	if id == "" {
		for _, c := range existing {
			if !expected[c.String("source_id")] {
				return errors.New("legacy profile has ambiguous class scope; preserve it and use a separate database")
			}
		}
	}
	return nil
}
