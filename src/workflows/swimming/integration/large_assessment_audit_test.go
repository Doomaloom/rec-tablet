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
		t.Fatalf("swimmer button does not show full name %q: %+v", name, button)
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

func largeModeSwitch(t *testing.T, f *uiFixture, label string) string {
	t.Helper()
	for i, option := range f.c.Snapshot().HeaderSwitch {
		if option.Label == label {
			return "header-switch-" + string(rune('0'+i))
		}
	}
	t.Fatalf("missing %s mode switch", label)
	return ""
}

func TestLargeCompletedSwimmerCanUndoLatestSuccess(t *testing.T) {
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

	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.shot(t, "large-audit-skills-complete")
	assertLargeSwimmerLabelFits(t, f, longAssessmentPerson(t, f, course.ID), longAssessmentSwimmer)
	button := f.view.Buttons["row:"+person.ID]
	if button == nil || button.Disabled || button.OnHold == nil {
		t.Fatal("completed swimmer must remain holdable for progress undo")
	}
	button.OnHold()
	if progress := must(f.s.Progress(person.ID, skill.ID, course.CurriculumVersion)); progress.Successes != 2 || progress.Complete {
		t.Fatalf("undo did not restore in-progress state: %+v", progress)
	}
}

func TestLargeSingleDraftSwitchAndCorrectionRoundtrip(t *testing.T) {
	f := setupUI(t)
	courses := must(f.s.ListCourses("2026-09-04"))
	course := courses[0]
	people := must(f.s.ListRoster(course.ID))
	if len(people) < 2 {
		t.Fatal("need two demonstration swimmers")
	}
	first, second := people[0], people[1]
	catalog := must(f.s.CatalogFor(course.CurriculumVersion))
	level, ok := catalog.Level(first.Level)
	if !ok || len(level.Skills) == 0 {
		t.Fatal("missing demonstration skill")
	}
	skill := level.Skills[0]

	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.shot(t, "large-audit-skills")
	assertLargeSwimmerLabelFits(t, f, longAssessmentPerson(t, f, course.ID), longAssessmentSwimmer)
	f.tap(t, "row:"+first.ID)
	if f.c.Snapshot().Screen != "assessment" {
		t.Fatalf("first swimmer opened %s", f.c.Snapshot().Screen)
	}
	f.shot(t, "large-audit-single")
	assertLargeSwimmerLabelFits(t, f, longAssessmentPerson(t, f, course.ID), longAssessmentSwimmer)
	firstAttempt := f.c.Frame().Attempt
	criterionID := f.c.Snapshot().Rows[0].ID
	f.tap(t, "criterion:"+criterionID)
	f.tap(t, "row:"+second.ID)
	if f.c.Frame().Person != second.ID || f.c.Frame().Attempt == firstAttempt {
		t.Fatal("sidebar did not switch to another swimmer draft")
	}
	if draft := must(f.s.Draft(firstAttempt)); draft[criterionID] != "successful" {
		t.Fatalf("first draft lost mark: %+v", draft)
	}
	f.tap(t, "row:"+first.ID)
	if f.c.Frame().Attempt != firstAttempt || f.c.Snapshot().Rows[0].Status != "successful" {
		t.Fatal("returning to swimmer did not restore saved draft")
	}
	f.tap(t, "home") // Submit this incomplete attempt.
	if f.c.Snapshot().Screen != "skills" {
		t.Fatalf("save returned to %s", f.c.Snapshot().Screen)
	}
	f.tap(t, "back")
	f.tap(t, "row:feedback_swimmers")
	if f.c.Frame().Person != first.ID {
		t.Fatal("feedback did not select the first swimmer")
	}
	entries := must(f.s.Feedback(first.ID))
	if len(entries) != 1 {
		t.Fatalf("saved attempt produced %d feedback entries", len(entries))
	}
	f.tap(t, "feedback-attempts:"+skill.ID+":"+entries[0].Version)
	if f.c.Snapshot().Screen != "attempts" {
		t.Fatalf("attempt action opened %s", f.c.Snapshot().Screen)
	}
	f.row(t, 0)
	if f.c.Snapshot().Screen != "assessment" || !f.c.Frame().Review {
		t.Fatal("attempt did not open for correction")
	}
	for _, row := range append([]ui.Item(nil), f.c.Snapshot().Rows...) {
		if row.Status != "successful" {
			f.tap(t, "criterion:"+row.ID)
		}
	}
	if !f.c.Snapshot().CanSubmit {
		t.Fatal("complete correction cannot be submitted")
	}
	f.tap(t, "submit")
	if f.c.Snapshot().Screen != "attempts" {
		t.Fatalf("correction returned to %s", f.c.Snapshot().Screen)
	}
	if progress := must(f.s.Progress(first.ID, skill.ID, course.CurriculumVersion)); progress.Successes != 1 {
		t.Fatalf("correction did not update progress: %+v", progress)
	}
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "feedback_swimmers" || f.c.Frame().Person != first.ID {
		t.Fatal("correction lost selected feedback swimmer")
	}
}

