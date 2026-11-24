package repository

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/transfer"
)

// oldFixture is a tablet that was last opened with the previous catalog.
func oldFixture(t *testing.T) *fixture {
	t.Helper()
	path := filepath.Join(t.TempDir(), "lesson.sqlite3")
	s := must(OpenWithCatalog(path, curriculum.BundledPrevious()))
	p := must(transfer.ReadPackage("../../../../examples/classes.json"))
	_ = must(s.Import(p))
	c := must(s.courseRecords("2026-09-04"))[0]
	person := must(s.rosterRecords(c.String("id")))[0]
	f := &fixture{s, p, c, person, c.String("curriculum_version"), path}
	t.Cleanup(func() { f.s.Close() })
	if f.version != curriculum.BundledPrevious().Version {
		t.Fatalf("fixture course is not on the previous version: %s", f.version)
	}
	return f
}

// markOld submits an attempt of an old-version skill; failing lists old
// subskill ids to mark not_yet, everything else is successful.
func (f *fixture) markOld(t *testing.T, skill string, failing ...string) string {
	t.Helper()
	a := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), skill))
	for _, sub := range must(f.s.Skill(curriculum.BundledPrevious().Version, skill)).Subskills {
		status := "successful"
		for _, bad := range failing {
			if bad == sub.ID {
				status = "not_yet"
			}
		}
		check(t, f.s.Mark(a, sub.ID, status))
	}
	_ = must(f.s.Submit(a))
	return a
}

func (f *fixture) upgrade(t *testing.T) {
	t.Helper()
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	f.version = curriculum.Bundled().Version
}

var markingTables = []string{"attempts", "results", "drafts", "revisions", "timed_attempts"}

