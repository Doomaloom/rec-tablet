package repository

import (
	"testing"
)

func TestResetClassProgressIsScopedAndAtomic(t *testing.T) {
	f := setup(t)
	courseID, lessonID := f.course.String("id"), f.course.String("lesson_id")
	people := must(f.s.ListRoster(courseID))
	saved := f.attempt(t, true, "")
	check(t, f.s.EditAttempt(saved))
	_ = must(f.s.Submit(saved)) // Include correction history.
	nextLesson := must(one(f.s.DB, "SELECT id FROM lessons WHERE course_id=? AND id<>? LIMIT 1", courseID, lessonID)).String("id")
	f.attempt(t, false, nextLesson)
	draft := must(f.s.NewAttempt(people[1].ID, lessonID, "Splash1:1"))
	check(t, f.s.Mark(draft, "Splash1:1:1", "successful"))
	_, err := f.s.DB.Exec("UPDATE enrollments SET level='Splash7' WHERE id=?", people[2].ID)
	check(t, err)
	_ = must(f.s.NewTimedHeat([]string{people[2].ID}, lessonID, "Splash7:10", 2))
	_, err = f.s.DB.Exec("UPDATE enrollments SET active=0 WHERE id=?", people[1].ID)
	check(t, err)
	check(t, f.s.SetAttendance(people[0].ID, lessonID, "present"))

	other := must(one(f.s.DB, "SELECT l.id AS lesson,e.id AS person,e.level FROM lessons l JOIN enrollments e ON e.course_id=l.course_id WHERE l.course_id<>? AND e.active=1 LIMIT 1", courseID))
	catalog := must(f.s.CatalogFor(f.version))
	level, _ := catalog.Level(other.String("level"))
	otherAttempt := must(f.s.NewAttempt(other.String("person"), other.String("lesson"), level.Skills[0].ID))
	_ = must(f.s.Submit(otherAttempt))

	// A late failure must roll back deletions already made in child tables.
	_, err = f.s.DB.Exec(`CREATE TRIGGER block_progress_reset BEFORE DELETE ON attempts BEGIN SELECT RAISE(ABORT, 'blocked reset'); END`)
	check(t, err)
	if err := f.s.ResetClassProgress(courseID); err == nil {
		t.Fatal("expected reset failure")
	}
	if p := f.progress(); p.Assessed != 2 || p.Successes != 1 {
		t.Fatalf("failed reset changed progress: %+v", p)
	}
	if values := must(answers(f.s.DB, saved)); len(values) == 0 {
		t.Fatal("failed reset removed results")
	}
	_, err = f.s.DB.Exec("DROP TRIGGER block_progress_reset")
	check(t, err)
	check(t, f.s.ResetClassProgress(courseID))
	for _, table := range []string{"attempts", "results", "drafts", "revisions", "timed_attempts"} {
		var count int
		query := "SELECT COUNT(*) FROM " + table
		if table == "attempts" {
			query += " WHERE id<>?"
		} else {
			query += " WHERE attempt_id<>?"
		}
		check(t, f.s.DB.QueryRow(query, otherAttempt).Scan(&count))
		if count != 0 {
			t.Fatalf("reset left %d rows in %s", count, table)
		}
	}
	if p := f.progress(); p.Assessed != 0 || p.Successes != 0 {
		t.Fatalf("progress remains: %+v", p)
	}
	if a := must(f.s.GetAttempt(otherAttempt)); a.SubmittedAt == nil {
		t.Fatal("other class attempt changed")
	}
	attendance := must(f.s.Attendance(lessonID))
	if attendance[people[0].ID] != "present" {
		t.Fatal("reset changed attendance")
	}
	if got := must(f.s.ListRoster(courseID)); len(got) != len(people) {
		t.Fatal("reset changed roster")
	}
	if err := f.s.ResetClassProgress("missing"); err == nil {
		t.Fatal("missing class accepted")
	}
}
