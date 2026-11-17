package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"rectablet/src/workflows/swimming/domain"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

// syncServer is a scriptable results API. status decides each upload receipt;
// pull supplies the complete read response.
type syncServer struct {
	mu        sync.Mutex
	revisions map[string]int
	batches   [][]SyncOperation
	bodies    []string
	status    func(SyncOperation) string
	pull      func() map[string]any
	*httptest.Server
}

func newSyncServer(t *testing.T) *syncServer {
	t.Helper()
	s := &syncServer{revisions: map[string]int{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch {
		case r.URL.Path == "/api/device-results/upload" && r.Header.Get("Authorization") == "Bearer upload-secret":
			raw, _ := io.ReadAll(r.Body)
			var req struct {
				Operations []SyncOperation `json:"operations"`
			}
			if json.Unmarshal(raw, &req) != nil {
				http.Error(w, "JSON", 400)
				return
			}
			s.bodies = append(s.bodies, string(raw))
			s.batches = append(s.batches, req.Operations)
			receipts := []map[string]any{}
			for _, o := range req.Operations {
				status := "accepted"
				if s.status != nil {
					status = s.status(o)
				}
				if status == "accepted" {
					s.revisions[o.RecordID]++
				}
				receipts = append(receipts, map[string]any{"operationId": o.OperationID, "status": status, "revision": s.revisions[o.RecordID]})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "receipts": receipts})
		case r.URL.Path == "/api/device-results" && r.Header.Get("Authorization") == "Bearer read-secret":
			body := map[string]any{"schemaVersion": 1, "records": []any{}, "conflicts": []any{}}
			if s.pull != nil {
				body = s.pull()
			}
			_ = json.NewEncoder(w).Encode(body)
		default:
			http.Error(w, "scope", 403)
		}
	}))
	t.Cleanup(s.Close)
	return s
}
func (s *syncServer) sent() []SyncOperation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []SyncOperation{}
	for _, b := range s.batches {
		out = append(out, b...)
	}
	return out
}
func stateRevision(t *testing.T, s *Store, record string) int {
	t.Helper()
	return must(s.queryOne("SELECT revision FROM sync_state WHERE record_id=?", record)).Int("revision")
}
func attendanceStatus(t *testing.T, f *fixture) string {
	t.Helper()
	return must(f.s.Attendance(f.course.String("lesson_id")))[f.person.String("id")]
}

func TestSyncPullPreservesUnsentLocalEdit(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	local := queue(t, f.s)[0]
	incoming := local
	incoming.Payload = json.RawMessage(`{"status":"absent"}`)
	server.pull = func() map[string]any {
		return map[string]any{"schemaVersion": 1, "records": []syncRecord{{SyncOperation: incoming, Revision: 5}}}
	}
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if attendanceStatus(t, f) != "present" {
		t.Fatal("pull overwrote an unsent local edit")
	}
	ops := queue(t, f.s)
	if len(ops) != 1 || ops[0].OperationID != local.OperationID || string(ops[0].Payload) != string(local.Payload) {
		t.Fatalf("unsent operation changed by pull: %+v", ops)
	}
	if stateRevision(t, f.s, local.RecordID) != 0 {
		t.Fatal("pull advanced the revision of an unsent edit")
	}
}

func TestSyncPullSkipsAttemptBeingCorrected(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	attempt := f.attempt(t, true, "")
	op := queue(t, f.s)[0]
	_, e := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, e)
	check(t, f.s.EditAttempt(attempt))
	check(t, f.s.Mark(attempt, "Splash1:1:1", "not_yet"))
	before := must(f.s.Draft(attempt))
	var payload attemptPayload
	check(t, json.Unmarshal(op.Payload, &payload))
	for i := range payload.Results {
		payload.Results[i].Status = "not_yet"
	}
	changed := op
	changed.Payload, _ = json.Marshal(payload)
	record := syncRecord{SyncOperation: changed, Revision: 2}
	server.pull = func() map[string]any { return map[string]any{"schemaVersion": 1, "records": []syncRecord{record}} }
	verify := func(stage string) {
		t.Helper()
		if f.progress().Successes != 1 || count(t, f.s, "attempts") != 1 {
			t.Fatal(stage, "replaced an attempt that is being corrected")
		}
		after := must(f.s.Draft(attempt))
		if len(after) != len(before) || after["Splash1:1:1"] != "not_yet" || after["Splash1:1:2"] != before["Splash1:1:2"] {
			t.Fatal(stage, "changed the correction draft", after)
		}
		if len(queue(t, f.s)) != 0 {
			t.Fatal(stage, "queued an operation")
		}
	}
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	verify("upsert")
	record.Revision, record.Deleted, record.Action, record.Payload = 3, true, "delete", json.RawMessage(`{}`)
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	verify("tombstone")
}