func TestUpgradeFromOldVersionKeepsEveryAttemptAndDraft(t *testing.T) {
	f := oldFixture(t)
	person, lesson := f.person.String("id"), f.course.String("lesson_id")
	oldVersion := f.version
	done := f.markOld(t, "Splash1:1")
	split := f.markOld(t, "Splash1:7")
	draft := must(f.s.NewAttempt(person, lesson, "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	check(t, f.s.SetAttendance(person, lesson, "present"))
	_, err := f.s.DB.Exec("INSERT INTO revisions(attempt_id,answers,saved_at) VALUES (?,?,?)", done, "{}", Now())
	check(t, err)
	_, err = f.s.DB.Exec("INSERT INTO timed_attempts(attempt_id,laps_required,laps_completed,elapsed_ms) VALUES (?,?,?,?)", draft, 3, 1, 250)
	check(t, err)
	before := must(f.s.Snapshot())

	f.upgrade(t)
	after := must(f.s.Snapshot())
	for _, table := range markingTables {
		if len(before.Tables[table]) == 0 && table != "timed_attempts" {
			t.Fatalf("fixture has no %s", table)
		}
		if table == "timed_attempts" {
			if count(t, f.s, table) != 1 {
				t.Fatal("timed row lost")
			}
			continue
		}
		if !reflect.DeepEqual(before.Tables[table], after.Tables[table]) {
			t.Fatalf("upgrade changed %s", table)
		}
	}
	for _, row := range after.Tables["attempts"] {
		if row["curriculum_version"] != oldVersion {
			t.Fatalf("attempt was converted: %v", row)
		}
	}
	if status := must(f.s.Attendance(lesson))[person]; status != "present" {
		t.Fatal("attendance lost")
	}
	// Courses move to the new version; both catalogs are registered.
	for _, c := range must(f.s.courseRecords("2026-09-04")) {
		if c.String("curriculum_version") != curriculum.Bundled().Version {
			t.Fatalf("course not moved: %v", c)
		}
	}
	for _, v := range []string{curriculum.Bundled().Version, oldVersion} {
		if _, err := f.s.CatalogFor(v); err != nil {
			t.Fatalf("catalog %s not registered: %v", v, err)
		}
	}
	// The open old draft is still submittable under the old catalog.
	if !must(f.s.Submit(draft)) {
		t.Fatal("old draft did not submit successfully")
	}
	row := must(f.s.queryOne("SELECT curriculum_version,skill_id FROM attempts WHERE id=?", draft))
	if row.String("curriculum_version") != oldVersion || row.String("skill_id") != "Splash1:2" {
		t.Fatalf("submitted draft changed version: %v", row)
	}
	// A saved old attempt can be corrected in its own version.
	check(t, f.s.EditAttempt(split))
	check(t, f.s.Mark(split, "Splash1:7:2", "not_yet"))
	if must(f.s.Submit(split)) {
		t.Fatal("corrected attempt still successful")
	}
	if v := must(f.s.queryOne("SELECT curriculum_version FROM attempts WHERE id=?", split)).String("curriculum_version"); v != oldVersion {
		t.Fatal("correction converted the attempt:", v)
	}
	// An upgrade is idempotent.
	f.upgrade(t)
	if count(t, f.s, "attempts") != 3 {
		t.Fatal("reopen changed attempts")
	}
}

func TestUpgradeWithSyncGrantMovesCoursesAndKeepsData(t *testing.T) {
	f := oldFixture(t)
	f.markOld(t, "Splash1:1")
	draft := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:2"))
	check(t, f.s.Mark(draft, "Splash1:2:1", "successful"))
	bindFixture(t, f, "http://localhost")
	_, err := f.s.DB.Exec("DELETE FROM sync_outbox")
	check(t, err)
	f.upgrade(t)
	if count(t, f.s, "attempts") != 2 || count(t, f.s, "drafts") != 1 {
		t.Fatal("bound upgrade lost marking data")
	}
	if v := must(f.s.queryOne("SELECT curriculum_version FROM courses LIMIT 1")).String("curriculum_version"); v != curriculum.Bundled().Version {
		t.Fatal("bound course not moved", v)
	}
	for _, o := range queue(t, f.s) {
		if o.Kind == "attempt" {
			t.Fatalf("old attempt requeued: %+v", o)
		}
	}
}

func TestProgressCountsTranslatedOldAttempts(t *testing.T) {
	f := oldFixture(t)
	person := f.person.String("id")
	f.markOld(t, "Splash1:7")                // all successful
	f.markOld(t, "Splash1:7", "Splash1:7:2") // misses one shared criterion
	f.markOld(t, "Splash1:1")                // unsplit skill
	f.markOld(t, "Splash1:10-back-crawl-wearing-pfd-5-m")
	f.upgrade(t)
	version := f.version
	for _, id := range []string{"Splash1:7-front-to-back", "Splash1:7-back-to-front"} {
		if p := must(f.s.Progress(person, id, version)); p.Assessed != 2 || p.Successes != 1 || p.Complete {
			t.Fatalf("%s: %+v", id, p)
		}
	}
	if p := must(f.s.Progress(person, "Splash1:1", version)); p.Assessed != 1 || p.Successes != 1 {
		t.Fatalf("%+v", p)
	}
	// A box with no old criterion follows the old attempt's overall success.
	if p := must(f.s.Progress(person, "Splash1:10-back-crawl-wearing-pfd-5-m", version)); p.Assessed != 1 || p.Successes != 1 {
		t.Fatalf("fromSkillSuccess box: %+v", p)
	}
	// Old-version counting does not leak into other skills or old-version queries.
	if p := must(f.s.Progress(person, "Splash1:2", version)); p.Assessed != 0 {
		t.Fatalf("%+v", p)
	}
	// New marking is added to the translated history, in the new version.
	lesson := f.course.String("lesson_id")
	a := must(f.s.NewAttempt(person, lesson, "Splash1:7-front-to-back"))
	for _, sub := range must(f.s.Skill(version, "Splash1:7-front-to-back")).Subskills {
		check(t, f.s.Mark(a, sub.ID, "successful"))
	}
	if !must(f.s.Submit(a)) {
		t.Fatal("new attempt failed")
	}
	if p := must(f.s.Progress(person, "Splash1:7-front-to-back", version)); p.Assessed != 3 || p.Successes != 2 {
		t.Fatalf("%+v", p)
	}
	if p := must(f.s.Progress(person, "Splash1:7-back-to-front", version)); p.Assessed != 2 {
		t.Fatalf("new attempt leaked into its sibling: %+v", p)
	}
	if v := must(f.s.queryOne("SELECT curriculum_version FROM attempts WHERE id=?", a)).String("curriculum_version"); v != version {
		t.Fatal("new marking not in the new version", v)
	}
}

func TestFromSkillSuccessBoxFailsWhenOldAttemptFailed(t *testing.T) {
	f := oldFixture(t)
	person := f.person.String("id")
	skill := "Splash1:10-back-crawl-wearing-pfd-5-m"
	f.markOld(t, skill, skill+":1")
	f.upgrade(t)
	if p := must(f.s.Progress(person, skill, f.version)); p.Assessed != 1 || p.Successes != 0 {
		t.Fatalf("%+v", p)
	}
}

func TestTranslatedAttemptsCompleteASkillAndAreListed(t *testing.T) {
	f := oldFixture(t)
	person := f.person.String("id")
	var ids []string
	for i := 0; i < 3; i++ {
		ids = append(ids, f.markOld(t, "Splash1:1"))
	}
	f.upgrade(t)
	if p := must(f.s.Progress(person, "Splash1:1", f.version)); !p.Complete || p.Successes != 3 {
		t.Fatalf("%+v", p)
	}
	list := must(f.s.ListAttempts(person, "Splash1:1", f.version))
	if len(list) != 3 {
		t.Fatalf("old attempts missing from the list: %d", len(list))
	}
	for _, a := range list {
		if a.CurriculumVersion != curriculum.BundledPrevious().Version || a.SkillID != "Splash1:1" {
			t.Fatalf("listed attempt was converted: %+v", a)
		}
	}
	// Both halves of a split skill list the one old attempt.
	f2 := oldFixture(t)
	f2.markOld(t, "Splash1:7")
	f2.upgrade(t)
	for _, id := range []string{"Splash1:7-front-to-back", "Splash1:7-back-to-front"} {
		if got := must(f2.s.ListAttempts(f2.person.String("id"), id, f2.version)); len(got) != 1 || got[0].SkillID != "Splash1:7" {
			t.Fatalf("%s: %+v", id, got)
		}
	}
	// Open the old attempt for review and correct it in its own version.
	check(t, f.s.EditAttempt(ids[0]))
	check(t, f.s.Mark(ids[0], "Splash1:1:1", "not_yet"))
	_ = must(f.s.Submit(ids[0]))
	if p := must(f.s.Progress(person, "Splash1:1", f.version)); p.Successes != 2 || p.Complete {
		t.Fatalf("correction not reflected: %+v", p)
	}
}

func TestFeedbackCountsMissesAcrossTranslatedAttempts(t *testing.T) {
	f := oldFixture(t)
	person := f.person.String("id")
	f.markOld(t, "Splash1:7", "Splash1:7:2")
	f.markOld(t, "Splash1:7", "Splash1:7:2")
	f.markOld(t, "Splash1:1")
	f.upgrade(t)
	byID := map[string]int{}
	for _, entry := range must(f.s.Feedback(person)) {
		byID[entry.Skill.ID] = entry.Assessed
		if entry.Version != f.version {
			t.Fatalf("%s reported under %s", entry.Skill.ID, entry.Version)
		}
		if strings.HasPrefix(entry.Skill.ID, "Splash1:7-") {
			if len(entry.MissedSubskills) != 1 || entry.MissedSubskills[0].Subskill.ID != entry.Skill.ID+":2" || entry.MissedSubskills[0].Count != 2 {
				t.Fatalf("%s misses: %+v", entry.Skill.ID, entry.MissedSubskills)
			}
			if len(entry.Remaining) != 1 || entry.Remaining[0].ID != entry.Skill.ID+":2" {
				t.Fatalf("%s remaining: %+v", entry.Skill.ID, entry.Remaining)
			}
		}
	}
	if byID["Splash1:7-front-to-back"] != 2 || byID["Splash1:7-back-to-front"] != 2 || byID["Splash1:1"] != 1 || len(byID) != 3 {
		t.Fatalf("%v", byID)
	}
	if got := must(f.s.CompletedFeedback(person)); len(got) != 0 {
		t.Fatalf("%+v", got)
	}
}

func TestRemoveLatestSuccessCanRemoveTranslatedOldSuccess(t *testing.T) {
	f := oldFixture(t)
	person := f.person.String("id")
	old := f.markOld(t, "Splash1:1")
	f.upgrade(t)
	check(t, f.s.RemoveLatestSuccess(person, "Splash1:1", f.version))
	if _, err := f.s.queryOne("SELECT id FROM attempts WHERE id=?", old); err == nil {
		t.Fatal("old success not removed")
	}
	if err := f.s.RemoveLatestSuccess(person, "Splash1:1", f.version); err == nil {
		t.Fatal("expected nothing left to remove")
	}
}

// Server history in an older registered version is stored as it is, in that
// version, for a course already on the new version.
func TestSyncReadAcceptsOldVersionRecordForNewVersionCourse(t *testing.T) {
	f := setup(t)
	server := newSyncServer(t)
	p := bindFixture(t, f, server.URL)
	person := f.person.String("id")
	attempt := f.attempt(t, true, "")
	op := queue(t, f.s)[0]
	if op.RecordID != attempt || op.CurriculumVersion != curriculum.Bundled().Version {
		t.Fatalf("new marking not recorded in the new version: %+v", op)
	}
	// The server later returns that record as it was originally saved by an
	// older tablet: previous version, previous skill and subskills.
	check(t, f.s.RemoveLatestSuccess(person, "Splash1:1", f.version))
	_, e := f.s.DB.Exec("DELETE FROM sync_outbox; DELETE FROM sync_state WHERE record_id=?", attempt)
	check(t, e)
	var payload attemptPayload
	check(t, json.Unmarshal(op.Payload, &payload))
	payload.Results = payload.Results[:0]
	for _, sub := range must(curriculum.BundledPrevious().Skill("Splash1:1")).Subskills {
		payload.Results = append(payload.Results, struct {
			SubskillID string `json:"subskillId"`
			Status     string `json:"status"`
		}{sub.ID, "successful"})
	}
	op.CurriculumVersion = curriculum.BundledPrevious().Version
	op.Payload, _ = json.Marshal(payload)
	server.pull = func() map[string]any {
		return map[string]any{"schemaVersion": 1, "records": []syncRecord{{SyncOperation: op, Revision: 1, Success: true}}}
	}
	check(t, f.s.ReadAccepted(context.Background(), p.Credentials))
	row := must(f.s.queryOne("SELECT curriculum_version,skill_id,success FROM attempts WHERE id=?", attempt))
	if row.String("curriculum_version") != curriculum.BundledPrevious().Version || row.Int("success") != 1 {
		t.Fatalf("old record not stored in its own version: %v", row)
	}
	if len(queue(t, f.s)) != 0 {
		t.Fatal("pulled record was echoed back")
	}
	if p := must(f.s.Progress(person, "Splash1:1", f.version)); p.Assessed != 1 || p.Successes != 1 {
		t.Fatalf("pulled old record not counted: %+v", p)
	}
	// A version this tablet has no catalog for is still refused.
	op.CurriculumVersion = "version-of-another-device"
	op.RecordID = "unknown-version-attempt"
	server.pull = func() map[string]any {
		return map[string]any{"schemaVersion": 1, "records": []syncRecord{{SyncOperation: op, Revision: 1}}}
	}
	if err := f.s.ReadAccepted(context.Background(), p.Credentials); err == nil || !strings.Contains(err.Error(), "curriculum") {
		t.Fatal("unknown curriculum accepted:", err)
	}
}

func TestPushKeepsOldAttemptVersion(t *testing.T) {
	f := oldFixture(t)
	f.markOld(t, "Splash1:1")
	bindFixture(t, f, "http://localhost")
	f.upgrade(t)
	found := false
	for _, o := range queue(t, f.s) {
		if o.Kind != "attempt" {
			continue
		}
		found = true
		if o.CurriculumVersion != curriculum.BundledPrevious().Version {
			t.Fatalf("old attempt pushed with the wrong version: %+v", o)
		}
	}
	if !found {
		t.Fatal("bound old attempt missing from the queue")
	}
}
