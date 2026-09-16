package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"rectablet/internal/platform/files"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/transfer"
	"reflect"
	"strings"
	"testing"
)

func must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}
func check(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}

type fixture struct {
	s              *Store
	p              transfer.ClassPackage
	course, person record
	version, path  string
}

func setup(t *testing.T) *fixture {
	t.Helper()
	path := filepath.Join(t.TempDir(), "lesson.sqlite3")
	s := must(Open(path))
	p := must(transfer.ReadPackage("../../../../examples/classes.json"))
	_ = must(s.Import(p))
	c := must(s.courseRecords("2026-09-04"))[0]
	person := must(s.rosterRecords(c.String("id")))[0]
	f := &fixture{s, p, c, person, c.String("curriculum_version"), path}
	t.Cleanup(func() { f.s.Close() })
	return f
}
func (f *fixture) attempt(t *testing.T, success bool, lesson string) string {
	t.Helper()
	if lesson == "" {
		lesson = f.course.String("lesson_id")
	}
	a := must(f.s.NewAttempt(f.person.String("id"), lesson, "Splash1:1"))
	check(t, f.s.Mark(a, "Splash1:1:1", "successful"))
	status := "successful"
	if !success {
		status = "not_yet"
	}
	check(t, f.s.Mark(a, "Splash1:1:2", status))
	_ = must(f.s.Submit(a))
	return a
}
func (f *fixture) progress() domain.Progress {
	return must(f.s.Progress(f.person.String("id"), "Splash1:1", f.version))
}
func TestAllCriteriaAndFailures(t *testing.T) {
	f := setup(t)
	a := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	check(t, f.s.Mark(a, "Splash1:1:1", "successful"))
	if must(f.s.Submit(a)) {
		t.Fatal("incomplete attempt was marked successful")
	}
	if p := f.progress(); p.Assessed != 1 || p.Successes != 0 {
		t.Fatal(f.progress())
	}
	var drafts int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM drafts WHERE attempt_id=?", a).Scan(&drafts))
	if drafts != 0 {
		t.Fatal("saved attempt retained its draft")
	}
	results := must(f.s.queryRows("SELECT * FROM results WHERE attempt_id=?", a))
	if len(results) != 2 {
		t.Fatalf("saved incomplete attempt has %d results, want 2", len(results))
	}
	for _, result := range results {
		if result.String("status") != "not_yet" && result.String("subskill_id") == "Splash1:1:2" {
			t.Fatal("unchecked subskill was not saved as not_yet")
		}
	}
	if p := f.progress(); p.Assessed != 1 || p.Successes != 0 {
		t.Fatal(p)
	}
	feedback := must(f.s.Feedback(f.person.String("id")))
	if len(feedback) != 1 || len(feedback[0].Remaining) != 1 {
		t.Fatalf("saved incomplete attempt missing from feedback: %+v", feedback)
	}
}
func TestThreeSameLessonSuccessesAccumulate(t *testing.T) {
	f := setup(t)
	f.attempt(t, true, "")
	f.attempt(t, false, "")
	f.attempt(t, true, "")
	if f.progress().Complete {
		t.Fatal("too early")
	}
	f.attempt(t, true, "")
	f.attempt(t, false, "")
	if p := f.progress(); !p.Complete || p.Successes != 3 {
		t.Fatal(p)
	}
}