func TestSyncResolvedConflictUnblocksRecord(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	mark := func(status string) {
		t.Helper()
		check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), status))
	}
	mark("present")
	check(t, f.s.UploadPending(context.Background()))
	server.mu.Lock()
	server.status = func(SyncOperation) string { return "pending_review" }
	server.mu.Unlock()
	mark("absent")
	check(t, f.s.UploadPending(context.Background()))
	reviewed := server.sent()[1]
	mark("present")
	resolved := false
	server.mu.Lock()
	server.pull = func() map[string]any {
		return map[string]any{"schemaVersion": 1, "records": []any{}, "conflicts": []any{map[string]any{"id": "conflict", "operation": reviewed, "resolved": resolved}}}
	}
	server.mu.Unlock()
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	check(t, f.s.UploadPending(context.Background()))
	pending, review, e := f.s.SyncStatus()
	check(t, e)
	if pending != 1 || review != 1 || len(server.sent()) != 2 {
		t.Fatalf("unresolved conflict released its record: pending=%d review=%d sent=%d", pending, review, len(server.sent()))
	}
	server.mu.Lock()
	resolved = true
	server.status = nil
	server.mu.Unlock()
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if _, review, _ = f.s.SyncStatus(); review != 0 {
		t.Fatal("resolved conflict still awaiting review")
	}
	check(t, f.s.UploadPending(context.Background()))
	pending, review, e = f.s.SyncStatus()
	check(t, e)
	sent := server.sent()
	if pending != 0 || review != 0 || len(sent) != 3 || sent[2].RecordID != reviewed.RecordID || sent[2].OperationID == reviewed.OperationID {
		t.Fatalf("queued edit not sent after resolution: pending=%d review=%d sent=%d", pending, review, len(sent))
	}
}

// synthetic queues n distinct captured records without touching domain tables.
func synthetic(t *testing.T, s *Store, prefix string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("%s%03d", prefix, i)
		o := SyncOperation{OperationID: ID(), SessionID: "session", DatasetID: "dataset", ClassID: "class", EnrollmentID: "enrollment", RecordID: id, Kind: "attendance", Action: "upsert", LessonDate: "2026-09-04", CurriculumVersion: s.Catalog.Version, Payload: json.RawMessage(`{"status":"present"}`)}
		b, _ := json.Marshal(o)
		_, e := s.DB.Exec("INSERT INTO sync_outbox(operation_id,record_id,body) VALUES(?,?,?)", o.OperationID, id, string(b))
		check(t, e)
		// Tombstone state keeps later captures from treating the record as removed.
		o.Action, o.Payload = "delete", json.RawMessage(`{}`)
		_, e = s.DB.Exec("INSERT INTO sync_state VALUES(?,?,0)", id, stateBody(o))
		check(t, e)
	}
}

func TestSyncUploadBatchesAtMost200(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	bindFixture(t, f, server.URL)
	synthetic(t, f.s, "record-", 201)
	check(t, f.s.UploadPending(context.Background()))
	server.mu.Lock()
	defer server.mu.Unlock()
	if len(server.batches) != 2 || len(server.batches[0]) != 200 || len(server.batches[1]) != 1 {
		t.Fatalf("unexpected batching: %d requests", len(server.batches))
	}
	if len(queue(t, f.s)) != 0 {
		t.Fatal("queue did not drain")
	}
}

