package assessment

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2"

	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

type largeSkillFlowStore struct {
	largeAssessmentStore
	failOpen  error
	discarded [][]string
}

func (s *largeSkillFlowStore) OpenMultiAttempts(people []string, lesson, skill string) ([]domain.MultiAttempt, error) {
	if s.failOpen != nil {
		return nil, s.failOpen
	}
	var attempts []domain.MultiAttempt
	for _, id := range people {
		attempt, err := s.NewAttempt(id, lesson, skill)
		if err != nil {
			return nil, err
		}
		person, err := s.GetEnrollment(id)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, domain.MultiAttempt{PersonID: id, Name: person.Name, AttemptID: attempt})
	}
	return attempts, nil
}

func (s *largeSkillFlowStore) DiscardAttempts(ids []string) error {
	s.discarded = append(s.discarded, append([]string(nil), ids...))
	return nil
}

func newLargeSkillFlow(t *testing.T, mode ui.DisplayMode) (*navigation.Controller, *Component, *largeSkillFlowStore) {
	t.Helper()
	store := &largeSkillFlowStore{largeAssessmentStore: largeAssessmentStore{
		catalog: curriculum.Bundled(),
		roster: []domain.Enrollment{
			{ID: "a", Name: "Alex", Level: "Splash7", Active: true},
			{ID: "b", Name: "Blair", Level: "Splash7", Active: true},
			{ID: "inactive", Level: "Splash7"},
			{ID: "other", Level: "Splash8", Active: true},
		},
	}}
	nav := navigation.New()
	nav.DisplayMode = mode
	nav.Push(navigation.Skills, func(f *navigation.Frame) { f.Course, f.Lesson = "course", "lesson" })
	c := New(nav, store)
	for _, route := range []navigation.Route{navigation.Skills, navigation.SkillSwimmers, navigation.ClassAssessment, navigation.TimedLaps} {
		nav.Features[route] = c
	}
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	return nav, c, store
}

// Compose the actual feature buttons so tests exercise the timed/ordinary tap branch.
func tapSkill(t *testing.T, state navigation.State, id string) {
	t.Helper()
	buttons := map[string]*ui.TouchButton{}
	state.ComposeBody(&ui.Surface{
		DisplayMode: state.DisplayMode,
		Add:         func(fyne.CanvasObject, func(float32, float32) (float32, float32, float32, float32)) {},
		AddNamed:    func(string, fyne.CanvasObject, func(float32, float32) (float32, float32, float32, float32)) {},
		Button: func(key, label string, run func()) *ui.TouchButton {
			button := ui.NewButton(label, run)
			buttons[key] = button
			return button
		},
	})
	button := buttons["row:"+id]
	if button == nil {
		t.Fatalf("skill %q is not visible", id)
	}
	button.Tapped(nil)
}

func TestLargeSkillTapOpensClassAndReturnsToSkills(t *testing.T) {
	for _, exit := range []string{"save", "discard", "back"} {
		t.Run(exit, func(t *testing.T) {
			nav, c, store := newLargeSkillFlow(t, ui.DisplayLarge)
			if len(nav.State.HeaderSwitch) != 0 {
				t.Fatal("large skill picker offers marking modes")
			}
			tapSkill(t, nav.State, "Splash7:1")
			f := nav.Frame()
			if f.Screen != navigation.ClassAssessment || f.Skill != "Splash7:1" || f.Level != "Splash7" || f.Version != store.catalog.Version || f.Lesson != "lesson" {
				t.Fatalf("class destination = %+v; notice = %s", f, nav.State.Notice)
			}
			if len(c.class.State.Attempts) != 2 || c.class.State.Attempts[0].PersonID != "a" || c.class.State.Attempts[1].PersonID != "b" {
				t.Fatalf("class participants = %+v", c.class.State.Attempts)
			}
			if len(nav.History.Frames) != 3 {
				t.Fatalf("unexpected intermediate route: %+v", nav.History.Frames)
			}
			var err error
			switch exit {
			case "save":
				err = c.class.SaveClass()
				if len(store.submitted) != 1 || len(store.submitted[0]) != 2 {
					t.Fatalf("submitted = %v", store.submitted)
				}
			case "discard":
				err = c.class.DiscardClass()
				if len(store.discarded) != 1 || len(store.discarded[0]) != 2 {
					t.Fatalf("discarded = %v", store.discarded)
				}
			case "back":
				err = nav.Back()
			}
			if err != nil {
				t.Fatal(err)
			}
			if nav.Frame().Screen != navigation.Skills || len(nav.State.HeaderSwitch) != 0 {
				t.Fatalf("exit returned to %+v", nav.Frame())
			}
		})
	}
}

func TestLargeTimedSkillRetainsSelectionAndLapsFlow(t *testing.T) {
	nav, c, _ := newLargeSkillFlow(t, ui.DisplayLarge)
	c.timed.State.Selection, c.timed.State.LapsInput = []string{"stale"}, "99"
	tapSkill(t, nav.State, "Splash7:10")
	if nav.Frame().Screen != navigation.SkillSwimmers || nav.Frame().Skill != "Splash7:10" || len(c.class.State.Attempts) != 0 {
		t.Fatalf("timed destination = %+v", nav.Frame())
	}
	if len(c.timed.State.Selection) != 0 || c.timed.State.LapsInput != "" {
		t.Fatal("timed setup retained a previous selection")
	}
	if err := nav.Select("a"); err != nil {
		t.Fatal(err)
	}
	if len(c.timed.State.Selection) != 1 || c.timed.State.Selection[0] != "a" {
		t.Fatalf("timed selection = %v", c.timed.State.Selection)
	}
	if err := c.timed.ConfirmTimedSelection(); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.TimedLaps {
		t.Fatalf("expected lap entry, got %+v", nav.Frame())
	}
}

func TestLargeClassOpenFailureStaysOnSkills(t *testing.T) {
	nav, _, store := newLargeSkillFlow(t, ui.DisplayLarge)
	store.failOpen = errors.New("cannot open attempts")
	tapSkill(t, nav.State, "Splash7:1")
	if nav.Frame().Screen != navigation.Skills || nav.State.Notice != store.failOpen.Error() || len(nav.History.Frames) != 2 {
		t.Fatalf("failed class entry: frame = %+v, notice = %q", nav.Frame(), nav.State.Notice)
	}
}

func TestCompactSkillStillOffersMarkingModes(t *testing.T) {
	nav, _, _ := newLargeSkillFlow(t, ui.DisplayCompact)
	if err := nav.Select("Splash7:1"); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.SkillSwimmers || len(nav.State.HeaderSwitch) < 2 || nav.State.HeaderSwitch[0].Label != "Single" {
		t.Fatalf("compact mode changed: frame = %+v, modes = %+v", nav.Frame(), nav.State.HeaderSwitch)
	}
}
