package integration_test

import (
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/transfer"
	"strings"
	"testing"
)

func TestLargeLessonPlanDetailsMarkingAndReturn(t *testing.T) {
	f := setupUI(t)
	pkg := must(transfer.ReadPackage("../../examples/classes.json"))
	rows := []transfer.PlannedRow{}
	for i := 0; i < 8; i++ {
		rows = append(rows, transfer.PlannedRow{Skill: "Enter and Exit Shallow Water", SkillID: "Splash1:1", Activity: "Jump safely with a partner. " + strings.Repeat("Practice. ", 80), Location: "Shallow end", Duration: 5, Equipment: []string{"Kickboards", "5 lb weights"}})
	}
	rows[7].Activity = "Practise entry with feedback."
	rows[1].SkillID = "Splash1:6"
	rows[1].Skill = "Float on front and back"
	plans := []transfer.LessonPlan{{Date: "2026-09-04", CurriculumLevel: "Splash1", Rows: rows}}
	pkg.Courses[0].LessonPlans = &plans
	_ = must(f.s.Import(pkg))
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	menu := f.c.Snapshot().Rows
	if len(menu) != 4 || menu[0].ID != "attendance" || menu[1].ID != "lesson_plan" || menu[2].ID != "skills" {
		t.Fatalf("menu: %+v", menu)
	}
	f.tap(t, "row:lesson_plan")
	if f.c.Snapshot().RowCount != 8 || f.view.Buttons["lesson-plan-start-marking"] != nil {
		t.Fatal("plan must start with full list")
	}
	f.shot(t, "lesson-plan-list")
	f.tap(t, "down")
	f.tap(t, "row:6")
	if f.c.Snapshot().Screen != "lesson_plan_detail" || f.view.Buttons["lesson-plan-start-marking"].Disabled {
		t.Fatalf("detail: %+v", f.c.Snapshot())
	}
	f.shot(t, "lesson-plan-details")
	f.tap(t, "row:7")
	f.shot(t, "lesson-plan-details-short")
	f.tap(t, "row:6")
	f.tap(t, "lesson-plan-start-marking")
	if f.c.Snapshot().Screen != "class_assessment" {
		t.Fatalf("marking: %+v", f.c.Snapshot())
	}
	f.tap(t, "cancel-class")
	if f.c.Snapshot().Screen != "lesson_plan_detail" || f.c.Frame().PlanRow != "6" || f.c.Snapshot().Page != 1 {
		t.Fatal("marking did not return to selected plan row/page")
	}
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "lesson_plan" || f.c.Snapshot().Page != 1 {
		t.Fatal("back did not restore full list page")
	}
	f.tap(t, "up")
	f.tap(t, "row:1")
	f.shot(t, "lesson-plan-components")
	f.tap(t, "lesson-plan-component:Splash1:6b")
	f.tap(t, "lesson-plan-start-marking")
	if f.c.Frame().Skill != "Splash1:6b" {
		t.Fatal("wrong marking component")
	}
	f.tap(t, "confirm-class")
	if f.c.Snapshot().Screen != "lesson_plan_detail" {
		t.Fatal("assessment save lost plan")
	}
}

func TestLargeLessonPlanEmptyUnmappedAndTimed(t *testing.T) {
	f := setupUI(t)
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:lesson_plan")
	if f.c.Snapshot().RowCount != 0 || !strings.Contains(f.c.Snapshot().Empty, "No lesson plan") {
		t.Fatal("missing empty state")
	}
	f.tap(t, "back")
	pkg := must(transfer.ReadPackage("../../examples/classes.json"))
	level := "Splash7"
	for i := range pkg.Courses[0].Swimmers {
		pkg.Courses[0].Swimmers[i].Level = &level
	}
	plans := []transfer.LessonPlan{{Date: "2026-09-04", CurriculumLevel: level, Rows: []transfer.PlannedRow{
		{Skill: "Custom game", Activity: "Play", Location: "Lane", Duration: 5},
		{Skill: "Timed swim", SkillID: "Splash7:10", Activity: "Timed swim", Location: "Lane", Duration: 10},
	}}}
	pkg.Courses[0].LessonPlans = &plans
	_ = must(f.s.Import(pkg))
	check(t, f.c.Refresh())
	f.tap(t, "row:lesson_plan")
	f.tap(t, "row:0")
	if !f.view.Buttons["lesson-plan-start-marking"].Disabled {
		t.Fatal("unmapped activity can mark")
	}
	f.shot(t, "lesson-plan-unmapped")
	f.tap(t, "row:1")
	f.tap(t, "lesson-plan-start-marking")
	if f.c.Snapshot().Screen != "skill_swimmers" || f.c.Frame().Skill != "Splash7:10" || f.view.Buttons["timed-create-heat"] == nil {
		t.Fatalf("timed plan did not open heat setup: %+v", f.c.Snapshot())
	}
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "lesson_plan_detail" || f.c.Frame().PlanRow != "1" {
		t.Fatal("timed back lost plan")
	}
}
