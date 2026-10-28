package integration_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"fyne.io/fyne/v2"

	"rectablet/next/app"
	"rectablet/src/shared/ui"
)

func TestLargeFeedbackPaginationAndAttemptsRoundtrip(t *testing.T) {
	f := setupUI(t)
	courses := must(f.s.ListCourses("2026-09-04"))
	if len(courses) == 0 {
		t.Fatal("missing demonstration course")
	}
	course := courses[0]
	roster := must(f.s.ListRoster(course.ID))
	if len(roster) == 0 {
		t.Fatal("missing demonstration swimmer")
	}
	person := roster[0]
	catalog := must(f.s.CatalogFor(course.CurriculumVersion))
	level, ok := catalog.Level(person.Level)
	if !ok || len(level.Skills) < 4 {
		t.Fatal("need several skills to exercise pagination")
	}
	for _, skill := range level.Skills[:min(6, len(level.Skills))] {
		attempt := must(f.s.NewAttempt(person.ID, course.LessonID, skill.ID))
		_ = must(f.s.Submit(attempt))
	}
	entries := must(f.s.Feedback(person.ID))
	if len(entries) < 4 {
		t.Fatalf("created %d feedback entries, want at least four", len(entries))
	}

	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:feedback_swimmers")
	if f.c.Snapshot().Screen != "feedback_swimmers" || f.c.Frame().Person != person.ID || f.c.Snapshot().PageSize != 2 {
		t.Fatalf("large feedback did not select first swimmer: %+v", f.c.Frame())
	}
	if f.view.Buttons["row:"+person.ID] == nil || !f.view.Buttons["row:"+person.ID].Selected {
		t.Fatal("selected swimmer is not highlighted in the sidebar")
	}
	if f.c.Snapshot().PrimaryInfo != "" {
		t.Fatal("feedback must not reserve a level header")
	}
	composition := app.ComposeScreen(f.view, f.c.Snapshot())
	for _, component := range composition.Components {
		if component.Name == "feature-feedback-selected-swimmer" {
			t.Fatal("feedback must not show student name header")
		}
	}
	firstAction := "feedback-attempts:" + entries[0].Skill.ID + ":" + entries[0].Version
	if f.view.Buttons[firstAction] == nil {
		t.Fatalf("missing attempt action %q", firstAction)
	}

	_ = mountedFeedbackCards(t, f)
	if got := f.view.Buttons[firstAction].Size().Height; got < 48 {
		t.Fatalf("attempt action height %.1f, want at least 48", got)
	}
	f.tap(t, "down")
	secondAction := "feedback-attempts:" + entries[2].Skill.ID + ":" + entries[2].Version
	if f.c.Snapshot().Page != 1 || f.view.Buttons[secondAction] == nil || f.view.Buttons[firstAction] != nil {
		t.Fatal("rail did not advance feedback")
	}
	f.tap(t, "feedback-filter:Complete")
	if f.c.Snapshot().Page != 0 || f.view.Buttons[firstAction] != nil || !f.view.Buttons["feedback-filter:Complete"].Selected {
		t.Fatal("complete switch did not reset page and filter feedback")
	}
	f.tap(t, "feedback-filter:In progress")
	if f.view.Buttons[firstAction] == nil {
		t.Fatal("in progress switch did not restore feedback")
	}
	f.tap(t, "up")

	subskillsAction := "feedback-subskills:" + entries[0].Skill.ID + ":" + entries[0].Version
	f.tap(t, subskillsAction)
	if f.c.Snapshot().Screen != "feedback_subskills" || f.c.Frame().Person != person.ID || f.c.Frame().Skill != entries[0].Skill.ID || f.c.Frame().Version != entries[0].Version {
		t.Fatal("subskills did not navigate with pinned student and skill")
	}
	f.shot(t, "large-feedback-subskills")
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "feedback_swimmers" || f.view.Buttons[firstAction] == nil {
		t.Fatal("back did not restore feedback")
	}
	f.tap(t, firstAction)
	if f.c.Snapshot().Screen != "attempts" || f.c.Frame().Person != person.ID || f.c.Frame().Skill != entries[0].Skill.ID || f.c.Frame().Version != entries[0].Version {
		t.Fatalf("attempts lost selected context: %+v", f.c.Frame())
	}
	f.shot(t, "large-feedback-attempts")
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "feedback_swimmers" || f.c.Frame().Person != person.ID {
		t.Fatalf("attempts did not return to selected feedback: %+v", f.c.Frame())
	}

	f.w.Resize(fyne.NewSize(1536, 900))
	f.view.Root.Refresh()
	_ = mountedFeedbackCards(t, f)
	if got := f.view.Buttons[firstAction].Size().Height; got < 72 {
		t.Fatalf("scaled attempt action height %.1f, want at least 72", got)
	}
	f.shot(t, "large-feedback-1536")
}

func mountedFeedbackCards(t *testing.T, f *uiFixture) fyne.CanvasObject {
	t.Helper()
	composition := app.ComposeScreen(f.view, f.c.Snapshot())
	for i, component := range composition.Components {
		if component.Name != "feature-feedback-skill-cards" {
			continue
		}
		object := f.view.Root.Objects[i]
		field := reflect.ValueOf(object).Elem().FieldByName("Scroll")
		if field.IsValid() {
			t.Fatal("feedback must use fixed cards without scrolling")
		}

		logicalX, logicalY, logicalW, logicalH := component.Position(1024, 600)
		scale := min(f.w.Canvas().Size().Width/1024, f.w.Canvas().Size().Height/600)
		for label, got := range map[string][2]float32{
			"x": {object.Position().X, logicalX * scale}, "y": {object.Position().Y, logicalY * scale},
			"width": {object.Size().Width, logicalW * scale}, "height": {object.Size().Height, logicalH * scale},
		} {
			if math.Abs(float64(got[0]-got[1])) > 1 {
				t.Fatalf("%s at scale %.2f = %.1f, want %.1f", label, scale, got[0], got[1])
			}
		}
		return object
	}
	t.Fatal(fmt.Sprintf("feedback cards missing from %d components", len(composition.Components)))
	return nil
}
