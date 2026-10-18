package repository

import (
	"os"
	"path/filepath"
	"testing"
)

// The fixture is emitted by DeckSupervisor's TypeScript exportPackage test.
// Set REC_TABLET_EXPORT_FIXTURES to test freshly generated exports directly.
func TestDeckSupervisorExportRoundTrip(t *testing.T) {
	dir := os.Getenv("REC_TABLET_EXPORT_FIXTURES")
	if dir == "" {
		dir = "testdata/decksupervisor"
	}
	s := must(Open(filepath.Join(t.TempDir(), "lessons.sqlite3")))
	defer s.Close()
	if count := must(s.ImportFile(filepath.Join(dir, "initial.json"))); count != 2 {
		t.Fatalf("expected Alex's two classes, got %d", count)
	}
	courses := must(s.courseRecords("2026-09-04"))
	if len(courses) != 2 || courses[0].String("time") != "16:00" {
		t.Fatal(courses)
	}
	course := courses[0]
	roster := must(s.rosterRecords(course.String("id")))
	if len(roster) != 2 || roster[0].String("name") != "Harrold" {
		t.Fatal(roster)
	}
	person := roster[0]
	check(t, s.SetAttendance(person.String("id"), course.String("lesson_id"), "present"))
	attempt := must(s.NewAttempt(person.String("id"), course.String("lesson_id"), "Splash1:1"))
	check(t, s.Mark(attempt, "Splash1:1:1", "successful"))
	check(t, s.Mark(attempt, "Splash1:1:2", "successful"))
	_ = must(s.Submit(attempt))
	_ = must(s.ImportFile(filepath.Join(dir, "updated.json")))
	updated := must(s.rosterRecords(course.String("id")))
	if len(updated) != 2 || updated[0].Int("active") != 1 || updated[1].Int("active") != 0 || updated[0].String("id") != person.String("id") || updated[0].String("level") != "Splash2A" {
		t.Fatal("export update lost identity, retained removed swimmer, or missed level edit", updated)
	}
	if must(s.Attendance(course.String("lesson_id")))[person.String("id")] != "present" {
		t.Fatal("repeat export lost attendance")
	}
	if must(s.Progress(person.String("id"), "Splash1:1", course.String("curriculum_version"))).Successes != 1 {
		t.Fatal("repeat export lost assessed progress")
	}
	_ = must(s.ImportFile(filepath.Join(dir, "returned.json")))
	returned := must(s.rosterRecords(course.String("id")))
	if len(returned) != 2 || returned[1].Int("active") != 1 || returned[1].String("id") != roster[1].String("id") {
		t.Fatal("returning swimmer received a different enrollment", returned)
	}
	privateRoster := must(s.rosterRecords(courses[1].String("id")))
	if len(privateRoster) != 1 || privateRoster[0].String("level") != "Splash2B" {
		t.Fatal("private swimmer's assessed level was lost", privateRoster)
	}
}
