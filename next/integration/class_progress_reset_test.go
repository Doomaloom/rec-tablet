package integration_test

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/ui"
)

func TestClassHoldConfirmsProgressReset(t *testing.T) {
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		t.Run(string(mode), func(t *testing.T) {
			f := setupUI(t)
			f.c.ConfigureDisplay(mode, nil)
			f.c.ConfigureSession(func() {}, "https://share.example", nil)
			course := must(f.s.ListCourses("2026-09-04"))[0]
			person := must(f.s.ListRoster(course.ID))[0]
			attempt := must(f.s.NewAttempt(person.ID, course.LessonID, "Splash1:1"))
			skill := must(f.s.Skill(course.CurriculumVersion, "Splash1:1"))
			for _, criterion := range skill.Subskills {
				check(t, f.s.Mark(attempt, criterion.ID, "successful"))
			}
			if !must(f.s.Submit(attempt)) {
				t.Fatal("seed attempt failed")
			}
			button := f.view.Buttons["row:"+course.ID]
			if button.OnHold == nil || button.HoldDuration != 5*time.Second {
				t.Fatal("class card has no five-second hold")
			}
			for _, dismiss := range []string{"Cancel", "Escape"} {
				button.OnHold()
				overlay := f.w.Canvas().Overlays().Top()
				if overlay == nil || f.c.Snapshot().Screen != "classes" {
					t.Fatal("hold did not open confirmation on class list")
				}
				if p := must(f.s.Progress(person.ID, skill.ID, course.CurriculumVersion)); p.Successes != 1 {
					t.Fatal("holding deleted progress without confirmation")
				}
				if dismiss == "Escape" {
					f.w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
				} else {
					test.Tap(dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == "Cancel" }))
				}
				if f.w.Canvas().Overlays().Top() != nil {
					t.Fatal("cancel left modal open")
				}
			}
			button.OnHold()
			f.shot(t, "class-reset-confirm-"+string(mode))
			overlay := f.w.Canvas().Overlays().Top()
			accept := dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == "Delete progress" })
			test.Tap(accept)
			test.Tap(accept) // A retained confirmation cannot delete twice.
			if f.c.Snapshot().Error {
				t.Fatal(f.c.Snapshot().Notice)
			}
			if p := must(f.s.Progress(person.ID, skill.ID, course.CurriculumVersion)); p.Assessed != 0 || p.Successes != 0 {
				t.Fatalf("reset failed: %+v", p)
			}
			if f.w.Canvas().Overlays().Top() != nil {
				t.Fatal("accepted modal remained open")
			}
			f.tap(t, "row:"+course.ID) // Ordinary taps still open the class.
			button.OnHold()            // A detached hold cannot open a new dialog.
			if f.c.Snapshot().Screen != "class" || f.w.Canvas().Overlays().Top() != nil {
				t.Fatal("stale class hold opened a modal")
			}
		})
	}
}