func TestLargeClassMarksAndKeepsSavedStudentWhenDiscardingRemainder(t *testing.T) {
	f := setupUI(t)
	course := must(f.s.ListCourses("2026-09-04"))[0]
	people := must(f.s.ListRoster(course.ID))
	if len(people) < 2 {
		t.Fatal("need two demonstration swimmers")
	}
	catalog := must(f.s.CatalogFor(course.CurriculumVersion))
	level, ok := catalog.Level(people[0].Level)
	if !ok || len(level.Skills) == 0 {
		t.Fatal("missing demonstration skill")
	}
	skill := level.Skills[0]
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.tap(t, largeModeSwitch(t, f, "Class"))
	if f.c.Snapshot().Screen != "class_assessment" {
		t.Fatalf("class switch opened %s", f.c.Snapshot().Screen)
	}
	f.shot(t, "large-audit-class")
	longPersonID := longAssessmentPerson(t, f, course.ID)
	assertLargeSwimmerLabelFits(t, f, longPersonID, longAssessmentSwimmer)
	if f.c.Snapshot().Title != longAssessmentSwimmer {
		f.tap(t, "row:"+longPersonID)
	}
	current := f.c.Snapshot().Title
	criterionID := f.c.Snapshot().Rows[0].ID
	f.tap(t, "criterion:"+criterionID)
	f.tap(t, "save-student")
	assertLargeSwimmerLabelFits(t, f, longPersonID, longAssessmentSwimmer)
	f.shot(t, "large-audit-class-saved-long-name")
	if f.c.Snapshot().Screen != "class_assessment" {
		t.Fatal("saving one student closed pending class batch")
	}
	var savedID string
	for _, p := range people {
		if p.Name == current {
			savedID = p.ID
			break
		}
	}
	if savedID == "" {
		t.Fatal("saved student not in roster")
	}
	if saved := must(f.s.ListAttempts(savedID, skill.ID, course.CurriculumVersion)); len(saved) != 1 || saved[0].SubmittedAt == nil {
		t.Fatalf("current student's marked attempt not saved: %+v", saved)
	}
	if button := f.view.Buttons["row:"+savedID]; button == nil || !button.Disabled {
		t.Fatal("saved student should be displayed as saved and unavailable in pending class batch")
	}
	f.tap(t, "discard-all")
	if f.c.Snapshot().Screen != "skill_swimmers" {
		t.Fatalf("discarding pending batch returned to %s", f.c.Snapshot().Screen)
	}
	if saved := must(f.s.ListAttempts(savedID, skill.ID, course.CurriculumVersion)); len(saved) != 1 || saved[0].SubmittedAt == nil {
		t.Fatal("discarding remaining students removed an already saved attempt")
	}
}

