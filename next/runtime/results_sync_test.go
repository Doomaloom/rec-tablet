package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
	"strings"
	"sync"
	"testing"
	"time"
)

func boundProfile(name, id, server string) transfer.ClassPackage {
	p := profilePackage(name)
	p.Courses[0].SourceID = "course-" + id
	p.Courses[0].Swimmers = []transfer.SwimmerInput{{SourceID: "pupil-" + id, Name: "Synthetic Pupil"}}
	p.Sync = &transfer.SyncIdentity{SchemaVersion: 1, SessionID: "session", DatasetID: p.DatasetID, InstructorID: id, Courses: []transfer.SyncCourse{{SourceID: p.Courses[0].SourceID, ClassID: "class-" + id, Swimmers: []transfer.SyncSwimmer{{SourceID: "pupil-" + id, EnrollmentID: "enrollment-" + id}}}}}
	p.Credentials = &transfer.SyncCredentials{SchemaVersion: 1, SessionID: "session", DatasetID: p.DatasetID, InstructorID: id, UploadToken: "upload-" + id, ReadToken: "read-" + id, ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339), Server: server}
	return p
}
func TestAnyCodeLoginRetriesPreviousProfileWithUploadOnlyGrant(t *testing.T) {
	var mu sync.Mutex
	offline := false
	uploads := map[string]int{}
	reads := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if offline {
			http.Error(w, "offline", 503)
			return
		}
		switch r.URL.Path {
		case "/api/device-results/upload":
			var body struct {
				Operations []store.SyncOperation `json:"operations"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Error("invalid upload")
			}
			receipts := []map[string]any{}
			for _, o := range body.Operations {
				uploads[token]++
				receipts = append(receipts, map[string]any{"operationId": o.OperationID, "status": "accepted", "revision": 1})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "receipts": receipts})
		case "/api/device-results":
			reads[token]++
			_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "records": []any{}, "conflicts": []any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	s, dispatch := sessionFixture(t)
	p := boundProfile("Alex", "a", server.URL)
	s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) { return p, nil }
	login := func() {
		t.Helper()
		var result error
		s.login("123456", func(err error) { result = err })
		complete(t, dispatch)
		checkSession(t, result)
	}
	login()
	var person, lesson string
	checkSession(t, s.store.DB.QueryRow("SELECT e.id,l.id FROM enrollments e JOIN lessons l ON l.course_id=e.course_id LIMIT 1").Scan(&person, &lesson))
	checkSession(t, s.store.SetAttendance(person, lesson, "present"))
	aPath := s.store.Path
	draft, err := s.store.NewAttempt(person, lesson, "Splash1:1")
	checkSession(t, err)
	checkSession(t, s.store.Mark(draft, "Splash1:1:1", "successful"))
	mu.Lock()
	offline = true
	mu.Unlock()
	checkSession(t, s.logout())
	previous, err := store.Open(aPath)
	checkSession(t, err)
	pending, _, err := previous.SyncStatus()
	checkSession(t, err)
	if pending != 1 {
		t.Fatal("offline logout discarded pending work")
	}
	checkSession(t, previous.Close())
	mu.Lock()
	offline = false
	mu.Unlock()
	p = boundProfile("Alex", "b", server.URL)
	login()
	if s.store.Path == aPath {
		t.Fatal("same-name canonical instructors collided")
	}
	mu.Lock()
	if uploads["upload-a"] != 1 || reads["read-a"] != 1 || reads["read-b"] != 1 {
		t.Fatalf("unexpected profile token use uploads=%v reads=%v", uploads, reads)
	}
	mu.Unlock()
	checkSession(t, s.logout())
	p = boundProfile("Renamed Instructor", "a", server.URL)
	login()
	if s.store.Path != aPath {
		t.Fatal("display rename created another profile")
	}
	var drafts int
	checkSession(t, s.store.DB.QueryRow("SELECT COUNT(*) FROM drafts WHERE attempt_id=?", draft).Scan(&drafts))
	if drafts != 1 {
		t.Fatal("draft lost on relogin")
	}
	raw, err := json.Marshal(s.profiles.index)
	checkSession(t, err)
	if strings.Contains(string(raw), "upload-") || strings.Contains(string(raw), "read-") {
		t.Fatal("profile index leaked capabilities")
	}
}
func TestCanonicalProfileMigratesCompatibleLegacyInPlace(t *testing.T) {
	m, err := newProfileManager(filepath.Join(t.TempDir(), "base.sqlite3"))
	checkSession(t, err)
	p := boundProfile("Alex", "a", "http://localhost")
	legacy := p
	legacy.Credentials = nil
	legacy.Sync = nil
	before, err := m.importPackage(legacy)
	checkSession(t, err)
	after, err := m.importPackage(p)
	checkSession(t, err)
	if before.Path != after.Path || len(m.index.Profiles) != 1 {
		t.Fatal("legacy profile not migrated in place")
	}
	p.Instructor = "Renamed"
	again, err := m.importPackage(p)
	checkSession(t, err)
	if again.Path != before.Path {
		t.Fatal("canonical rename split profile")
	}
	// New same-name instructor must stay separate from the bound profile.
	other := boundProfile("Renamed", "b", "http://localhost")
	separate, err := m.importPackage(other)
	checkSession(t, err)
	if separate.Path == again.Path {
		t.Fatal("canonical scope leaked")
	}
}

func TestGuestPackageCannotReopenMigratedCanonicalDatabase(t *testing.T) {
	m, err := newProfileManager(filepath.Join(t.TempDir(), "base.sqlite3"))
	checkSession(t, err)
	p := boundProfile("Alex", "a", "http://localhost")
	legacy := p
	legacy.Credentials = nil
	legacy.Sync = nil
	before, err := m.importPackage(legacy)
	checkSession(t, err)
	after, err := m.importPackage(p)
	checkSession(t, err)
	if before.Path != after.Path {
		t.Fatal("migration split profile")
	}
	guest, err := m.importPackage(legacy)
	checkSession(t, err)
	if guest.Path == after.Path {
		t.Fatal("guest reopened canonical database")
	}
	db, err := store.Open(guest.Path)
	checkSession(t, err)
	defer db.Close()
	id, err := db.SyncInstructorID()
	checkSession(t, err)
	if id != "" {
		t.Fatal("guest inherited capability")
	}
}

func TestLogoutUploadRunsOutsideUIAndReportsPrivateFailure(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		http.Error(w, "Synthetic Pupil upload-secret", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	s, dispatch := sessionFixture(t)
	checkSession(t, s.open(boundProfile("Private Instructor", "a", server.URL)))
	var person, lesson string
	checkSession(t, s.store.DB.QueryRow("SELECT e.id,l.id FROM enrollments e JOIN lessons l ON l.course_id=e.course_id LIMIT 1").Scan(&person, &lesson))
	checkSession(t, s.store.SetAttendance(person, lesson, "present"))
	path := s.store.Path
	finished := false
	s.logoutAsync(func(err error) { checkSession(t, err); finished = true })
	if s.store != nil || !s.loggingOut || finished {
		t.Fatal("logout did not detach promptly")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("upload not started")
	}
	if !strings.Contains(s.notice, "Finishing") {
		t.Fatal(s.notice)
	}
	close(release)
	complete(t, dispatch)
	if !finished || s.loggingOut || !strings.Contains(s.notice, "will retry") {
		t.Fatal(s.notice)
	}
	for _, private := range []string{"Private Instructor", "Synthetic Pupil", "upload-secret"} {
		if strings.Contains(s.notice, private) {
			t.Fatal("private status", s.notice)
		}
	}
	db, err := store.Open(path)
	checkSession(t, err)
	defer db.Close()
	pending, _, err := db.SyncStatus()
	checkSession(t, err)
	if pending != 1 {
		t.Fatal("logout lost durable work")
	}
}