func TestTimedHeatPersistsLapsAndElapsedResults(t *testing.T) {
	f := setup(t)
	level := "Splash7"
	for i := range f.p.Courses[0].Swimmers[:2] {
		f.p.Courses[0].Swimmers[i].Level = &level
	}
	_ = must(f.s.Import(f.p))
	roster := must(f.s.rosterRecords(f.course.String("id")))
	people := []string{}
	for _, swimmer := range roster {
		if swimmer.String("level") == level {
			people = append(people, swimmer.String("id"))
		}
	}
	if len(people) != 2 {
		t.Fatalf("expected two Splash7 swimmers, got %d", len(people))
	}
	participants := must(f.s.NewTimedHeat(people, f.course.String("lesson_id"), "Splash7:10", 3))
	if len(participants) != 2 || participants[0].LapsRequired != 3 {
		t.Fatalf("unexpected heat participants: %+v", participants)
	}
	for i := 0; i < 3; i++ {
		if laps := must(f.s.RecordTimedLap(participants[0].AttemptID)); laps != i+1 {
			t.Fatalf("lap %d recorded as %d", i+1, laps)
		}
	}
	check(t, f.s.SaveTimedHeat([]domain.TimedResult{
		{AttemptID: participants[0].AttemptID, LapsRequired: 3, LapsCompleted: 3, ElapsedMS: 2500, Success: true},
		{AttemptID: participants[1].AttemptID, LapsRequired: 3, LapsCompleted: 1, ElapsedMS: 5000, Success: false},
	}))
	if p := must(f.s.Progress(people[0], "Splash7:10", f.version)); p.Successes != 1 || p.Assessed != 1 {
		t.Fatalf("successful timed progress incorrect: %+v", p)
	}
	if p := must(f.s.Progress(people[1], "Splash7:10", f.version)); p.Successes != 0 || p.Assessed != 1 {
		t.Fatalf("failed timed progress incorrect: %+v", p)
	}
	feedback := must(f.s.Feedback(people[1]))
	if len(feedback) != 1 || !feedback[0].Timed || feedback[0].TimedLaps != 1 || feedback[0].TimedRequired != 3 || feedback[0].TimedElapsedMS != 5000 {
		t.Fatalf("timed feedback data missing: %+v", feedback)
	}
}

func TestMultiAttemptsAreAtomicAndRestoreReusedDrafts(t *testing.T) {
	f := setup(t)
	roster := must(f.s.rosterRecords(f.course.String("id")))
	people := []string{roster[0].String("id"), roster[1].String("id")}
	lesson := f.course.String("lesson_id")
	opened := must(f.s.OpenMultiAttempts(people, lesson, "Splash1:1"))
	if len(opened) != 2 || !opened[0].Created || !opened[1].Created {
		t.Fatalf("unexpected newly opened attempts: %+v", opened)
	}

	// Reopening an unfinished attempt reuses its draft and snapshots the
	// current answers so discarding the session can restore them.
	check(t, f.s.Mark(opened[0].AttemptID, "Splash1:1:1", "successful"))
	reused := must(f.s.OpenMultiAttempts([]string{people[0]}, lesson, "Splash1:1"))
	if len(reused) != 1 || reused[0].Created || reused[0].OriginalAnswers["Splash1:1:1"] != "successful" {
		t.Fatalf("unfinished draft was not reused: %+v", reused)
	}
	check(t, f.s.Mark(reused[0].AttemptID, "Splash1:1:2", "successful"))
	check(t, f.s.DiscardMultiAttempts(reused))
	restored := must(f.s.Draft(opened[0].AttemptID))
	if len(restored) != 1 || restored["Splash1:1:1"] != "successful" {
		t.Fatalf("discard did not restore original draft: %+v", restored)
	}
	check(t, f.s.DiscardMultiAttempts(opened))
	var remaining int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM attempts WHERE submitted_at IS NULL").Scan(&remaining))
	if remaining != 0 {
		t.Fatalf("discard left %d unfinished attempts", remaining)
	}

	opened = must(f.s.OpenMultiAttempts(people, lesson, "Splash1:1"))
	check(t, f.s.Mark(opened[0].AttemptID, "Splash1:1:1", "successful"))
	check(t, f.s.Mark(opened[0].AttemptID, "Splash1:1:2", "successful"))
	check(t, f.s.Mark(opened[1].AttemptID, "Splash1:1:1", "successful"))
	if err := f.s.SubmitAttempts([]string{opened[0].AttemptID, "missing-attempt"}); err == nil {
		t.Fatal("batch save accepted a missing attempt")
	}
	for _, attempt := range opened {
		var submitted any
		check(t, f.s.DB.QueryRow("SELECT submitted_at FROM attempts WHERE id=?", attempt.AttemptID).Scan(&submitted))
		if submitted != nil {
			t.Fatal("failed batch save partially submitted an attempt")
		}
	}
	check(t, f.s.SubmitAttempts([]string{opened[0].AttemptID, opened[1].AttemptID}))
	if p := must(f.s.Progress(people[0], "Splash1:1", f.version)); p.Successes != 1 || p.Assessed != 1 {
		t.Fatalf("complete multi attempt progress incorrect: %+v", p)
	}
	if p := must(f.s.Progress(people[1], "Splash1:1", f.version)); p.Successes != 0 || p.Assessed != 1 {
		t.Fatalf("partial multi attempt progress incorrect: %+v", p)
	}
}

