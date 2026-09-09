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
	if _, e := f.s.Submit(a); e == nil {
		t.Fatal("incomplete attempt submitted")
	}
	if f.progress().Assessed != 0 {
		t.Fatal(f.progress())
	}
	check(t, f.s.Mark(a, "Splash1:1:2", "not_yet"))
	if must(f.s.Submit(a)) {
		t.Fatal("mixed attempt passed")
	}
	if p := f.progress(); p.Assessed != 1 || p.Successes != 0 {
		t.Fatal(p)
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
func TestCurriculumVersionPinned(t *testing.T) {
	f := setup(t)
	f.attempt(t, true, "")
	updated := curriculum.Bundled()
	updated.Version = "future-version"
	check(t, f.s.Close())
	f.s = must(OpenWithCatalog(f.path, updated))
	_ = must(f.s.Import(f.p))
	if must(f.s.courseRecords("2026-09-04"))[0].String("curriculum_version") != f.version || f.progress().Successes != 1 {
		t.Fatal("version changed")
	}
}
func TestPythonDatabaseCompatibility(t *testing.T) {
	path := filepath.Join(t.TempDir(), "python.sqlite3")
	// Restore the exact Python-generated schema/data into an empty SQLite file.
	db := must(OpenRawFixture(path))
	check(t, db.Close())
	s := must(Open(path))
	defer s.Close()
	var expected transfer.Snapshot
	check(t, json.Unmarshal(must(os.ReadFile("testdata/python-v1-export.json")), &expected))
	actual := must(s.Snapshot())
	// Go JSON normalization makes integer representations and embedded JSON
	// formatting comparable without weakening field/content compatibility.
	normalize := func(snap transfer.Snapshot) map[string]any {
		b := must(json.Marshal(snap.Tables))
		var m map[string]any
		check(t, json.Unmarshal(b, &m))
		for _, table := range []string{"curricula", "drafts", "revisions"} {
			key := "answers"
			if table == "curricula" {
				key = "body"
			}
			for _, row := range m[table].([]any) {
				r := row.(map[string]any)
				var value any
				check(t, json.Unmarshal([]byte(r[key].(string)), &value))
				r[key] = value
			}
		}
		return m
	}
	if !reflect.DeepEqual(normalize(actual), normalize(expected)) {
		t.Fatal("Go altered Python database or export semantics")
	}
	var ids map[string]string
	check(t, json.Unmarshal(must(os.ReadFile("testdata/identities.json")), &ids))
	p := must(s.Progress(ids["person"], "Splash1:1", ids["version"]))
	if p.Successes != 2 {
		t.Fatal(p)
	}
	if must(s.Draft(ids["draft"]))["Splash1:3:1"] != "successful" {
		t.Fatal("Python draft lost")
	}
	if must(s.Draft(ids["correction"]))["Splash1:1:2"] != "not_yet" {
		t.Fatal("Python correction lost")
	}
	_ = must(s.Submit(ids["correction"]))
	if must(s.Progress(ids["person"], "Splash1:1", ids["version"])).Successes != 1 {
		t.Fatal("cannot correct Python attempt")
	}
	check(t, s.Mark(ids["draft"], "Splash1:3:2", "not_yet"))
	if len(must(s.Draft(ids["draft"]))) != 2 {
		t.Fatal("cannot continue Python draft")
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