func keys(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var m map[string]json.RawMessage
	check(t, json.Unmarshal(raw, &m))
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func TestSyncUploadWireFormatAndPrivacy(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	bindFixture(t, f, server.URL)
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	check(t, f.s.SetAttendance(person, lesson, "present"))
	attempt := f.attempt(t, true, "")
	heat := must(f.s.NewTimedHeat([]string{person}, lesson, "Splash1:2", 2))
	check(t, f.s.SaveTimedHeat([]domain.TimedResult{{AttemptID: heat[0].AttemptID, LapsRequired: 2, LapsCompleted: 2, ElapsedMS: 1000, Success: true}}))
	draft := must(f.s.NewAttempt(person, lesson, "Splash1:3"))
	_ = draft
	check(t, f.s.RemoveLatestSuccess(person, "Splash1:1", f.version))
	check(t, f.s.UploadPending(context.Background()))
	server.mu.Lock()
	defer server.mu.Unlock()
	uuid := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	date := must(f.s.queryOne("SELECT date FROM lessons WHERE id=?", lesson)).String("date")
	seen := map[string]int{}
	for _, body := range server.bodies {
		if keys(t, json.RawMessage(body)) != "operations,schemaVersion" {
			t.Fatal("unexpected request fields", keys(t, json.RawMessage(body)))
		}
		forbidden := []string{f.p.Instructor, `"success"`, "draft", "upload-secret", "read-secret", `"name"`, `"instructor"`, draft}
		for _, c := range f.p.Courses {
			forbidden = append(forbidden, c.Name, c.Location)
			for _, w := range c.Swimmers {
				forbidden = append(forbidden, w.Name)
			}
		}
		for _, private := range forbidden {
			if private != "" && strings.Contains(body, private) {
				t.Fatalf("upload disclosed %q", private)
			}
		}
		var req struct {
			SchemaVersion int               `json:"schemaVersion"`
			Operations    []json.RawMessage `json:"operations"`
		}
		check(t, json.Unmarshal([]byte(body), &req))
		if req.SchemaVersion != 1 {
			t.Fatal("schemaVersion", req.SchemaVersion)
		}
		for _, raw := range req.Operations {
			if keys(t, raw) != "action,baseRevision,classId,curriculumVersion,datasetId,enrollmentId,kind,lessonDate,operationId,payload,recordId,sessionId" {
				t.Fatal("operation fields", keys(t, raw))
			}
			var o SyncOperation
			check(t, json.Unmarshal(raw, &o))
			if !uuid.MatchString(o.OperationID) || o.SessionID != "session" || o.DatasetID != f.p.DatasetID || o.LessonDate != date || o.CurriculumVersion != f.version || !strings.HasPrefix(o.ClassID, "class-") || !strings.HasPrefix(o.EnrollmentID, "enrollment-") {
				t.Fatalf("operation identity %+v", o)
			}
			seen[o.Kind+":"+o.Action]++
			switch {
			case o.Action == "delete":
				if string(o.Payload) != "{}" || o.RecordID != attempt || o.BaseRevision != 1 {
					t.Fatalf("tombstone %s", raw)
				}
			case o.Kind == "attendance":
				if o.RecordID != "attendance:"+o.EnrollmentID+":"+o.LessonDate || string(o.Payload) != `{"status":"present"}` || o.BaseRevision != 0 {
					t.Fatalf("attendance %s", raw)
				}
			case o.Kind == "attempt":
				want := "createdAt,results,skillId,submittedAt"
				timed := o.RecordID == heat[0].AttemptID
				if timed {
					want += ",timed"
				}
				if !uuid.MatchString(o.RecordID) || keys(t, o.Payload) != want || o.BaseRevision != 0 {
					t.Fatalf("attempt %s", raw)
				}
				var p struct {
					Results []json.RawMessage `json:"results"`
					Timed   json.RawMessage   `json:"timed"`
				}
				check(t, json.Unmarshal(o.Payload, &p))
				if len(p.Results) == 0 {
					t.Fatal("criteria missing")
				}
				for _, r := range p.Results {
					var v struct{ SubskillID, Status string }
					check(t, json.Unmarshal(r, &v))
					if keys(t, r) != "status,subskillId" || (v.Status != "successful" && v.Status != "not_yet") {
						t.Fatalf("criterion %s", r)
					}
				}
				if timed && (keys(t, p.Timed) != "elapsedMs,lapsCompleted,lapsRequired" || string(p.Timed) != `{"elapsedMs":1000,"lapsCompleted":2,"lapsRequired":2}`) {
					t.Fatalf("timing %s", p.Timed)
				}
			default:
				t.Fatalf("unexpected operation %s", raw)
			}
		}
	}
	if seen["attendance:upsert"] != 1 || seen["attempt:upsert"] != 2 || seen["attempt:delete"] != 1 || len(seen) != 3 {
		t.Fatalf("operations sent: %v", seen)
	}
}

func TestSyncRejectedReceiptRetainedWithoutBlockingOtherRecords(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	bindFixture(t, f, server.URL)
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	rejected := queue(t, f.s)[0]
	server.status = func(o SyncOperation) string {
		if o.RecordID == rejected.RecordID {
			return "rejected"
		}
		return "accepted"
	}
	attempt := f.attempt(t, true, "")
	if f.s.UploadPending(context.Background()) == nil {
		t.Fatal("rejected result reported as uploaded")
	}
	ops := queue(t, f.s)
	if len(ops) != 1 || ops[0].OperationID != rejected.OperationID || stateRevision(t, f.s, attempt) != 1 || stateRevision(t, f.s, rejected.RecordID) != 0 {
		t.Fatalf("rejected receipt handling: %+v", ops)
	}
	// A later edit to another record is still delivered, and the rejected
	// operation is offered again unchanged.
	check(t, f.s.EditAttempt(attempt))
	check(t, f.s.Mark(attempt, "Splash1:1:1", "not_yet"))
	_ = must(f.s.Submit(attempt))
	_ = f.s.UploadPending(context.Background())
	ops = queue(t, f.s)
	if len(ops) != 1 || ops[0].OperationID != rejected.OperationID || stateRevision(t, f.s, attempt) != 2 {
		t.Fatalf("rejected record blocked another record: %+v", ops)
	}
	offered := 0
	for _, o := range server.sent() {
		if o.RecordID == rejected.RecordID {
			offered++
			b, _ := json.Marshal(o)
			want, _ := json.Marshal(rejected)
			if o.OperationID != rejected.OperationID || string(b) != string(want) {
				t.Fatal("rejected operation changed on retry")
			}
		}
	}
	if offered != 2 {
		t.Fatalf("rejected operation offered %d times in two uploads", offered)
	}
	pending, review, e := f.s.SyncStatus()
	check(t, e)
	if pending != 1 || review != 0 {
		t.Fatal("rejected result left the durable queue")
	}
}

func TestSyncRejectedBacklogDoesNotStarveQueue(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	bindFixture(t, f, server.URL)
	synthetic(t, f.s, "reject-", 200)
	server.status = func(o SyncOperation) string {
		if strings.HasPrefix(o.RecordID, "reject-") {
			return "rejected"
		}
		return "accepted"
	}
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	if f.s.UploadPending(context.Background()) == nil {
		t.Fatal("rejected results reported as uploaded")
	}
	ops := queue(t, f.s)
	if len(ops) != 200 {
		t.Fatalf("queue holds %d operations; rejected backlog starved or lost work", len(ops))
	}
	for _, o := range ops {
		if !strings.HasPrefix(o.RecordID, "reject-") {
			t.Fatal("acceptable result stuck behind rejected backlog")
		}
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if len(server.batches) != 2 {
		t.Fatalf("rejected operations resent within one upload: %d requests", len(server.batches))
	}
}

// A device installed after a catalog version change stores its courses at the
// new version. Accepted history recorded at another version is refused, and
// that refusal rolls back the entire refresh; local work is left untouched.
func TestSyncPullOnFreshDeviceAfterCatalogBump(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	attendance := queue(t, f.s)[0]
	attempt := f.attempt(t, true, "")
	history := queue(t, f.s)[1]
	_, e := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, e)
	check(t, f.s.RemoveLatestSuccess(f.person.String("id"), "Splash1:1", f.version))
	_, e = f.s.DB.Exec("DELETE FROM sync_outbox; DELETE FROM sync_state WHERE record_id=?", attempt)
	check(t, e)
	attendance.Payload = json.RawMessage(`{"status":"absent"}`)
	history.CurriculumVersion = "version-of-another-device"
	server.pull = func() map[string]any {
		return map[string]any{"schemaVersion": 1, "records": []syncRecord{{SyncOperation: attendance, Revision: 2}, {SyncOperation: history, Revision: 1, Success: true}}}
	}
	err := f.s.ReadAccepted(context.Background(), p.Credentials)
	if err == nil || !strings.Contains(err.Error(), "curriculum") {
		t.Fatal("history from another curriculum version accepted:", err)
	}
	if attendanceStatus(t, f) != "present" || count(t, f.s, "attempts") != 0 || stateRevision(t, f.s, attendance.RecordID) != 0 || len(queue(t, f.s)) != 0 {
		t.Fatal("refused refresh was partially applied")
	}
}

func TestBindingExistingDatabaseQueuesHistoryOnce(t *testing.T) {
	f := setup(t)
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	check(t, f.s.SetAttendance(person, lesson, "present"))
	attempt := f.attempt(t, true, "")
	draft := must(f.s.NewAttempt(person, lesson, "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	if count(t, f.s, "sync_outbox") != 0 {
		t.Fatal("import-only database queued results")
	}
	p := bindFixture(t, f, "http://localhost")
	verify := func(stage string) {
		t.Helper()
		if count(t, f.s, "attendance") != 1 || count(t, f.s, "attempts") != 2 || count(t, f.s, "drafts") != 1 || f.progress().Successes != 1 {
			t.Fatal(stage, "lost existing work")
		}
		ops := queue(t, f.s)
		if len(ops) != 2 {
			t.Fatalf("%s queued %d operations", stage, len(ops))
		}
		kinds := map[string]bool{}
		for _, o := range ops {
			kinds[o.Kind] = true
			if o.Action != "upsert" || o.BaseRevision != 0 || o.RecordID == draft || (o.Kind == "attempt" && o.RecordID != attempt) {
				t.Fatalf("%s queued %+v", stage, o)
			}
		}
		if !kinds["attendance"] || !kinds["attempt"] {
			t.Fatal(stage, "did not queue existing history")
		}
	}
	verify("binding")
	first := queue(t, f.s)
	_ = must(f.s.Import(p))
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	verify("re-import")
	for i, o := range queue(t, f.s) {
		if o.OperationID != first[i].OperationID {
			t.Fatal("re-import replaced a queued operation")
		}
	}
}

func TestSyncLocalEditAfterPullUsesPulledRevision(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	op := queue(t, f.s)[0]
	_, e := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, e)
	op.Payload = json.RawMessage(`{"status":"absent"}`)
	record := syncRecord{SyncOperation: op, Revision: 4}
	server.pull = func() map[string]any { return map[string]any{"schemaVersion": 1, "records": []syncRecord{record}} }
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if attendanceStatus(t, f) != "absent" || len(queue(t, f.s)) != 0 {
		t.Fatal("correction not applied quietly")
	}
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	check(t, f.s.UploadPending(context.Background()))
	sent := server.sent()
	if len(sent) != 1 || sent[0].BaseRevision != 4 {
		t.Fatalf("edit after correction sent %+v", sent)
	}
	// The same holds after an accepted deletion.
	record.Revision, record.Deleted, record.Action, record.Payload = 6, true, "delete", json.RawMessage(`{}`)
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	if attendanceStatus(t, f) != "" || len(queue(t, f.s)) != 0 {
		t.Fatal("tombstone not applied quietly")
	}
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "absent"))
	if ops := queue(t, f.s); len(ops) != 1 || ops[0].BaseRevision != 6 {
		t.Fatalf("edit after tombstone queued %+v", ops)
	}
}

