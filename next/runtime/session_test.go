package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"rectablet/src/platform/instance"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
)

func sessionFixture(t *testing.T) (*deviceSession, chan func()) {
	t.Helper()
	profiles, err := newProfileManager(filepath.Join(t.TempDir(), "lessons.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	dispatch := make(chan func(), 10)
	s := &deviceSession{profiles: profiles, dispatch: func(fn func()) { dispatch <- fn }, fetch: transfer.FetchSharedPackage}
	t.Cleanup(s.close)
	return s, dispatch
}
func complete(t *testing.T, dispatch chan func()) {
	t.Helper()
	select {
	case fn := <-dispatch:
		fn()
	case <-time.After(5 * time.Second):
		t.Fatal("login completion did not arrive")
	}
}
func checkSession(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifiedLoginLogoutIsolationAndSavedWork(t *testing.T) {
	s, dispatch := sessionFixture(t)
	pkg, err := transfer.ReadPackage("../../examples/classes.json")
	checkSession(t, err)
	pkg.Instructor = "Alex"
	// Simulate an existing installation whose last active session was Alex.
	base, err := store.Open(s.profiles.base)
	checkSession(t, err)
	_, err = base.Import(pkg)
	checkSession(t, err)
	checkSession(t, s.profiles.registerStore(base))
	checkSession(t, base.Close())
	checkSession(t, s.profiles.activate(profileKey(pkg.DatasetID, pkg.Instructor)))
	reloaded, err := newProfileManager(s.profiles.base)
	checkSession(t, err)
	s.profiles = reloaded
	if s.store != nil {
		t.Fatal("startup resumed persisted active profile")
	}

	requests := 0
	active := true
	instructor := "Alex"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		if req.URL.Path != "/api/device-shares/redeem" || req.Method != http.MethodPost {
			t.Error("wrong redeem request")
		}
		var body struct {
			Code string `json:"code"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		if body.Code != "001234" {
			t.Errorf("lost code: %q", body.Code)
		}
		if !active {
			http.Error(w, "session code is invalid or sharing has ended", http.StatusNotFound)
			return
		}
		current := pkg
		current.Instructor = instructor
		_ = json.NewEncoder(w).Encode(map[string]any{"package": current})
	}))
	defer server.Close()
	s.server = server.URL
	login := func() error {
		var result error
		called := false
		s.login("001234", func(err error) { called = true; result = err })
		complete(t, dispatch)
		if !called {
			t.Fatal("completion missing")
		}
		return result
	}
	checkSession(t, login())
	if s.store.DatabasePath() != s.profiles.base {
		t.Fatal("existing base profile was replaced")
	}
	courses, err := s.store.ListCourses("2026-09-04")
	checkSession(t, err)
	people, err := s.store.ListRoster(courses[0].ID)
	checkSession(t, err)
	person, lesson := people[0].ID, courses[0].LessonID
	checkSession(t, s.store.SetAttendance(person, lesson, "present"))
	saved, err := s.store.NewAttempt(person, lesson, "Splash1:1")
	checkSession(t, err)
	checkSession(t, s.store.Mark(saved, "Splash1:1:1", "successful"))
	_, err = s.store.Submit(saved)
	checkSession(t, err)
	draft, err := s.store.NewAttempt(person, lesson, "Splash1:2")
	checkSession(t, err)
	checkSession(t, s.store.Mark(draft, "Splash1:2:1", "successful"))
	oldStore := s.store
	checkSession(t, s.logout())
	if s.store != nil || s.lock != nil || oldStore.DB.Ping() == nil {
		t.Fatal("logout left database open")
	}
	instructor = "Sam"
	checkSession(t, login())
	var count int
	checkSession(t, s.store.DB.QueryRow("SELECT count(*) FROM attendance").Scan(&count))
	if count != 0 {
		t.Fatal("Sam received Alex's attendance")
	}
	checkSession(t, s.store.DB.QueryRow("SELECT count(*) FROM attempts").Scan(&count))
	if count != 0 {
		t.Fatal("Sam received Alex's assessments")
	}
	samPath := s.store.DatabasePath()
	checkSession(t, s.logout())
	// Secondary instructor locks must be released at logout.
	lock, err := instance.Acquire(samPath)
	checkSession(t, err)
	checkSession(t, lock.Close())
	instructor = "Alex"
	checkSession(t, login())
	marks, err := s.store.Attendance(lesson)
	checkSession(t, err)
	if marks[person] != "present" {
		t.Fatal("attendance lost on re-login")
	}
	checkSession(t, s.store.DB.QueryRow("SELECT count(*) FROM attempts WHERE id=? AND submitted_at IS NOT NULL", saved).Scan(&count))
	if count != 1 {
		t.Fatal("saved assessment lost on re-login")
	}
	answers, err := s.store.Draft(draft)
	checkSession(t, err)
	if answers["Splash1:2:1"] != "successful" {
		t.Fatal("draft lost on re-login")
	}
	checkSession(t, s.logout())
	active = false
	if login() == nil || s.store != nil {
		t.Fatal("expired sharing reopened saved data")
	}
	if requests != 4 {
		t.Fatalf("expected verification on each login, got %d", requests)
	}
}

func TestLoginFailuresNeverOpenSession(t *testing.T) {
	for _, scenario := range []string{"network", "malformed", "invalid-package", "missing-database", "locked-database", "index-write"} {
		t.Run(scenario, func(t *testing.T) {
			s, dispatch := sessionFixture(t)
			pkg := profilePackage("Alex")
			var lock *instance.Lock
			switch scenario {
			case "network":
				s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) {
					return pkg, errors.New("could not connect")
				}
			case "malformed":
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"package":`)) }))
				defer server.Close()
				s.server = server.URL
			default:
				s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) { return pkg, nil }
				if scenario == "invalid-package" {
					pkg.Courses = nil
				}
				if scenario == "missing-database" {
					s.profiles.index.Profiles = []savedProfile{{Key: profileKey(pkg.DatasetID, pkg.Instructor), Path: filepath.Join(t.TempDir(), "missing.sqlite3")}}
				}
				if scenario == "locked-database" {
					path := s.profiles.packagePath(pkg)
					checkSession(t, os.MkdirAll(filepath.Dir(path), 0700))
					var err error
					lock, err = instance.Acquire(path)
					checkSession(t, err)
					defer lock.Close()
				}
				if scenario == "index-write" {
					checkSession(t, os.Mkdir(s.profiles.base+".profiles.json.tmp", 0700))
				}
			}
			var result error
			s.login("001234", func(err error) { result = err })
			complete(t, dispatch)
			if result == nil || s.store != nil || s.lock != nil {
				t.Fatalf("failed login exposed session: %v", result)
			}
		})
	}
}

