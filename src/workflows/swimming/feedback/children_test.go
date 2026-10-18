package feedback

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestSkillSummaryAttemptCallbacksPreserveEachVersion(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	type target struct{ person, skill, version string }
	var opened []target
	summary := SkillSummary{
		detail:     &SkillDetail{},
		onAttempts: func(person, skill, version string) { opened = append(opened, target{person, skill, version}) },
	}
	data := summaryData{
		person: domain.Enrollment{ID: "former", Name: "Alex"},
		entries: []domain.Feedback{
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

func TestSkillDetailUsesOnlyPinnedVersion(t *testing.T) {
	detail := SkillDetail{}
	state := shell.State{PrimaryInfo: "Swimming"}
	entries := []domain.Feedback{
		{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "older", Remaining: []curriculum.Subskill{{Text: "Older criterion"}}},
		{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "pinned", Remaining: []curriculum.Subskill{{Text: "Pinned criterion"}}},
	}
	rows := detail.Configure(&state, domain.Enrollment{Name: "Alex"}, entries, "skill", "pinned")
	if len(rows) != 1 || rows[0].Label != "Pinned criterion" || !rows[0].Disabled {
		t.Fatalf("pinned criteria = %+v", rows)
	}
	if state.PageSize != 1 || state.Title != "Alex" || state.HeaderAction != "Attempts" {
		t.Fatalf("detail state = %+v", state)
	}
}
