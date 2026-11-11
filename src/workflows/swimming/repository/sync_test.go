package repository

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/transfer"
	"strings"
	"sync"
	"testing"
	"time"
)

func bindFixture(t *testing.T, f *fixture, server string) transfer.ClassPackage {
	t.Helper()
	p := f.p
	p.Sync = &transfer.SyncIdentity{SchemaVersion: 1, SessionID: "session", DatasetID: p.DatasetID, InstructorID: "instructor"}
	for _, c := range p.Courses {
		course := transfer.SyncCourse{SourceID: c.SourceID, ClassID: "class-" + c.SourceID}
		for _, w := range c.Swimmers {
			course.Swimmers = append(course.Swimmers, transfer.SyncSwimmer{SourceID: w.SourceID, EnrollmentID: "enrollment-" + c.SourceID + "-" + w.SourceID})
		}
		p.Sync.Courses = append(p.Sync.Courses, course)
	}
	p.Credentials = &transfer.SyncCredentials{SchemaVersion: 1, SessionID: "session", DatasetID: p.DatasetID, InstructorID: "instructor", UploadToken: "upload-secret", ReadToken: "read-secret", Server: server, ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)}
	check(t, f.s.ConfigureSync(p))
	return p
}
func queue(t *testing.T, s *Store) []SyncOperation {
	t.Helper()
	out := []SyncOperation{}
	for _, r := range must(s.queryRows("SELECT body FROM sync_outbox ORDER BY seq")) {
		var o SyncOperation
		check(t, json.Unmarshal([]byte(r.String("body")), &o))
		out = append(out, o)
	}
	return out
}
func count(t *testing.T, s *Store, table string) int {
	t.Helper()
	return must(s.queryOne("SELECT COUNT(*) count FROM " + table)).Int("count")
}

func TestSyncOutboxAtomicCoverageAndPrivateMigration(t *testing.T) {
	f := setup(t)
	bindFixture(t, f, "http://localhost")
	_, e := f.s.DB.Exec(`CREATE TRIGGER fail_sync BEFORE INSERT ON sync_outbox BEGIN SELECT RAISE(ABORT,'disk failure'); END`)
	check(t, e)
	if f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present") == nil {
		t.Fatal("expected transactional failure")
	}
	if count(t, f.s, "attendance") != 0 {
		t.Fatal("attendance committed without outbox")
	}
	_, e = f.s.DB.Exec("DROP TRIGGER fail_sync")
	check(t, e)
	check(t, f.s.MarkPastAttendanceAbsent(f.course.String("lesson_id"), "2099-01-01"))
	if len(queue(t, f.s)) == 0 {
		t.Fatal("automatic absence missing")
	}
	before := len(queue(t, f.s))
	draft := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	if len(queue(t, f.s)) != before {
		t.Fatal("draft transmitted")
	}
	attempt := f.attempt(t, true, "")
	ops := queue(t, f.s)
	if ops[len(ops)-1].RecordID != attempt || ops[len(ops)-1].Kind != "attempt" {
		t.Fatal("submitted criteria missing")
	}
	check(t, f.s.EditAttempt(attempt))
	check(t, f.s.Mark(attempt, "Splash1:1:1", "not_yet"))
	_ = must(f.s.Submit(attempt))
	if len(queue(t, f.s)) != before+2 {
		t.Fatal("correction missing")
	}
	check(t, f.s.ResetClassProgress(f.course.String("id")))
	ops = queue(t, f.s)
	if ops[len(ops)-1].Action != "delete" || ops[len(ops)-1].RecordID != attempt {
		t.Fatal("reset tombstone missing")
	}
	snap := must(f.s.Snapshot())
	data, _ := json.Marshal(snap)
	if strings.Contains(string(data), "secret") || strings.Contains(string(data), "sync_") {
		t.Fatal("export leaked transport state")
	}
	// Recreate the original v1 schema boundary with an unfinished assessment.
	draft = must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	for _, table := range []string{"sync_config", "sync_mapping", "sync_state", "sync_outbox", "sync_review"} {
		_, e = f.s.DB.Exec("DROP TABLE " + table)
		check(t, e)
	}
	_, e = f.s.DB.Exec("PRAGMA user_version=1")
	check(t, e)
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	if count(t, f.s, "drafts") != 1 {
		t.Fatal("v1 draft lost")
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		info, e := os.Stat(f.path + suffix)
		check(t, e)
		if info.Mode().Perm() != 0600 {
			t.Fatalf("unsafe SQLite %s mode %o", suffix, info.Mode().Perm())
		}
	}
}

