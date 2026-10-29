package integration_test

import (
	"testing"

	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func TestAssessmentOnlyIncludesPresentStudents(t *testing.T) {
	for _, mode := range []string{"single", "class", "multi", "timed"} {
		t.Run(mode, func(t *testing.T) {
			f := setupUI(t)
			f.c.ConfigureDisplay(ui.DisplayLarge, nil)
			f.row(t, 0)
			frame := f.c.Frame()
			course := must(f.s.GetCourse(frame.Course))
			people := must(f.s.ListRoster(frame.Course))
			skill, level := "Splash1:1", "Splash1"
			if mode == "timed" {
				skill, level = "Splash7:10", "Splash7"
				for _, person := range people {
					_, err := f.s.DB.Exec("UPDATE enrollments SET level=? WHERE id=?", level, person.ID)
					check(t, err)
				}
			}
			// One present, one absent, and the remainder unmarked.
			check(t, f.s.SetAttendance(people[0].ID, frame.Lesson, "present"))
			check(t, f.s.SetAttendance(people[1].ID, frame.Lesson, "absent"))
			f.c.Push(navigation.SkillSwimmers, func(n *navigation.Frame) {
				n.Skill, n.Level, n.Version = skill, level, course.CurriculumVersion
			})
			check(t, f.c.Refresh())
			state := f.c.Snapshot()
			if state.RowCount != 1 || state.Rows[0].ID != people[0].ID {
				t.Fatalf("selection includes students who are not present: %+v", state.Rows)
			}
			switch mode {
			case "class":
				f.enterClass(t)
				if f.c.Snapshot().PageCount != 1 {
					t.Fatal("Class created attempts for students who are not present")
				}
			case "multi":
				f.tap(t, "header-switch-1")
				f.tap(t, "submit")
				if f.c.Snapshot().RowCount != 1 {
					t.Fatal("Multi included students who are not present")
				}
			case "single":
				f.tap(t, "row:"+people[0].ID)
				if f.view.Buttons["row:"+people[0].ID] == nil {
					t.Fatal("missing present student in marking column")
				}
				for _, person := range people[1:] {
					if f.view.Buttons["row:"+person.ID] != nil {
						t.Fatal("marking column includes student who is not present")
					}
				}
			case "timed":
				f.tap(t, "row:"+people[0].ID)
				f.tap(t, "timed-digit:1")
				f.tap(t, "timed-create-heat")
				if f.c.Snapshot().RowCount != 1 {
					t.Fatal("heat included students who are not present")
				}
			}
		})
	}
}