func TestRemoveLatestSuccessDecrementsProgressAndRemovesAttempt(t *testing.T) {
	f := setup(t)
	first := f.attempt(t, true, "")
	second := f.attempt(t, true, "")
	if f.progress().Successes != 2 {
		t.Fatal(f.progress())
	}
	check(t, f.s.RemoveLatestSuccess(f.person.String("id"), "Splash1:1", f.version))
	if f.progress().Successes != 1 {
		t.Fatal(f.progress())
	}
	if _, e := f.s.GetAttempt(second); e == nil {
		t.Fatal("latest successful attempt was not removed")
	}
	if _, e := f.s.GetAttempt(first); e != nil {
		t.Fatal("older successful attempt was removed")
	}
}
func TestRepeatedDraftAndSubmit(t *testing.T) {
	f := setup(t)
	a := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	b := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	if a != b {
		t.Fatal("duplicate draft")
	}
	f.attempt(t, true, "")
	_ = must(f.s.Submit(a))
	if f.progress().Successes != 1 {
		t.Fatal(f.progress())
	}
}
func TestCorrectionCommitAndRestart(t *testing.T) {
	f := setup(t)
	a := f.attempt(t, true, "")
	f.attempt(t, true, "")
	f.attempt(t, true, "")
	check(t, f.s.EditAttempt(a))
	check(t, f.s.Mark(a, "Splash1:1:1", "not_yet"))
	if !f.progress().Complete {
		t.Fatal("draft affected progress")
	}
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	if !f.progress().Complete {
		t.Fatal("restart affected progress")
	}
	_ = must(f.s.Submit(a))
	if f.progress().Successes != 2 {
		t.Fatal(f.progress())
	}
	if len(must(f.s.queryRows("SELECT * FROM revisions"))) != 1 {
		t.Fatal("revision missing")
	}
	if len(must(f.s.Feedback(f.person.String("id")))) != 1 {
		t.Fatal("feedback missing")
	}
}
func TestFeedbackAcrossLessons(t *testing.T) {
	f := setup(t)
	if len(must(f.s.Feedback(f.person.String("id")))) != 0 {
		t.Fatal("untouched skills shown")
	}
	f.attempt(t, false, "")
	next := must(f.s.courseRecords("2026-09-11"))[0].String("lesson_id")
	f.attempt(t, true, next)
	fb := must(f.s.Feedback(f.person.String("id")))
	if len(fb) != 1 || fb[0].Successes != 1 || len(fb[0].Remaining) != 0 {
		t.Fatal(fb)
	}
	_ = must(f.s.NewAttempt(f.person.String("id"), next, "Splash1:1"))
	if !reflect.DeepEqual(fb, must(f.s.Feedback(f.person.String("id")))) {
		t.Fatal("draft appears in feedback")
	}
	f.attempt(t, true, next)
	f.attempt(t, true, next)
	if len(must(f.s.Feedback(f.person.String("id")))) != 0 {
		t.Fatal("complete skill shown")
	}
}
func TestOldCorrectionIsNotLatestObservation(t *testing.T) {
	f := setup(t)
	a := f.attempt(t, true, "")
	f.attempt(t, true, "")
	check(t, f.s.EditAttempt(a))
	check(t, f.s.Mark(a, "Splash1:1:1", "not_yet"))
	_ = must(f.s.Submit(a))
	if len(must(f.s.Feedback(f.person.String("id")))[0].Remaining) != 0 {
		t.Fatal("correction changed observation order")
	}
}
func TestReimportPreservesData(t *testing.T) {
	f := setup(t)
	f.attempt(t, true, "")
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	swimmers := f.p.Courses[0].Swimmers
	for i, j := 0, len(swimmers)-1; i < j; i, j = i+1, j-1 {
		swimmers[i], swimmers[j] = swimmers[j], swimmers[i]
	}
	f.p.Courses[0].Name = "Updated"
	_ = must(f.s.Import(f.p))
	if f.progress().Successes != 1 {
		t.Fatal(f.progress())
	}
	c := must(f.s.courseRecords("2026-09-04"))[0]
	if c.String("id") != f.course.String("id") {
		t.Fatal("course identity changed")
	}
	if must(f.s.Attendance(f.course.String("lesson_id")))[f.person.String("id")] != "present" {
		t.Fatal("attendance lost")
	}
}
func TestInvalidImportAndWriteFailureRollback(t *testing.T) {
	f := setup(t)
	f.p.Courses[0].Name = "Must not save"
	f.p.Courses[1].LessonDates = []string{"bad"}
	if _, e := f.s.Import(f.p); e == nil {
		t.Fatal("bad date accepted")
	}
	if must(f.s.courseRecords("2026-09-04"))[0].String("name") != f.course.String("name") {
		t.Fatal("partial import")
	}
	f.p.Courses[1].LessonDates = []string{"2026-09-04"}
	_, e := f.s.DB.Exec(`CREATE TEMP TRIGGER reject_swimmer BEFORE INSERT ON swimmers BEGIN SELECT RAISE(ABORT,'simulated write failure'); END`)
	check(t, e)
	if _, e = f.s.Import(f.p); e == nil {
		t.Fatal("failure missing")
	}
	if must(f.s.courseRecords("2026-09-04"))[0].String("name") != f.course.String("name") {
		t.Fatal("transaction did not roll back")
	}
}
func TestIdentityAndDatasetRejection(t *testing.T) {
	f := setup(t)
	f.p.Courses[0].Swimmers[0].Name = "Someone else"
	if _, e := f.s.Import(f.p); e == nil || !strings.Contains(e.Error(), "different swimmer") {
		t.Fatal(e)
	}
	f.p.DatasetID = "another"
	if _, e := f.s.Import(f.p); e == nil || !strings.Contains(e.Error(), "another dataset") {
		t.Fatal(e)
	}
}
func TestInactiveEnrollmentHistoryAndReactivation(t *testing.T) {
	f := setup(t)
	f.attempt(t, true, "")
	saved := f.p.Courses[0].Swimmers
	f.p.Courses[0].Swimmers = []transfer.SwimmerInput{}
	_ = must(f.s.Import(f.p))
	if must(f.s.enrollmentRecord(f.person.String("id"))).Int("active") != 0 || f.progress().Successes != 1 {
		t.Fatal("inactive history lost")
	}
	if _, e := f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"); e == nil {
		t.Fatal("inactive assessment accepted")
	}
	f.p.Courses[0].Swimmers = saved
	_ = must(f.s.Import(f.p))
	if must(f.s.enrollmentRecord(f.person.String("id"))).Int("active") != 1 {
		t.Fatal("reactivation failed")
	}
}
func TestCrossClassAndWrongLevelRejected(t *testing.T) {
	f := setup(t)
	other := must(f.s.courseRecords("2026-09-04"))[1]
	if e := f.s.SetAttendance(f.person.String("id"), other.String("lesson_id"), "present"); e == nil {
		t.Fatal("wrong class")
	}
	if _, e := f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash2B:1"); e == nil {
		t.Fatal("wrong level")
	}
}

