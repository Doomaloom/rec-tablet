package repository

import (
	"encoding/json"
	"rectablet/src/workflows/swimming/transfer"
	"reflect"
	"testing"
)

func TestLessonPlanSnapshotsReplacePreserveAndReopen(t *testing.T) {
	f := setup(t)
	lesson := f.course.String("lesson_id")
	courseID := f.course.String("source_id")
	index := -1
	for i, c := range f.p.Courses {
		if c.SourceID == courseID {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("fixture course missing")
	}
	plans := []transfer.LessonPlan{{Date: "2026-09-04", CurriculumLevel: "Splash1", Rows: []transfer.PlannedRow{
		{Skill: "Entry", SkillID: "Splash1:1", Activity: "Jump", Location: "Lane", Duration: 5, Equipment: []string{"5 lb weights", "Custom aid"}},
		{Skill: "Entry", SkillID: "Splash1:1", Activity: "Repeat", Location: "Deep end", Duration: 10},
	}}}
	f.p.Courses[index].LessonPlans = &plans
	_ = must(f.s.Import(f.p))
	got := must(f.s.LessonPlan(lesson))
	if got == nil || !reflect.DeepEqual(*got, plans[0]) {
		t.Fatalf("plan = %+v", got)
	}
	attempt := f.attempt(t, true, lesson)
	// Older packages must preserve a cached plan and all saved assessment data.
	f.p.Courses[index].LessonPlans = nil
	_ = must(f.s.Import(f.p))
	if must(f.s.LessonPlan(lesson)) == nil {
		t.Fatal("old package cleared plan")
	}
	check(t, f.s.Close())
	f.s = must(Open(f.path))
	if must(f.s.LessonPlan(lesson)) == nil {
		t.Fatal("plan lost after restart")
	}
	_ = must(f.s.GetAttempt(attempt))
	// A malformed replacement must not replace the earlier plan or roster.
	invalid := []transfer.LessonPlan{{Date: "2099-01-01", Rows: []transfer.PlannedRow{}}}
	f.p.Courses[index].LessonPlans = &invalid
	if _, err := f.s.Import(f.p); err == nil {
		t.Fatal("accepted unscheduled plan")
	}
	if must(f.s.LessonPlan(lesson)) == nil {
		t.Fatal("invalid import cleared plan")
	}
	// Explicit empty snapshot clears plans, not marking data.
	empty := []transfer.LessonPlan{}
	f.p.Courses[index].LessonPlans = &empty
	_ = must(f.s.Import(f.p))
	if must(f.s.LessonPlan(lesson)) != nil {
		t.Fatal("empty snapshot did not clear plan")
	}
	_ = must(f.s.GetAttempt(attempt))
}

func TestLessonPlanValidation(t *testing.T) {
	base := transfer.ClassInput{LessonDates: []string{"2026-09-04"}}
	for _, raw := range []string{
		`[{"date":"2026-09-04","rows":[{"duration":0}]}]`,
		`[{"date":"2026-09-04","rows":[{"duration":5,"equipment":[" "]}]}]`,
		`[{"date":"2026-09-04","rows":[{"duration":5,"equipment":["Fins","fins"]}]}]`,
		`[{"date":"2026-09-04","rows":[{"duration":5,"activity":"Wrong","activities":[{"kind":"custom","text":"Right"}]}]}]`,
		`[{"date":"2026-09-04","rows":[]},{"date":"2026-09-04","rows":[]}]`,
	} {
		var plans []transfer.LessonPlan
		check(t, json.Unmarshal([]byte(raw), &plans))
		base.LessonPlans = &plans
		if err := transfer.ValidateLessonPlans(base); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