func TestBoundRosterRemovalKeepsHistoryWithoutTombstones(t *testing.T) {
	f := setup(t)
	p := bindFixture(t, f, "http://localhost")
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	check(t, f.s.SetAttendance(person, lesson, "present"))
	attempt := f.attempt(t, true, "")
	before := queue(t, f.s)
	source := must(f.s.queryOne("SELECT w.source_id FROM enrollments e JOIN swimmers w ON w.id=e.swimmer_id WHERE e.id=?", person)).String("source_id")
	removed := 0
	for i := range p.Courses {
		if p.Courses[i].SourceID != f.course.String("source_id") {
			continue
		}
		for j, w := range p.Courses[i].Swimmers {
			if w.SourceID == source {
				p.Courses[i].Swimmers = append(p.Courses[i].Swimmers[:j:j], p.Courses[i].Swimmers[j+1:]...)
				removed++
				break
			}
		}
		for j, w := range p.Sync.Courses[i].Swimmers {
			if w.SourceID == source {
				p.Sync.Courses[i].Swimmers = append(p.Sync.Courses[i].Swimmers[:j:j], p.Sync.Courses[i].Swimmers[j+1:]...)
				removed++
				break
			}
		}
	}
	if removed != 2 {
		t.Fatal("fixture swimmer not found")
	}
	_ = must(f.s.Import(p))
	if must(f.s.queryOne("SELECT active FROM enrollments WHERE id=?", person)).Int("active") != 0 {
		t.Fatal("removed swimmer still active")
	}
	after := queue(t, f.s)
	if len(after) != len(before) || count(t, f.s, "attendance") != 1 || count(t, f.s, "attempts") != 1 {
		t.Fatalf("roster removal changed history or outbox: %d -> %d", len(before), len(after))
	}
	for i := range after {
		if after[i].OperationID != before[i].OperationID || after[i].Action != "upsert" {
			t.Fatal("roster removal queued a deletion")
		}
	}
	if _, e := f.s.queryOne("SELECT id FROM attempts WHERE id=?", attempt); e != nil {
		t.Fatal(e)
	}
}