func TestPastAttendanceDefaultsUncheckedStudentsAbsent(t *testing.T) {
	f := setup(t)
	roster := must(f.s.rosterRecords(f.course.String("id")))
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	check(t, f.s.MarkPastAttendanceAbsent(f.course.String("lesson_id"), "2026-09-05"))
	marks := must(f.s.Attendance(f.course.String("lesson_id")))
	for _, swimmer := range roster {
		if swimmer.Int("active") == 0 {
			continue
		}
		status := marks[swimmer.String("id")]
		if swimmer.String("id") == f.person.String("id") {
			if status != "present" {
				t.Fatalf("present student was changed to %q", status)
			}
			continue
		}
		if status != "absent" {
			t.Fatalf("unchecked student %s was marked %q", swimmer.String("id"), status)
		}
	}

	future := must(f.s.courseRecords("2026-09-11"))[0]
	check(t, f.s.SetAttendance(f.person.String("id"), future.String("lesson_id"), "present"))
	check(t, f.s.MarkPastAttendanceAbsent(future.String("lesson_id"), "2026-09-05"))
	futureMarks := must(f.s.Attendance(future.String("lesson_id")))
	if len(futureMarks) != 1 || futureMarks[f.person.String("id")] != "present" {
		t.Fatalf("future lesson attendance was finalized: %+v", futureMarks)
	}
}

