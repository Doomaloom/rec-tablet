package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	store "rectablet/internal/workflows/swimming/repository"
	"rectablet/internal/workflows/swimming/transfer"
)

func profilePackage(instructor string) transfer.ClassPackage {
	return transfer.ClassPackage{
		SchemaVersion: 1, Kind: "rec-tablet-classes", DatasetID: "dataset-a", Instructor: instructor,
		Courses: []transfer.ClassInput{{SourceID: "course-1", Name: "Beginner", Level: "Splash1", StartTime: "09:00", Location: "Pool", LessonDates: []string{"2026-09-22"}, Swimmers: []transfer.SwimmerInput{}}},
	}
}

func TestInstructorProfilesPersistAndStaySeparate(t *testing.T) {
	base := filepath.Join(t.TempDir(), "lessons.sqlite3")
	manager, err := newProfileManager(base)
	if err != nil {
		t.Fatal(err)
	}
	alex, err := manager.importPackage(profilePackage("Alex"))
	if err != nil {
		t.Fatal(err)
	}
	sam, err := manager.importPackage(profilePackage("Sam"))
	if err != nil {
		t.Fatal(err)
	}
	if alex.Path == sam.Path || alex.Key == sam.Key {
		t.Fatal("instructors shared the same profile")
	}
	if _, err = manager.importPackage(profilePackage("Alex")); err != nil {
		t.Fatal(err)
	}
	if len(manager.rows()) != 2 {
		t.Fatalf("got %d saved profiles", len(manager.rows()))
	}
	if err = manager.activate(alex.Key); err != nil {
		t.Fatal(err)
	}
	reloaded, err := newProfileManager(base)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.activePath(); got != alex.Path {
		t.Fatalf("active profile path %q, want %q", got, alex.Path)
	}
	for _, profile := range []savedProfile{alex, sam} {
		session, err := store.Open(profile.Path)
		if err != nil {
			t.Fatal(err)
		}
		instructor, err := session.Meta("instructor")
		closeErr := session.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if instructor != profile.Instructor {
			t.Fatalf("profile %s contains %q", profile.Instructor, instructor)
		}
	}
}

func TestShareServerSettingPersists(t *testing.T) {
	base := filepath.Join(t.TempDir(), "lessons.sqlite3")
	manager, err := newProfileManager(base)
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.setShareServer("http://localhost:8080/"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := newProfileManager(base)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.shareServer("https://default.example"); got != "http://localhost:8080" {
		t.Fatalf("server setting %q", got)
	}
}

func TestNearestLessonDate(t *testing.T) {
	if got := nearestLessonDate([]string{"2026-09-10", "2026-09-23", "2026-09-18"}, "2026-09-20"); got != "2026-09-18" {
		t.Fatalf("nearest date %q", got)
	}
	if got := nearestLessonDate(nil, "2026-09-20"); got != "2026-09-20" {
		t.Fatalf("empty schedule date %q", got)
	}
}

func TestMissingSavedProfileIsNotRecreated(t *testing.T) {
	manager, err := newProfileManager(filepath.Join(t.TempDir(), "lessons.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	key := profileKey("dataset-a", "Alex")
	missing := filepath.Join(t.TempDir(), "removed.sqlite3")
	manager.index.Profiles = []savedProfile{{Key: key, DatasetID: "dataset-a", Instructor: "Alex", Path: missing}}
	if _, err := manager.importPackage(profilePackage("Alex")); err == nil {
		t.Fatal("silently recreated a missing saved profile")
	}
	if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing profile file was recreated: %v", err)
	}
}
