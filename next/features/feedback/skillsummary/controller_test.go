package skillsummary

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"rectablet/next/features/feedback/skilldetail"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestSkillSummaryAttemptCallbacksPreserveEachVersion(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	type target struct{ Person, skill, version string }
	var opened []target
	summary := SkillSummary{
		Detail:     (&skilldetail.SkillDetail{}).Text,
		OnAttempts: func(person, skill, version string) { opened = append(opened, target{person, skill, version}) },
	}
	data := Snapshot{
		Person: domain.Enrollment{ID: "former", Name: "Alex"},
		Entries: []domain.Feedback{
			{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "older"},
			{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "newer"},
		},
	}
	surface := &ui.Surface{Button: func(_, label string, run func()) *ui.TouchButton { return ui.NewButton(label, run) }}
	scroll := newFeedbackScroll(surface, &summary, data)
	scroll.buttons[0].OnTap()
	scroll.buttons[1].OnTap()
	if len(opened) != 2 || opened[0] != (target{"former", "skill", "older"}) || opened[1] != (target{"former", "skill", "newer"}) {
		t.Fatalf("attempt targets = %+v", opened)
	}
}
