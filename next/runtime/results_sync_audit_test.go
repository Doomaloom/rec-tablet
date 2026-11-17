package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"rectablet/src/platform/instance"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
	"strings"
	"sync"
	"testing"
	"time"
)

// resultsAPI is a scriptable results server keyed by bearer capability.
type resultsAPI struct {
	mu           sync.Mutex
	offered      map[string][]store.SyncOperation
	accepted     map[string]int
	reads        map[string]int
	uploadStatus map[string]int
	readStatus   map[string]int
	offline      bool
	block        chan struct{}
	started      chan struct{}
	*httptest.Server
}

func newResultsAPI(t *testing.T) *resultsAPI {
	t.Helper()
	api := &resultsAPI{offered: map[string][]store.SyncOperation{}, accepted: map[string]int{}, reads: map[string]int{}, uploadStatus: map[string]int{}, readStatus: map[string]int{}}
	api.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		api.mu.Lock()
		block, started := api.block, api.started
		api.started = nil
		api.mu.Unlock()
		if block != nil && r.URL.Path == "/api/device-results/upload" {
			if started != nil {
				close(started)
			}
			select {
			case <-block:
			case <-r.Context().Done():
			}
		}
		api.mu.Lock()
		defer api.mu.Unlock()
		if api.offline {
			http.Error(w, "offline Synthetic Pupil "+token, http.StatusServiceUnavailable)
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
			api.offered[token] = append(api.offered[token], body.Operations...)
			if status := api.uploadStatus[token]; status != 0 {
				http.Error(w, "refused Synthetic Pupil "+token, status)
				return
			}
			receipts := []map[string]any{}
			for _, o := range body.Operations {
				api.accepted[token]++
				receipts = append(receipts, map[string]any{"operationId": o.OperationID, "status": "accepted", "revision": 1})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "receipts": receipts})
		case "/api/device-results":
			api.reads[token]++
			if status := api.readStatus[token]; status != 0 {
				http.Error(w, "refused Synthetic Pupil "+token, status)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"schemaVersion": 1, "records": []any{}, "conflicts": []any{}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(api.Close)
	return api
}
func (api *resultsAPI) set(fn func()) {
	api.mu.Lock()
	defer api.mu.Unlock()
	fn()
}
func loginAs(t *testing.T, s *deviceSession, dispatch chan func(), p transfer.ClassPackage) error {
	t.Helper()
	s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) { return p, nil }
	var result error
	called := false
	s.login("123456", func(err error) { called = true; result = err })
	complete(t, dispatch)
	if !called {
		t.Fatal("login completion missing")
	}
	return result
}
func markPresent(t *testing.T, s *deviceSession) {
	t.Helper()
	var person, lesson string
	checkSession(t, s.store.DB.QueryRow("SELECT e.id,l.id FROM enrollments e JOIN lessons l ON l.course_id=e.course_id LIMIT 1").Scan(&person, &lesson))
	checkSession(t, s.store.SetAttendance(person, lesson, "present"))
}
func pendingAt(t *testing.T, path string) int {
	t.Helper()
	db, err := store.Open(path)
	checkSession(t, err)
	defer db.Close()
	pending, _, err := db.SyncStatus()
	checkSession(t, err)
	return pending
}
func noPrivateText(t *testing.T, notice string, extra ...string) {
	t.Helper()
	for _, private := range append([]string{"Synthetic Pupil", "upload-", "read-", "dataset-a", "course-", "enrollment-"}, extra...) {
		if strings.Contains(notice, private) {
			t.Fatalf("notice disclosed %q: %s", private, notice)
		}
	}
}