func TestSyncInterruptedRetrySequentialRevisionAndReview(t *testing.T) {
	f := setup(t)
	var mu sync.Mutex
	accepted := map[string]string{}
	revisions := map[string]int{}
	seen := []SyncOperation{}
	disconnect := true
	review := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/device-results/upload" || r.Header.Get("Authorization") != "Bearer upload-secret" {
			http.Error(w, "scope", 403)
			return
		}
		var req struct {
			Operations []SyncOperation `json:"operations"`
		}
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			http.Error(w, "JSON", 400)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		receipts := []map[string]any{}
		for _, o := range req.Operations {
			seen = append(seen, o)
			b, _ := json.Marshal(o)
			status := "accepted"
			if previous, ok := accepted[o.OperationID]; ok {
				if previous != string(b) {
					t.Error("immutable operation changed")
				}
				status = "already_accepted"
			} else {
				if o.BaseRevision != revisions[o.RecordID] {
					t.Errorf("base revision %d != %d", o.BaseRevision, revisions[o.RecordID])
				}
				accepted[o.OperationID] = string(b)
				revisions[o.RecordID]++
			}
			if review {
				status = "pending_review"
			}
			receipts = append(receipts, map[string]any{"operationId": o.OperationID, "status": status, "revision": revisions[o.RecordID]})
		}
		if disconnect {
			disconnect = false
			http.Error(w, "interrupted after server commit", 503)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "receipts": receipts})
	}))
	defer server.Close()
	bindFixture(t, f, server.URL)
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	if f.s.UploadPending(context.Background()) == nil {
		t.Fatal("expected lost response")
	}
	first := queue(t, f.s)[0]
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "absent"))
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	check(t, f.s.UploadPending(context.Background()))
	if len(queue(t, f.s)) != 0 {
		t.Fatal("queue did not drain")
	}
	mu.Lock()
	if len(seen) != 3 || seen[0].OperationID != seen[1].OperationID || seen[2].BaseRevision != 1 || seen[1].OperationID != first.OperationID {
		t.Fatalf("retry sequence %+v", seen)
	}
	review = true
	mu.Unlock()
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	check(t, f.s.UploadPending(context.Background()))
	pending, needsReview, e := f.s.SyncStatus()
	check(t, e)
	if pending != 0 || needsReview != 1 {
		t.Fatal("review mixed with upload queue")
	}
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "absent"))
	mu.Lock()
	sent := len(seen)
	mu.Unlock()
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	check(t, f.s.UploadPending(context.Background()))
	mu.Lock()
	if len(seen) != sent {
		t.Fatal("sent edit before review resolution")
	}
	mu.Unlock()
}

func TestSyncPullCorrectionsDeletesDraftsAndNoEcho(t *testing.T) {
	f := setup(t)
	records := []syncRecord{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/device-results" || r.Header.Get("Authorization") != "Bearer read-secret" {
			http.Error(w, "no", 403)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "records": records})
	}))
	defer server.Close()
	p := bindFixture(t, f, server.URL)
	submitted := f.attempt(t, true, "")
	draft := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	op := queue(t, f.s)[0]
	if op.RecordID != submitted {
		t.Fatal(op)
	}
	// Represent a previously acknowledged local submit before a website correction.
	_, e := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, e)
	var payload attemptPayload
	check(t, json.Unmarshal(op.Payload, &payload))
	payload.Results[0].Status = "not_yet"
	op.Payload, _ = json.Marshal(payload)
	records = []syncRecord{{SyncOperation: op, Revision: 2, Success: false}}
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if f.progress().Successes != 0 {
		t.Fatal("correction not applied")
	}
	if count(t, f.s, "drafts") != 1 {
		t.Fatal("draft overwritten")
	}
	check(t, f.s.Mark(draft, "Splash1:2:1", "not_yet"))
	if len(queue(t, f.s)) != 0 {
		t.Fatal("incoming result echoed")
	}
	records[0].Deleted = true
	records[0].Action = "delete"
	records[0].Payload = json.RawMessage(`{}`)
	records[0].Revision = 3
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if f.progress().Assessed != 0 || count(t, f.s, "drafts") != 1 {
		t.Fatal("tombstone or draft handling")
	}
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	if len(queue(t, f.s)) != 0 {
		t.Fatal("tombstone echoed")
	}
	// Foreign records abort the whole inbound transaction.
	records[0].EnrollmentID = "foreign"
	if f.s.ReadAccepted(context.Background(), p.Credentials) == nil {
		t.Fatal("accepted foreign enrollment")
	}
}

