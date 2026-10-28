package feedback

import (
	"fmt"
	"fyne.io/fyne/v2"
	"rectablet/src/shared/ui"
	"testing"

	"rectablet/next/contracts"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestSubskillsUsesSixCardPagesAndPinnedVersion(t *testing.T) {
	state := contracts.State{}
	entry := domain.Feedback{
		Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "pinned",
		Progress: domain.Progress{Assessed: 5},
		MissedSubskills: []domain.SubskillMiss{
			{Subskill: curriculum.Subskill{ID: "head", Text: "Keep head low"}, Count: 2},
			{Subskill: curriculum.Subskill{ID: "legs", Text: "Keep legs straight"}, Count: 3},
		},
	}
	other := entry
	other.Version = "other"
	rows := configureSubskills(&state, domain.Enrollment{Name: "Alex"}, []domain.Feedback{other, entry}, "skill", "pinned")
	if state.PageSize != 6 || len(rows) != 2 || rows[0].Label != "Keep head low" || rows[0].Status != "Missed 2 of 5 saved attempts" || rows[1].Status != "Missed 3 of 5 saved attempts" {
		t.Fatalf("subskill pages: %+v, %+v", state, rows)
	}
	if state.PrimaryInfo != "Float" || state.Title != "Subskills · Alex" {
		t.Fatal("subskills lost swimmer or skill context")
	}
}

func TestSixSubskillCardsFitPage(t *testing.T) {
	state := contracts.State{Rows: make([]contracts.Row, 6)}
	for i := range state.Rows {
		state.Rows[i] = contracts.Row{ID: fmt.Sprint(i), Label: "Keep head low", Status: "Missed 2 of 5 saved attempts"}
	}
	cards := 0
	previousBottom := float32(0)
	surface := &ui.Surface{AddNamed: func(name string, object fyne.CanvasObject, bounds func(float32, float32) (float32, float32, float32, float32)) {
		x, y, w, h := bounds(1024, 600)
		if x < ui.ContentInset || y < 52 || x+w > 1024-ui.ContentInset+1 || y+h > 596+1 || w <= 0 || h <= 0 {
			t.Fatalf("%s exceeds page: %v %v %v %v", name, x, y, w, h)
		}
		if name == "feedback-subskill-card-"+fmt.Sprint(cards) {
			if x != ui.ContentInset || w != 1024-2*ui.ContentInset || y < previousBottom {
				t.Fatal("subskill cards must form one full-width column")
			}
			previousBottom = y + h
			cards++
		}
	}}
	composeSubskills(state)(surface)
	if cards != 6 {
		t.Fatalf("rendered %d cards", cards)
	}
}