// The running app holds the base database lock for its whole lifetime. An
// upgraded installation keeps its first instructor in that base database.
func TestRetryUploadsIncludesBaseDatabaseProfile(t *testing.T) {
	api := newResultsAPI(t)
	s, dispatch := sessionFixture(t)
	a := boundProfile("Alex", "a", api.URL)
	legacy := a
	legacy.Sync, legacy.Credentials = nil, nil
	base, err := store.Open(s.profiles.base)
	checkSession(t, err)
	_, err = base.Import(legacy)
	checkSession(t, err)
	checkSession(t, s.profiles.registerStore(base))
	checkSession(t, base.Close())
	process, err := instance.Acquire(s.profiles.base)
	checkSession(t, err)
	defer process.Close()
	checkSession(t, loginAs(t, s, dispatch, a))
	if s.store.Path != s.profiles.base || s.lock != nil {
		t.Fatal("existing installation was not bound in place")
	}
	markPresent(t, s)
	api.set(func() { api.offline = true })
	checkSession(t, s.logout())
	api.set(func() { api.offline = false; api.uploadStatus["upload-a"] = http.StatusServiceUnavailable })
	// While the base profile cannot upload, its count is part of the aggregate.
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Sam", "b", api.URL)))
	if !strings.Contains(s.notice, "1 earlier results waiting to upload") {
		t.Fatal("base profile missing from aggregate:", s.notice)
	}
	noPrivateText(t, s.notice, "Alex", "Sam")
	checkSession(t, s.logout())
	api.set(func() { delete(api.uploadStatus, "upload-a") })
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Sam", "b", api.URL)))
	api.set(func() {
		if api.accepted["upload-a"] != 1 {
			t.Fatalf("base profile was not retried: %v", api.accepted)
		}
	})
	if strings.Contains(s.notice, "earlier") {
		t.Fatal("stale earlier status:", s.notice)
	}
	checkSession(t, s.logout())
	if pendingAt(t, s.profiles.base) != 0 {
		t.Fatal("base profile queue retained after acknowledgement")
	}
}

func TestBoundLoginWithLegacyProfileContainingUnassignedClass(t *testing.T) {
	api := newResultsAPI(t)
	s, dispatch := sessionFixture(t)
	legacy := profilePackage("Alex")
	legacy.Courses = append(legacy.Courses, transfer.ClassInput{SourceID: "course-2", Name: "Earlier", Level: "Splash1", StartTime: "10:00", Location: "Pool", LessonDates: []string{"2026-09-22"}, Swimmers: []transfer.SwimmerInput{{SourceID: "pupil-2", Name: "Synthetic Pupil"}}})
	before, err := s.profiles.importPackage(legacy)
	checkSession(t, err)
	db, err := store.Open(before.Path)
	checkSession(t, err)
	var person, lesson string
	checkSession(t, db.DB.QueryRow("SELECT e.id,l.id FROM enrollments e JOIN lessons l ON l.course_id=e.course_id LIMIT 1").Scan(&person, &lesson))
	checkSession(t, db.SetAttendance(person, lesson, "present"))
	checkSession(t, db.Close())
	// The same person now logs in with a bound code after losing course-2.
	bound := profilePackage("Alex")
	bound.Sync = &transfer.SyncIdentity{SchemaVersion: 1, SessionID: "session", DatasetID: bound.DatasetID, InstructorID: "a", Courses: []transfer.SyncCourse{{SourceID: "course-1", ClassID: "class-1", Swimmers: []transfer.SyncSwimmer{}}}}
	bound.Credentials = &transfer.SyncCredentials{SchemaVersion: 1, SessionID: "session", DatasetID: bound.DatasetID, InstructorID: "a", UploadToken: "upload-a", ReadToken: "read-a", ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339), Server: api.URL}
	checkSession(t, loginAs(t, s, dispatch, bound))
	path := s.store.Path
	if path == before.Path {
		t.Fatal("bound session adopted a legacy database with another class scope")
	}
	pending, review, err := s.store.SyncStatus()
	checkSession(t, err)
	var attendance int
	checkSession(t, s.store.DB.QueryRow("SELECT COUNT(*) FROM attendance").Scan(&attendance))
	if pending != 0 || review != 0 || attendance != 0 {
		t.Fatal("legacy work entered the bound profile")
	}
	api.set(func() {
		if len(api.offered) != 0 || api.reads["read-a"] != 1 {
			t.Fatalf("unexpected results traffic: %v %v", api.offered, api.reads)
		}
	})
	// The session holds the lock of the database it actually opened.
	if lock, err := instance.Acquire(path); err == nil {
		lock.Close()
		t.Fatal("bound profile is not locked")
	}
	lock, err := instance.Acquire(before.Path)
	checkSession(t, err)
	checkSession(t, lock.Close())
	checkSession(t, s.logout())
	// Legacy work is preserved, unbound, and remains reachable as import-only.
	db, err = store.Open(before.Path)
	checkSession(t, err)
	id, err := db.SyncInstructorID()
	checkSession(t, err)
	checkSession(t, db.DB.QueryRow("SELECT COUNT(*) FROM attendance").Scan(&attendance))
	checkSession(t, db.Close())
	if id != "" || attendance != 1 {
		t.Fatal("legacy database was bound or lost saved work")
	}
	checkSession(t, loginAs(t, s, dispatch, bound))
	if s.store.Path != path {
		t.Fatal("canonical profile changed on the next login")
	}
	checkSession(t, s.logout())
	checkSession(t, loginAs(t, s, dispatch, legacy))
	if s.store.Path != before.Path {
		t.Fatal("import-only package no longer opens the legacy profile")
	}
}

