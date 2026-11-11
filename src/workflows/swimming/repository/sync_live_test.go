package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"rectablet/src/workflows/swimming/transfer"
	"strings"
	"testing"
	"time"
)

// TestLiveResultsRoundTrip runs only against the disposable real Go/Postgres QA
// fixture. Ordinary tests remain independent of any external service.
func TestLiveResultsRoundTrip(t *testing.T) {
	server := strings.TrimRight(os.Getenv("REC_TEC_QA_URL"), "/")
	if server == "" {
		t.Skip("set REC_TEC_QA_URL to disposable loopback API")
	}
	if !strings.HasPrefix(server, "http://127.0.0.1:") && !strings.HasPrefix(server, "http://localhost:") {
		t.Fatal("QA API must be loopback")
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}
	request := func(method, path string, input, output any) {
		t.Helper()
		var reader io.Reader
		if input != nil {
			b, e := json.Marshal(input)
			check(t, e)
			reader = bytes.NewReader(b)
		}
		req, e := http.NewRequest(method, server+path, reader)
		check(t, e)
		req.Header.Set("Content-Type", "application/json")
		res, e := client.Do(req)
		check(t, e)
		defer res.Body.Close()
		b, e := io.ReadAll(io.LimitReader(res.Body, 16<<20))
		check(t, e)
		if res.StatusCode < 200 || res.StatusCode > 299 {
			t.Fatalf("%s %s: HTTP %d %s", method, path, res.StatusCode, b)
		}
		if output != nil {
			check(t, json.Unmarshal(b, output))
		}
	}
	request("POST", "/api/auth/sign-in", map[string]string{"email": "supervisor@example.invalid", "password": "Synthetic-test-only-42!"}, nil)
	const session = "20000000-0000-0000-0000-000000000004"
	const dataset = "decksupervisor:60000000-0000-4000-8000-000000000010"
	const enrollment = "60000000-0000-4000-8000-000000000001"
	const date = "2026-09-22"
	const courseSource = "qa-course-a"
	const swimmerSource = "qa-swimmer-a"
	path := "/api/sessions/" + session
	var classes struct {
		Classes []struct {
			ID           string `json:"id"`
			Code         string `json:"code"`
			InstructorID string `json:"instructorId"`
		} `json:"classes"`
	}
	request("GET", path+"/result-classes", nil, &classes)
	classID, instructorID := "", ""
	for _, c := range classes.Classes {
		if c.Code == "A" {
			classID, instructorID = c.ID, c.InstructorID
		}
	}
	if classID == "" || instructorID == "" {
		t.Fatal("missing fixture class A")
	}
	hash := sha256.Sum256([]byte("opaque-synthetic-result-identity-a"))
	matchKey := hex.EncodeToString(hash[:])
	registry := map[string]any{"schemaVersion": 1, "sessionId": session, "datasetId": dataset, "courses": []any{map[string]any{"sourceId": courseSource, "classId": classID, "instructorId": instructorID, "lessonDates": []string{date}, "swimmers": []any{map[string]any{"sourceId": swimmerSource, "enrollmentId": enrollment, "matchKey": matchKey}}}}}
	request("PUT", path+"/result-identities", registry, nil)
	p := transfer.ClassPackage{SchemaVersion: 1, Kind: "rec-tablet-classes", DatasetID: dataset, Instructor: "Alex", Courses: []transfer.ClassInput{{SourceID: courseSource, Name: "Synthetic QA Class", Level: "Splash1", StartTime: "09:00", Location: "Pool", LessonDates: []string{date}, Swimmers: []transfer.SwimmerInput{{SourceID: swimmerSource, Name: "Synthetic QA Swimmer"}}}}}
	var share struct {
		Codes []struct {
			Code string `json:"code"`
		} `json:"codes"`
	}
	request("POST", "/api/device-shares", map[string]any{"sessionId": session, "packages": []any{map[string]any{"instructor": p.Instructor, "package": p}}}, &share)
	if len(share.Codes) != 1 {
		t.Fatal("expected one instructor code")
	}
	ctx, stop := context.WithTimeout(context.Background(), 30*time.Second)
	defer stop()
	redeemed, e := transfer.FetchSharedPackage(ctx, server, share.Codes[0].Code)
	check(t, e)
	if redeemed.Credentials == nil || redeemed.Sync == nil {
		t.Fatal("missing bound capability")
	}
	dbPath := filepath.Join(t.TempDir(), "live.sqlite3")
	s := must(Open(dbPath))
	defer func() { _ = s.Close() }()
	_, e = s.Import(redeemed)
	check(t, e)
	check(t, s.ConfigureSync(redeemed))
	person := must(s.queryOne("SELECT id FROM enrollments")).String("id")
	lesson := must(s.queryOne("SELECT id FROM lessons")).String("id")
	check(t, s.SetAttendance(person, lesson, "present"))
	submit := func() string {
		t.Helper()
		id := must(s.NewAttempt(person, lesson, "Splash1:1"))
		check(t, s.Mark(id, "Splash1:1:1", "successful"))
		check(t, s.Mark(id, "Splash1:1:2", "successful"))
		_ = must(s.Submit(id))
		return id
	}
	attempt := submit()
	deletedAttempt := submit()
	draft := must(s.NewAttempt(person, lesson, "Splash1:2"))
	check(t, s.Mark(draft, "Splash1:2:1", "successful"))
	check(t, s.UploadPending(ctx))
	pending, review, e := s.SyncStatus()
	check(t, e)
	if pending != 0 || review != 0 {
		t.Fatalf("upload incomplete pending=%d review=%d", pending, review)
	}
	var accepted struct {
		Records []syncRecord `json:"records"`
	}
	request("GET", path+"/results", nil, &accepted)
	found := map[string]syncRecord{}
	for _, r := range accepted.Records {
		found[r.RecordID] = r
	}
	if !found[attempt].Success || found[attempt].Revision != 1 {
		t.Fatal("server did not accept full criteria")
	}
	corrections := []SyncOperation{}
	for _, id := range []string{attempt, deletedAttempt, "attendance:" + enrollment + ":" + date} {
		r, ok := found[id]
		if !ok {
			t.Fatal("missing accepted record", id)
		}
		o := r.SyncOperation
		o.OperationID = ID()
		o.BaseRevision = r.Revision
		if id == attempt {
			var p attemptPayload
			check(t, json.Unmarshal(o.Payload, &p))
			p.Results[0].Status = "not_yet"
			o.Payload, _ = json.Marshal(p)
		} else if id == deletedAttempt {
			o.Action = "delete"
			o.Payload = json.RawMessage(`{}`)
		} else {
			o.Payload = json.RawMessage(`{"status":"absent"}`)
		}
		corrections = append(corrections, o)
	}
	var receipts struct {
		Receipts []struct {
			Status string `json:"status"`
		} `json:"receipts"`
	}
	request("POST", path+"/results", map[string]any{"schemaVersion": 1, "operations": corrections}, &receipts)
	if len(receipts.Receipts) != 3 {
		t.Fatal("missing website receipts")
	}
	for _, r := range receipts.Receipts {
		if r.Status != "accepted" {
			t.Fatal("website correction", r.Status)
		}
	}
	check(t, s.ReadAccepted(ctx, redeemed.Credentials))
	if must(s.queryOne("SELECT success FROM attempts WHERE id=?", attempt)).Int("success") != 0 {
		t.Fatal("website correction not applied")
	}
	if _, e = s.queryOne("SELECT id FROM attempts WHERE id=?", deletedAttempt); e != ErrNotFound {
		t.Fatal("website tombstone not applied")
	}
	if must(s.queryOne("SELECT status FROM attendance")).String("status") != "absent" {
		t.Fatal("attendance correction not applied")
	}
	check(t, s.Close())
	s = must(Open(dbPath))
	if count(t, s, "drafts") != 1 || count(t, s, "attempts") != 2 || must(s.queryOne("SELECT COUNT(*) count FROM drafts WHERE attempt_id=?", draft)).Int("count") != 1 {
		t.Fatal("reopen lost results/draft")
	}
	pending, review, e = s.SyncStatus()
	check(t, e)
	if pending != 0 || review != 0 {
		t.Fatal("canonical changes echoed into outbox")
	}
	if evidence := os.Getenv("REC_TEC_QA_EVIDENCE"); evidence != "" {
		b, e := json.MarshalIndent(map[string]any{"sessionId": session, "datasetId": dataset, "classId": classID, "instructorId": instructorID, "courseSourceId": courseSource, "swimmerSourceId": swimmerSource, "enrollmentId": enrollment, "lessonDate": date, "attemptId": attempt, "deletedAttemptId": deletedAttempt, "curriculumVersion": s.Catalog.Version, "matchKey": matchKey, "attendance": "absent", "success": false, "revision": 2}, "", "  ")
		check(t, e)
		check(t, os.WriteFile(evidence, b, 0600))
	}
}
