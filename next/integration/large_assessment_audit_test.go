package integration_test

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"

	"rectablet/src/shared/ui"
)

const longAssessmentSwimmer = "Alexandria Montgomery-Sutherland"

func assertLargeSwimmerLabelFits(t *testing.T, f *uiFixture, personID, name string) {
	t.Helper()
	button := f.view.Buttons["row:"+personID]
	if button == nil || !strings.Contains(button.Text, name) {
		t.Fatalf("swimmer button does not show roster label %q: %+v", name, button)
	}
	button.Refresh()
	for _, object := range test.WidgetRenderer(button).Objects() {
		if label, ok := object.(*ui.TextBlock); ok {
			if label.Truncated {
				t.Fatalf("swimmer label %q truncated at %.0f×%.0f", button.Text, button.Size().Width, button.Size().Height)
			}
			return
		}
	}
	t.Fatal("swimmer button has no rendered text")
}

func longAssessmentPerson(t *testing.T, f *uiFixture, courseID string) string {
	t.Helper()
	for _, person := range must(f.s.ListRoster(courseID)) {
		if person.Name == longAssessmentSwimmer {
			return person.ID
		}
	}
	t.Fatalf("course %s missing long-name swimmer", courseID)
	return ""
}

func TestCompactCompletedSwimmerCanUndoLatestSuccess(t *testing.T) {
	f := setupUI(t)
	courses := must(f.s.ListCourses("2026-09-04"))
	course := courses[0]
	people := must(f.s.ListRoster(course.ID))
	person := people[0]
	catalog := must(f.s.CatalogFor(course.CurriculumVersion))
	level, ok := catalog.Level(person.Level)
	if !ok || len(level.Skills) == 0 {
		t.Fatal("missing demonstration skill")
	}
	skill := level.Skills[0]
	for i := 0; i < 3; i++ {
		attempt := must(f.s.NewAttempt(person.ID, course.LessonID, skill.ID))
		for _, criterion := range skill.Subskills {
			check(t, f.s.Mark(attempt, criterion.ID, "successful"))
		}
		if !must(f.s.Submit(attempt)) {
			t.Fatal("all-successful attempt did not pass")
		}
	}
	if progress := must(f.s.Progress(person.ID, skill.ID, course.CurriculumVersion)); !progress.Complete {
		t.Fatal("skill did not reach three successful observations")
	}

	f.markAssessmentPresent(t)
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.tap(t, "row:"+skill.ID)
	button := f.view.Buttons["row:"+person.ID]
	if button == nil || button.Disabled || button.OnHold == nil {
		t.Fatal("completed swimmer must remain holdable for progress undo")
	}
	button.OnHold()
	if progress := must(f.s.Progress(person.ID, skill.ID, course.CurriculumVersion)); progress.Successes != 2 || progress.Complete {
		t.Fatalf("undo did not restore in-progress state: %+v", progress)
	}
}

// Large assessment enters a class batch directly from a skill. Single and Multi
// journeys are covered by the compact integration tests.
func TestLargeSkillClassSaveAndCancel(t *testing.T) {
	for _, save := range []bool{true, false} {
		t.Run(map[bool]string{true: "save", false: "cancel"}[save], func(t *testing.T) {
			f := setupUI(t)
			f.markAssessmentPresent(t)
			course := must(f.s.ListCourses("2026-09-04"))[0]
			people := must(f.s.ListRoster(course.ID))
			f.c.ConfigureDisplay(ui.DisplayLarge, nil)
			f.row(t, 0)
			f.tap(t, "row:skills")
			skillID := f.c.Snapshot().Rows[0].ID
			if len(f.c.Snapshot().HeaderSwitch) != 0 {
				t.Fatal("large skill picker offers obsolete mode switches")
			}
			f.tap(t, "row:"+skillID)
			if f.c.Snapshot().Screen != "class_assessment" {
				t.Fatalf("skill opened %s", f.c.Snapshot().Screen)
			}
			assertLargeSwimmerLabelFits(t, f, longAssessmentPerson(t, f, course.ID), "Alexandria M")
			f.shot(t, "large-audit-class")
			f.tap(t, "criterion:"+f.c.Snapshot().Rows[0].ID)
			action := "cancel-class"
			if save {
				action = "confirm-class"
			}
			f.tap(t, action)
			if f.c.Snapshot().Screen != "skills" {
				t.Fatalf("class exit returned to %s", f.c.Snapshot().Screen)
			}
			for _, person := range people {
				attempts := must(f.s.ListAttempts(person.ID, skillID, course.CurriculumVersion))
				if save && (len(attempts) != 1 || attempts[0].SubmittedAt == nil) {
					t.Fatalf("student %s was not saved: %+v", person.ID, attempts)
				}
				if !save && len(attempts) != 0 {
					t.Fatalf("cancelled student %s retained attempts: %+v", person.ID, attempts)
				}
			}
		})
	}
}

func TestLargeSkillSelectionBackAndCourseSwitchUsesNewRoster(t *testing.T) {
	f := setupUI(t)
	f.markAssessmentPresent(t)
	courses := must(f.s.ListCourses("2026-09-04"))
	firstPeople := must(f.s.ListRoster(courses[0].ID))
	secondPeople := must(f.s.ListRoster(courses[1].ID))
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.tap(t, "row:"+courses[0].ID)
	f.tap(t, "row:skills")
	rows := f.c.Snapshot().Rows
	if len(rows) < 7 {
		t.Fatal("need at least seven skills")
	}
	skillID := rows[6].ID
	f.tap(t, "row:"+skillID)
	if f.c.Snapshot().Screen != "class_assessment" || f.c.Frame().Skill != skillID {
		t.Fatal("skill did not open class assessment")
	}
	f.tap(t, "cancel-class")
	if button := f.view.Buttons["row:"+skillID]; button == nil || !button.Selected {
		t.Fatal("skill selection was not retained on return")
	}
	f.tap(t, "home")
	f.tap(t, "row:"+courses[1].ID)
	f.tap(t, "row:skills")
	f.row(t, 0)
	if f.c.Snapshot().Screen != "class_assessment" {
		t.Fatal("new course skill did not open class")
	}
	if f.view.Buttons["row:"+firstPeople[0].ID] != nil {
		t.Fatal("old course swimmer leaked into new roster")
	}
	if f.view.Buttons["row:"+secondPeople[0].ID] == nil {
		t.Fatal("new course swimmer missing")
	}
	f.tap(t, "row:"+secondPeople[0].ID)
	if f.c.Frame().Course != courses[1].ID {
		t.Fatal("assessment retained previous course")
	}
}