func TestRevokedGrantRetainsQueueAndRenewedCodeDrains(t *testing.T) {
	api := newResultsAPI(t)
	api.uploadStatus["upload-a"] = http.StatusForbidden
	s, dispatch := sessionFixture(t)
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Alex", "a", api.URL)))
	markPresent(t, s)
	path := s.store.Path
	checkSession(t, s.logout())
	if !strings.Contains(s.notice, "will retry") {
		t.Fatal(s.notice)
	}
	noPrivateText(t, s.notice, "Alex")
	if pendingAt(t, path) != 1 {
		t.Fatal("refused grant discarded saved work")
	}
	// A new code for the same instructor carries an independent grant.
	renewed := boundProfile("Alex", "a", api.URL)
	renewed.Credentials.UploadToken = "upload-a2"
	checkSession(t, loginAs(t, s, dispatch, renewed))
	if s.store.Path != path {
		t.Fatal("renewed code opened another profile")
	}
	api.set(func() {
		old, current := api.offered["upload-a"], api.offered["upload-a2"]
		if len(old) == 0 || len(current) != 1 || current[0].OperationID != old[0].OperationID || api.accepted["upload-a2"] != 1 || api.accepted["upload-a"] != 0 {
			t.Fatalf("renewed grant did not deliver the retained operation: %v", api.accepted)
		}
	})
	pending, _, err := s.store.SyncStatus()
	checkSession(t, err)
	if pending != 0 || !strings.Contains(s.notice, "Results uploaded") {
		t.Fatal("queue not drained with renewed grant:", s.notice)
	}
	raw, err := os.ReadFile(s.profiles.base + ".profiles.json")
	checkSession(t, err)
	if strings.Contains(string(raw), "upload-") {
		t.Fatal("profile index stores a capability")
	}
}

func TestLoginNoticeIsAggregateAndReadFailureDoesNotBlock(t *testing.T) {
	api := newResultsAPI(t)
	api.uploadStatus["upload-a"] = http.StatusServiceUnavailable
	api.readStatus["read-b"] = http.StatusInternalServerError
	s, dispatch := sessionFixture(t)
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Private Alex", "a", api.URL)))
	if !strings.Contains(s.notice, "Results uploaded") || !strings.Contains(s.notice, "Website results refreshed") {
		t.Fatal(s.notice)
	}
	markPresent(t, s)
	checkSession(t, s.logout())
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Other Sam", "b", api.URL)))
	if s.store == nil {
		t.Fatal("refresh failure blocked teaching")
	}
	for _, expected := range []string{"Results uploaded", "Website refresh unavailable; local work preserved", "1 earlier results waiting to upload", "Some earlier uploads could not finish"} {
		if !strings.Contains(s.notice, expected) {
			t.Fatalf("missing %q in %q", expected, s.notice)
		}
	}
	noPrivateText(t, s.notice, "Alex", "Sam", "Private", "Other")
	markPresent(t, s)
	checkSession(t, s.logout())
	if s.notice != "Results uploaded" {
		t.Fatal(s.notice)
	}
	checkSession(t, loginAs(t, s, dispatch, profilePackage("Guest")))
	if !strings.HasPrefix(s.notice, "Import-only session; work saved on this device") {
		t.Fatal(s.notice)
	}
	noPrivateText(t, s.notice, "Alex", "Sam", "Guest")
	api.set(func() {
		if api.reads["read-a"] != 1 || api.reads["read-b"] != 1 || len(api.reads) != 2 {
			t.Fatalf("read capabilities used outside their own login: %v", api.reads)
		}
	})
}