func TestLargeMultiAssignMarkAndSaveOneGroup(t *testing.T) {
	f := setupUI(t)
	course := must(f.s.ListCourses("2026-09-04"))[0]
	people := must(f.s.ListRoster(course.ID))
	if len(people) < 4 {
		t.Fatal("need four demonstration swimmers")
	}
	catalog := must(f.s.CatalogFor(course.CurriculumVersion))
	level, ok := catalog.Level(people[0].Level)
	if !ok || len(level.Skills) == 0 {
		t.Fatal("missing demonstration skill")
	}
	skill := level.Skills[0]
	for _, person := range people {
		check(t, f.s.SetAttendance(person.ID, course.LessonID, "present"))
	}
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.row(t, 0)
	f.tap(t, "row:skills")
	f.tap(t, largeModeSwitch(t, f, "Multi"))
	if f.c.Snapshot().Screen != "multi_group_count" {
		t.Fatalf("multi switch opened %s", f.c.Snapshot().Screen)
	}
	f.tap(t, "multi-count-plus")
	f.tap(t, "submit")
	if f.c.Snapshot().Screen != "multi_selection" {
		t.Fatalf("group count opened %s", f.c.Snapshot().Screen)
	}
	for _, person := range people[:2] {
		f.tap(t, "row:"+person.ID)
	}
	f.tap(t, "row:1")
	for _, person := range people[2:4] {
		f.tap(t, "row:"+person.ID)
	}
	if !f.c.Snapshot().CanSubmit {
		t.Fatal("fully assigned groups cannot start marking")
	}
	f.shot(t, "large-audit-multi-assignment")
	assertLargeSwimmerLabelFits(t, f, longAssessmentPerson(t, f, course.ID), longAssessmentSwimmer)
	f.tap(t, "submit")
	if f.c.Snapshot().Screen != "multi_assessment" {
		t.Fatalf("group assignment opened %s", f.c.Snapshot().Screen)
	}
	f.shot(t, "large-audit-multi-marking")
	f.tap(t, "multi-pass:"+people[0].ID)
	f.tap(t, "multi-fail:"+people[1].ID)
	if f.c.Snapshot().Screen != "multi_student" {
		t.Fatal("incomplete action did not open student criteria")
	}
	f.tap(t, "criterion:"+f.c.Snapshot().Rows[0].ID)
	f.tap(t, "home")
	if f.c.Snapshot().Screen != "multi_assessment" {
		t.Fatal("saving student criteria did not return to group")
	}
	f.tap(t, "save-group")
	if f.c.Snapshot().Screen != "multi_assessment" {
		t.Fatal("saving first group closed second pending group")
	}
	f.tap(t, "discard-group")
	if f.c.Snapshot().Screen != "skill_swimmers" {
		t.Fatalf("discarding final group returned to %s", f.c.Snapshot().Screen)
	}
	for _, person := range people[:2] {
		attempts := must(f.s.ListAttempts(person.ID, skill.ID, course.CurriculumVersion))
		if len(attempts) != 1 || attempts[0].SubmittedAt == nil {
			t.Fatalf("saved group member %s lost attempt: %+v", person.Name, attempts)
		}
	}
	for _, person := range people[2:4] {
		if attempts := must(f.s.ListAttempts(person.ID, skill.ID, course.CurriculumVersion)); len(attempts) != 0 {
			t.Fatalf("discarded group member %s retained attempt: %+v", person.Name, attempts)
		}
	}
}

func TestLargeSkillPageSurvivesBackAndCourseSwitchUsesNewRoster(t *testing.T) {
	f := setupUI(t)
	courses := must(f.s.ListCourses("2026-09-04"))
	if len(courses) < 2 {
		t.Fatal("need two demonstration courses")
	}
	firstPeople := must(f.s.ListRoster(courses[0].ID))
	secondPeople := must(f.s.ListRoster(courses[1].ID))
	if len(firstPeople) == 0 || len(secondPeople) == 0 {
		t.Fatal("both courses need swimmers")
	}
	catalog := must(f.s.CatalogFor(courses[0].CurriculumVersion))
	level, ok := catalog.Level(firstPeople[0].Level)
	if !ok || len(level.Skills) < 7 {
		t.Fatal("need a second skill page")
	}
	secondPageSkill := level.Skills[6].ID
	f.c.ConfigureDisplay(ui.DisplayLarge, nil)
	f.tap(t, "row:"+courses[0].ID)
	f.tap(t, "row:skills")
	f.tap(t, "large-skills-page-1")
	if f.view.Buttons["row:"+secondPageSkill] == nil {
		t.Fatal("second page skill did not appear")
	}
	f.tap(t, "row:"+secondPageSkill)
	if button := f.view.Buttons["row:"+secondPageSkill]; button == nil || !button.Selected {
		t.Fatal("second page skill was not selected")
	}
	f.tap(t, "row:"+firstPeople[0].ID)
	if f.c.Snapshot().Screen != "assessment" {
		t.Fatal("selected skill did not open single assessment")
	}
	f.tap(t, "back")
	if f.c.Snapshot().Screen != "skills" {
		t.Fatal("discarding draft did not return to Skills")
	}
	if button := f.view.Buttons["row:"+secondPageSkill]; button == nil || !button.Selected {
		t.Fatal("skill page and selection were not retained on return")
	}
	f.tap(t, "home")
	f.tap(t, "row:"+courses[1].ID)
	f.tap(t, "row:skills")
	if f.view.Buttons["row:"+firstPeople[0].ID] != nil {
		t.Fatal("old course swimmer leaked into new course roster")
	}
	if f.view.Buttons["row:"+secondPeople[0].ID] == nil {
		t.Fatal("new course swimmer is missing")
	}
	f.tap(t, "row:"+secondPeople[0].ID)
	if f.c.Frame().Person != secondPeople[0].ID {
		t.Fatal("new course assessment retained previous swimmer")
	}
}
