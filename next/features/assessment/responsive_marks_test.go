package assessment

import (
	"errors"
	"testing"

	"fyne.io/fyne/v2/test"
	host "rectablet/next/app"
	"rectablet/next/features/assessment/classassessment"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestClassChecklistTapReusesControlsAndDoesNotReloadProgress(t *testing.T) {
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		t.Run(string(mode), func(t *testing.T) {
			app := test.NewApp()
			defer app.Quit()
			window := app.NewWindow("Class marking")
			defer window.Close()
			catalog := curriculum.Bundled()
			var skill curriculum.Skill
			for _, level := range catalog.Levels {
				if len(level.Skills) > 0 && len(level.Skills[0].Subskills) > 0 {
					skill = level.Skills[0]
					break
				}
			}
			store := &largeAssessmentStore{catalog: catalog, answers: map[string]map[string]string{"attempt": {}}}
			nav := navigation.New()
			nav.DisplayMode = mode
			nav.Push(navigation.ClassAssessment, func(f *navigation.Frame) { f.Course = "course"; f.Skill = skill.ID; f.Version = catalog.Version })
			feature := New(nav, store)
			people := []domain.MultiAttempt{{AttemptID: "attempt", PersonID: "person", Name: "Student"}}
			feature.class.State = classassessment.State{Attempts: people, Roster: people, Status: map[string]string{}}
			nav.Features[navigation.ClassAssessment] = feature
			if err := nav.Refresh(); err != nil {
				t.Fatal(err)
			}
			view := host.New(nav, window)
			defer view.Close()
			window.Show()
			prefix := "row:"
			if mode == ui.DisplayLarge {
				prefix = "criterion:"
			}
			id := skill.Subskills[0].ID
			button := view.Buttons[prefix+id]
			if button == nil {
				t.Fatal("criterion missing")
			}
			// An old screen rebuild would query every student's progress again here.
			store.failProgress = errors.New("progress should not be reloaded on a draft tap")
			if err := nav.Select(id); err != nil {
				t.Fatal(err)
			}
			if view.Buttons[prefix+id] != button || button.Fill != ui.Green || store.answers["attempt"][id] != "successful" {
				t.Fatal("class tap rebuilt controls or failed to save/paint")
			}
			if err := nav.Select(id); err != nil {
				t.Fatal(err)
			}
			if button.Fill == ui.Green || store.answers["attempt"][id] != "unmarked" {
				t.Fatal("class tap did not toggle off")
			}
		})
	}
}