func TestSyncTimedAttemptsAndUndo(t *testing.T) {
	f := setup(t)
	bindFixture(t, f, "http://localhost")
	heat := must(f.s.NewTimedHeat([]string{f.person.String("id")}, f.course.String("lesson_id"), "Splash1:1", 2))
	id := heat[0].AttemptID
	if len(queue(t, f.s)) != 0 {
		t.Fatal("unfinished timed heat transmitted")
	}
	check(t, f.s.SaveTimedHeat([]domain.TimedResult{{AttemptID: id, LapsRequired: 2, LapsCompleted: 2, ElapsedMS: 1000, Success: true}}))
	ops := queue(t, f.s)
	var payload attemptPayload
	check(t, json.Unmarshal(ops[0].Payload, &payload))
	if payload.Timed == nil || payload.Timed.ElapsedMs != 1000 {
		t.Fatal("timing missing")
	}
	check(t, f.s.RemoveLatestSuccess(f.person.String("id"), "Splash1:1", f.version))
	ops = queue(t, f.s)
	if ops[len(ops)-1].Action != "delete" {
		t.Fatal("undo tombstone missing")
	}
}

func TestSyncConfigurationRollbackAndCredentialIsolation(t *testing.T) {
	f := setup(t)
	p := bindFixture(t, f, "http://localhost")
	p.Sync.Courses[0].Swimmers[0].EnrollmentID = "different"
	if f.s.ConfigureSync(p) == nil {
		t.Fatal("canonical rebinding accepted")
	}
	raw := must(f.s.queryOne("SELECT body FROM sync_config")).String("body")
	if strings.Contains(raw, "read-secret") {
		t.Fatal("read capability persisted")
	}
	if f.s.transaction(func(q queryer) error {
		_, e := q.Exec("INSERT OR REPLACE INTO attendance VALUES(?,?,?,?)", f.course.String("lesson_id"), f.person.String("id"), "present", Now())
		if e != nil {
			return e
		}
		return errors.New("write interrupted")
	}) == nil {
		t.Fatal("expected rollback")
	}
	if count(t, f.s, "attendance") != 0 || len(queue(t, f.s)) != 0 {
		t.Fatal("interrupted mutation committed")
	}
}

func TestBoundImportRollsBackRosterOnCanonicalMismatch(t *testing.T) {
	f := setup(t)
	p := bindFixture(t, f, "http://localhost")
	old := p.Courses[0].Swimmers[0].Name
	p.Courses[0].Swimmers[0].Name = "Changed Name"
	p.Sync.Courses[0].Swimmers[0].EnrollmentID = "foreign-binding"
	if _, err := f.s.Import(p); err == nil {
		t.Fatal("import accepted canonical rebinding")
	}
	person := must(f.s.queryOne("SELECT name FROM swimmers WHERE source_id=?", p.Courses[0].Swimmers[0].SourceID))
	if person.String("name") != old {
		t.Fatal("failed canonical import partially changed roster")
	}
}

func TestBoundCatalogUpgradePreservesAcceptedHistoryAndDrafts(t *testing.T) {
	f := setup(t)
	bound := bindFixture(t, f, "http://localhost")
	_ = bound
	attempt := f.attempt(t, true, "")
	draft := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	_, err := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, err)
	check(t, f.s.Close())
	newer := f.s.Catalog
	newer.Version += "-next"
	f.s = must(OpenWithCatalog(f.path, newer))
	if count(t, f.s, "attempts") != 2 || count(t, f.s, "drafts") != 1 || len(queue(t, f.s)) != 0 {
		t.Fatal("catalog upgrade erased bound history/drafts or queued tombstones")
	}
	if _, err = f.s.queryOne("SELECT id FROM attempts WHERE id=?", attempt); err != nil {
		t.Fatal(err)
	}
	if _, err = f.s.CatalogFor(f.version); err != nil {
		t.Fatal("historical catalog lost", err)
	}
}