func TestPrivateAndMissingLevel(t *testing.T) {
	f := setup(t)
	level := "SplashPrivate"
	f.p.Courses[2].Swimmers[0].Level = &level
	if _, e := f.s.Import(f.p); e == nil {
		t.Fatal("private reference accepted as curriculum")
	}
	level = "Mystery"
	_ = must(f.s.Import(f.p))
	if _, ok := must(f.s.CatalogFor(f.version)).Level(level); ok {
		t.Fatal("unknown level fell back")
	}
}
func TestExportAtomicityAndCompleteness(t *testing.T) {
	f := setup(t)
	a := f.attempt(t, true, "")
	check(t, f.s.EditAttempt(a))
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "absent"))
	path := filepath.Join(t.TempDir(), "results.json")
	check(t, f.s.Export(path))
	var snap transfer.Snapshot
	check(t, json.Unmarshal(must(os.ReadFile(path)), &snap))
	for _, table := range []string{"curricula", "attempts", "results", "drafts", "attendance", "enrollments"} {
		if len(snap.Tables[table]) == 0 {
			t.Fatal(table)
		}
	}
	if f.progress().Successes != 1 {
		t.Fatal("export cleared data")
	}
	if e := f.s.Export(f.path); e == nil {
		t.Fatal("database overwritten")
	}
	check(t, os.Symlink(f.path, path+".alias"))
	if e := f.s.Export(path + ".alias"); e == nil {
		t.Fatal("database alias overwritten")
	}
	before := must(os.ReadFile(path))
	if e := files.WriteJSON(path, func() {}); e == nil {
		t.Fatal("serialization failure missing")
	}
	if string(before) != string(must(os.ReadFile(path))) {
		t.Fatal("prior export replaced")
	}
}
func TestAbruptExitPreservesDraft(t *testing.T) {
	if path := os.Getenv("REC_TABLET_CRASH_DB"); path != "" {
		s := must(Open(path))
		c := must(s.courseRecords("2026-09-04"))[0]
		p := must(s.rosterRecords(c.String("id")))[0]
		if e := s.SetAttendance(p.String("id"), c.String("lesson_id"), "present"); e != nil {
			panic(e)
		}
		a := must(s.NewAttempt(p.String("id"), c.String("lesson_id"), "Splash1:1"))
		if e := s.Mark(a, "Splash1:1:1", "successful"); e != nil {
			panic(e)
		}
		os.Exit(0)
	}
	f := setup(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestAbruptExitPreservesDraft$")
	cmd.Env = append(os.Environ(), "REC_TABLET_CRASH_DB="+f.path)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("%s %v", out, e)
	}
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	if must(f.s.Attendance(f.course.String("lesson_id")))[f.person.String("id")] != "present" {
		t.Fatal("lost attendance")
	}
	a := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	if must(f.s.Draft(a))["Splash1:1:1"] != "successful" {
		t.Fatal("lost draft")
	}
	var integrity string
	check(t, f.s.DB.QueryRow("PRAGMA integrity_check").Scan(&integrity))
	if integrity != "ok" {
		t.Fatal(integrity)
	}
}
func TestCurriculumChangeResetsMarkingAndMovesCourses(t *testing.T) {
	f := setup(t)
	f.attempt(t, true, "")
	draft := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	check(t, f.s.Mark(draft, "Splash1:1:1", "successful"))
	check(t, f.s.SetAttendance(f.person.String("id"), f.course.String("lesson_id"), "present"))
	_, err := f.s.DB.Exec("INSERT INTO revisions(attempt_id,answers,saved_at) VALUES (?,?,?)", draft, "{}", Now())
	check(t, err)
	_, err = f.s.DB.Exec("INSERT INTO timed_attempts(attempt_id,laps_required,laps_completed,elapsed_ms) VALUES (?,?,?,?)", draft, 3, 1, 250)
	check(t, err)
	updated := curriculum.Bundled()
	updated.Version = "future-version"
	check(t, f.s.Close())
	f.s = must(OpenWithCatalog(f.path, updated))
	course := must(f.s.courseRecords("2026-09-04"))[0]
	if course.String("curriculum_version") != updated.Version {
		t.Fatalf("course still uses old curriculum version: %s", course.String("curriculum_version"))
	}
	for _, table := range []string{"attempts", "results", "drafts", "revisions", "timed_attempts"} {
		var count int
		check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&count))
		if count != 0 {
			t.Fatalf("catalog reset left %d rows in %s", count, table)
		}
	}
	if status := must(f.s.Attendance(course.String("lesson_id")))[f.person.String("id")]; status != "present" {
		t.Fatalf("catalog reset lost attendance: %q", status)
	}
	if progress := must(f.s.Progress(f.person.String("id"), "Splash1:1", updated.Version)); progress.Assessed != 0 || progress.Successes != 0 {
		t.Fatalf("catalog reset left progress: %+v", progress)
	}
}
func TestPythonDatabaseCompatibilityAfterCatalogReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "python.sqlite3")
	// Restore the exact Python-generated schema/data into an empty SQLite file.
	db := must(OpenRawFixture(path))
	check(t, db.Close())
	s := must(Open(path))
	defer s.Close()
	actual := must(s.Snapshot())
	for _, table := range []string{"attempts", "results", "drafts", "revisions"} {
		if len(actual.Tables[table]) != 0 {
			t.Fatalf("Python database reset left rows in %s", table)
		}
	}
	if len(actual.Tables["attendance"]) == 0 || len(actual.Tables["courses"]) == 0 || len(actual.Tables["enrollments"]) == 0 {
		t.Fatal("catalog reset lost roster or attendance data")
	}
	for _, course := range actual.Tables["courses"] {
		if course["curriculum_version"] != curriculum.Bundled().Version {
			t.Fatalf("course was not moved to the bundled curriculum: %v", course["curriculum_version"])
		}
	}
}
func TestValidationEdgeCases(t *testing.T) {
	base := must(os.ReadFile("../../../../examples/classes.json"))
	for _, replacement := range []string{`"schemaVersion": true`, `"schemaVersion": 1.0`} {
		bad := strings.Replace(string(base), `"schemaVersion": 1`, replacement, 1)
		if _, e := transfer.DecodePackage([]byte(bad)); e == nil {
			t.Fatal("invalid schema type accepted")
		}
	}
	for _, d := range []string{"2026-2-03", "2026-02-30", "0000-01-01", ""} {
		if transfer.Date(d) == nil {
			t.Fatal(d)
		}
	}
	f := setup(t)
	f.p.Courses[0].Swimmers = append(f.p.Courses[0].Swimmers, f.p.Courses[0].Swimmers[0])
	if _, e := f.s.Import(f.p); e == nil {
		t.Fatal("duplicate source ID accepted")
	}
	if _, e := transfer.DecodePackage([]byte(fmt.Sprintf(`{"schemaVersion":1,"courses":[{"location":null}]}`))); e == nil {
		t.Fatal("null location accepted")
	}
}

func TestDatabaseFilenameAndReconnectedSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lessons?#.sqlite3")
	s := must(Open(path))
	defer s.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatal("database filename was interpreted as URI parameters:", err)
	}
	// Force a new connection for every query to exercise reconnection settings.
	s.DB.SetMaxIdleConns(0)
	for pragma, expected := range map[string]int{"foreign_keys": 1, "synchronous": 2, "busy_timeout": 3000} {
		var actual int
		check(t, s.DB.QueryRow("PRAGMA "+pragma).Scan(&actual))
		if actual != expected {
			t.Fatalf("%s after reconnect = %d; want %d", pragma, actual, expected)
		}
	}
}