func TestEveryPreviousProfileRetriesWithItsOwnGrant(t *testing.T) {
	api := newResultsAPI(t)
	s, dispatch := sessionFixture(t)
	paths := map[string]string{}
	// The code was verified, but the results service stays unreachable.
	api.set(func() { api.offline = true })
	for _, id := range []string{"a", "c"} {
		checkSession(t, loginAs(t, s, dispatch, boundProfile("Alex", id, api.URL)))
		markPresent(t, s)
		paths[id] = s.store.Path
		checkSession(t, s.logout())
	}
	if paths["a"] == paths["c"] || pendingAt(t, paths["a"]) != 1 || pendingAt(t, paths["c"]) != 1 {
		t.Fatal("offline logouts did not retain separate queues")
	}
	api.set(func() { api.offline = false })
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Alex", "b", api.URL)))
	api.set(func() {
		if api.accepted["upload-a"] != 1 || api.accepted["upload-c"] != 1 || api.accepted["upload-b"] != 0 {
			t.Fatalf("saved profiles not retried with their own grants: %v", api.accepted)
		}
		for token, ops := range api.offered {
			for _, o := range ops {
				if o.EnrollmentID != "enrollment-"+strings.TrimPrefix(token, "upload-") {
					t.Fatalf("%s carried another profile's result %s", token, o.EnrollmentID)
				}
			}
		}
		if api.reads["read-b"] != 1 || len(api.reads) != 1 {
			t.Fatalf("inactive profiles were read: %v", api.reads)
		}
	})
	if strings.Contains(s.notice, "earlier") {
		t.Fatal("stale earlier status:", s.notice)
	}
	checkSession(t, s.logout())
	if pendingAt(t, paths["a"]) != 0 || pendingAt(t, paths["c"]) != 0 {
		t.Fatal("acknowledged queues retained")
	}
}

func TestLoginRefusedWhileLogoutFinishing(t *testing.T) {
	api := newResultsAPI(t)
	s, dispatch := sessionFixture(t)
	checkSession(t, s.open(boundProfile("Alex", "a", api.URL)))
	markPresent(t, s)
	started, release := make(chan struct{}), make(chan struct{})
	api.set(func() { api.block, api.started = release, started })
	finished := false
	s.logoutAsync(func(err error) { checkSession(t, err); finished = true })
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("logout upload not started")
	}
	fetches := 0
	s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) {
		fetches++
		return boundProfile("Alex", "a", api.URL), nil
	}
	var refused error
	s.login("123456", func(err error) { refused = err })
	if refused == nil || fetches != 0 || s.store != nil || s.cancel != nil {
		t.Fatal("login started while the previous session was still closing")
	}
	close(release)
	complete(t, dispatch)
	if !finished || s.loggingOut {
		t.Fatal("logout did not finish")
	}
	checkSession(t, loginAs(t, s, dispatch, boundProfile("Alex", "a", api.URL)))
}

func TestCancelLoginDuringBoundUploadReleasesProfile(t *testing.T) {
	api := newResultsAPI(t)
	s, dispatch := sessionFixture(t)
	p := boundProfile("Alex", "a", api.URL)
	checkSession(t, loginAs(t, s, dispatch, p))
	markPresent(t, s)
	path := s.store.Path
	api.set(func() { api.offline = true })
	checkSession(t, s.logout())
	started, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	api.set(func() { api.offline = false; api.block, api.started = release, started })
	s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) { return p, nil }
	s.login("123456", func(error) { t.Error("cancelled login completed") })
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("login upload not started")
	}
	s.cancelLogin()
	deadline := time.Now().Add(5 * time.Second)
	for {
		lock, err := instance.Acquire(path)
		if err == nil {
			checkSession(t, lock.Close())
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancelled login kept the profile locked:", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	select {
	case fn := <-dispatch:
		fn()
	default:
	}
	if s.store != nil || s.lock != nil || s.cancel != nil {
		t.Fatal("cancelled login opened a session")
	}
	if pendingAt(t, path) != 1 {
		t.Fatal("cancelled login lost queued work")
	}
}
