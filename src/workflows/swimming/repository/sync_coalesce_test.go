package repository

import (
	"context"
	"testing"
)

// strictServer behaves like result_apply: a stale base revision goes to review
// and deleting a record the server has never stored is rejected.
func strictServer(t *testing.T) *syncServer {
	t.Helper()
	server := newSyncServer(t)
	known := map[string]bool{}
	server.status = func(o SyncOperation) string {
		if o.Action == "delete" && !known[o.RecordID] {
			return "rejected"
		}
		if o.BaseRevision != server.revisions[o.RecordID] {
			return "pending_review"
		}
		known[o.RecordID] = true
		return "accepted"
	}
	return server
}
func requests(server *syncServer) int {
	server.mu.Lock()
	defer server.mu.Unlock()
	return len(server.batches)
}

func TestSyncManyTogglesUploadInOneRequest(t *testing.T) {
	f := setup(t)
	server := strictServer(t)
	bindFixture(t, f, server.URL)
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	for i := 0; i < 50; i++ {
		status := "present"
		if i%2 == 1 {
			status = "absent"
		}
		check(t, f.s.SetAttendance(person, lesson, status))
	}
	if ops := queue(t, f.s); len(ops) != 1 || string(ops[0].Payload) != `{"status":"absent"}` {
		t.Fatalf("superseded edits kept: %d queued", len(ops))
	}
	check(t, f.s.UploadPending(context.Background()))
	if n := requests(server); n != 1 {
		t.Fatalf("50 toggles took %d requests", n)
	}
	if attendanceStatus(t, f) != "absent" || len(queue(t, f.s)) != 0 {
		t.Fatal("latest state not uploaded")
	}
	// After an accepted upload, more toggles still cost one request.
	for i := 0; i < 50; i++ {
		check(t, f.s.SetAttendance(person, lesson, []string{"present", "unmarked"}[i%2]))
	}
	check(t, f.s.UploadPending(context.Background()))
	sent := server.sent()
	if n := requests(server); n != 2 || sent[1].BaseRevision != 1 || string(sent[1].Payload) != `{"status":"unmarked"}` {
		t.Fatalf("second burst: %d requests, %+v", n, sent[len(sent)-1])
	}
}

// Submitting and undoing an assessment before any upload must still create the
// record before deleting it, because the server rejects deleting unknown records.
func TestSyncCoalescedDeleteKeepsUnsentHead(t *testing.T) {
	f := setup(t)
	server := strictServer(t)
	bindFixture(t, f, server.URL)
	for i := 0; i < 10; i++ {
		f.attempt(t, true, "")
		check(t, f.s.RemoveLatestSuccess(f.person.String("id"), "Splash1:1", f.version))
	}
	ops := queue(t, f.s)
	if len(ops) != 20 {
		t.Fatalf("each record should keep its upsert and delete: %d queued", len(ops))
	}
	check(t, f.s.UploadPending(context.Background()))
	pending, review, e := f.s.SyncStatus()
	check(t, e)
	if pending != 0 || review != 0 || requests(server) != 2 {
		t.Fatalf("pending=%d review=%d requests=%d", pending, review, requests(server))
	}
}

// An operation that was sent but not acknowledged may already be stored, so it
// is never coalesced away: it is retried unchanged, then the newest state.
func TestSyncCoalescingKeepsFrozenRetry(t *testing.T) {
	f := setup(t)
	server := strictServer(t)
	bindFixture(t, f, server.URL)
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	check(t, f.s.SetAttendance(person, lesson, "present"))
	// Freeze the head as an interrupted upload would, without a receipt.
	_, e := f.s.DB.Exec("UPDATE sync_outbox SET frozen=1")
	check(t, e)
	frozen := queue(t, f.s)[0]
	for i := 0; i < 20; i++ {
		check(t, f.s.SetAttendance(person, lesson, []string{"absent", "present"}[i%2]))
	}
	ops := queue(t, f.s)
	if len(ops) != 2 || ops[0].OperationID != frozen.OperationID {
		t.Fatalf("frozen operation lost or edits not coalesced: %d queued", len(ops))
	}
	check(t, f.s.UploadPending(context.Background()))
	sent := server.sent()
	if requests(server) != 2 || sent[0].OperationID != frozen.OperationID || sent[1].BaseRevision != 1 || string(sent[1].Payload) != `{"status":"present"}` {
		t.Fatalf("retry order: %d requests %+v", requests(server), sent)
	}
	if len(queue(t, f.s)) != 0 {
		t.Fatal("queue did not drain")
	}
}