func TestCancelledAndStaleLoginsCannotReopenSession(t *testing.T) {
	s, dispatch := sessionFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	calls := 0
	s.fetch = func(ctx context.Context, _, _ string) (transfer.ClassPackage, error) {
		calls++
		close(started)
		<-release
		return profilePackage("Alex"), nil
	}
	done := 0
	s.login("001234", func(error) { done++ })
	<-started
	s.login("001234", func(error) { done++ })
	close(release)
	// Queue a successful response, then invalidate it before the UI handles it.
	var queued func()
	select {
	case queued = <-dispatch:
	case <-time.After(5 * time.Second):
		t.Fatal("response missing")
	}
	s.cancelLogin()
	queued()
	if done != 0 || s.store != nil || calls != 1 {
		t.Fatal("stale or duplicate login completed")
	}
	s.fetch = func(context.Context, string, string) (transfer.ClassPackage, error) {
		return profilePackage("Sam"), nil
	}
	s.login("001234", func(error) { done++ })
	select {
	case queued = <-dispatch:
	case <-time.After(5 * time.Second):
		t.Fatal("response missing")
	}
	s.close()
	queued()
	if done != 0 || s.store != nil {
		t.Fatal("shutdown response reopened session")
	}
}

func TestShutdownCancelsPendingDownload(t *testing.T) {
	s, _ := sessionFixture(t)
	started, cancelled := make(chan struct{}), make(chan struct{})
	s.fetch = func(ctx context.Context, _, _ string) (transfer.ClassPackage, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		return transfer.ClassPackage{}, ctx.Err()
	}
	s.login("001234", func(error) { t.Error("cancelled download completed") })
	<-started
	s.close()
	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("download was not cancelled")
	}
}
