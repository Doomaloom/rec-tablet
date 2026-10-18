package integration_test

import (
	"fmt"
	"testing"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
	"rectablet/src/workflows/swimming/transfer"
)

func TestLargeTimedSetupAndHeatPaging(t *testing.T) {
	f := setupUI(t)
	pkg := must(transfer.ReadPackage("../../../../examples/classes.json"))
	level := "Splash7"
	for i := range pkg.Courses[0].Swimmers {
		pkg.Courses[0].Swimmers[i].Level = &level
	}
	for i := 0; i < 9; i++ {
		pkg.Courses[0].Swimmers = append(pkg.Courses[0].Swimmers, transfer.SwimmerInput{SourceID: fmt.Sprintf("timed-large-%d", i), Name: fmt.Sprintf("Swimmer %02d", i), Level: &level})
	}
	_ = must(f.s.Import(pkg))
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	frame := f.c.Frame()
	course := must(f.s.GetCourse(frame.Course))
	f.c.Push(navigation.SkillSwimmers, func(n *navigation.Frame) { n.Skill, n.Level, n.Version = "Splash7:10", level, course.CurriculumVersion })
	check(t, f.c.Refresh())
	if !f.view.Buttons["timed-create-heat"].Disabled {
		t.Fatal("heat enabled without swimmers/laps")
	}
	people := append([]navigation.Row(nil), f.c.Snapshot().Rows...)
	for i := 0; i < 12; i++ {
		f.tap(t, "row:"+people[i].ID)
	}
	f.tap(t, "timed-setup-swimmers-page-1")
	f.tap(t, "row:"+people[12].ID)
	f.tap(t, "timed-digit:0")
	if !f.view.Buttons["timed-create-heat"].Disabled {
		t.Fatal("zero laps accepted")
	}
	f.tap(t, "timed-clear")
	f.tap(t, "timed-digit:2")
	f.shot(t, "large-timed-setup")
	f.tap(t, "timed-create-heat")
	if f.c.Snapshot().Screen != "timed_heat" || f.c.Snapshot().RowCount != 13 {
		t.Fatalf("heat = %s/%d", f.c.Snapshot().Screen, f.c.Snapshot().RowCount)
	}
	f.shot(t, "large-timed-heat")
	f.tap(t, "header-action-0")
	f.tap(t, "row:timed-lap:"+people[0].ID)
	f.tap(t, "timed-heat-swimmers-page-1")
	f.tap(t, "row:timed-lap:"+people[12].ID)
	f.tap(t, "timed-heat-swimmers-page--1")
	if f.c.Snapshot().Rows[0].ProgressValue != 1 || f.c.Snapshot().Rows[12].ProgressValue != 1 {
		t.Fatal("paging lost lap counts")
	}
	f.tap(t, "header-action-0")
	f.tap(t, "home")
	var saved int
	check(t, f.s.DB.QueryRow("SELECT COUNT(*) FROM timed_attempts").Scan(&saved))
	if saved != 13 {
		t.Fatalf("saved %d heat members, want all13 across pages", saved)
	}
}
