package skillsummary

import (
	"rectablet/src/workflows/swimming/curriculum"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

func TestLongFeedbackTextCannotExpandCards(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	long := strings.Repeat("Rescue safely while watching every swimmer in the pool ", 20)
	data := Snapshot{Entries: []domain.Feedback{{Skill: curriculum.Skill{Name: "Short skill"}}, {Skill: curriculum.Skill{Name: long}}}}
	summary := SkillSummary{Detail: func(domain.Feedback) string { return long }}
	surface := &ui.Surface{Button: func(_, label string, run func()) *ui.TouchButton { return ui.NewButton(label, run) }}
	cards := newFeedbackCards(surface, &summary, data)
	for _, scale := range []float32{1, 1.5} {
		cards.SetScale(scale)
		cards.Resize(fyne.NewSize(600*scale, 450*scale))
		expected := fyne.NewSize(600*scale, 221*scale)
		for _, object := range cards.content.Objects {
			if object.Size() != expected {
				t.Fatalf("card size %v, want fixed %v", object.Size(), expected)
			}
		}
		for _, index := range []int{2, 3, 5} {
			label := cards.labels[index]
			renderer := test.WidgetRenderer(label)
			renderer.Layout(label.Size())
			if !label.Truncated {
				t.Fatal("overflowing feedback text must be truncated")
			}
			for _, line := range renderer.Objects() {
				if line.Position().Y+line.Size().Height > label.Size().Height {
					t.Fatal("text draws outside its allocated area")
				}
			}
		}
	}
	// A final page with one card retains the same slot size.
	single := newFeedbackCards(surface, &summary, Snapshot{Entries: data.Entries[:1]})
	single.Resize(fyne.NewSize(600, 450))
	if single.content.Objects[0].Size() != fyne.NewSize(600, 221) {
		t.Fatal("last page changed card size")
	}
}

func TestSkillCardsFitPageWithoutScrolling(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	data := Snapshot{Entries: make([]domain.Feedback, 2)}
	summary := SkillSummary{Detail: func(domain.Feedback) string { return "Needs practice" }}
	surface := &ui.Surface{Button: func(_, label string, run func()) *ui.TouchButton { return ui.NewButton(label, run) }}
	cards := newFeedbackCards(surface, &summary, data)
	renderer := test.WidgetRenderer(cards)
	if len(renderer.Objects()) != 1 || renderer.Objects()[0] != cards.content {
		t.Fatal("feedback widget must expose its cards to the canvas")
	}
	cards.SetScale(1)
	cards.Resize(fyne.NewSize(600, 450))
	for _, object := range cards.content.Objects {
		if object.Position().Y+object.Size().Height > 450 {
			t.Fatal("feedback card exceeds page")
		}
		card := object.(*fyne.Container)
		for _, child := range card.Objects {
			if child.Position().Y+child.Size().Height > object.Size().Height {
				t.Fatal("feedback content exceeds card")
			}
		}
	}
}

func TestSubskillsButtonOpensMissSummary(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("Feedback")
	defer window.Close()
	window.Resize(fyne.NewSize(1024, 600))
	entry := domain.Feedback{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "pinned", Progress: domain.Progress{Assessed: 5}, MissedSubskills: []domain.SubskillMiss{{Subskill: curriculum.Subskill{Text: "Keep head low"}, Count: 2}}}

	buttons := map[string]*ui.TouchButton{}
	surface := &ui.Surface{Window: window, Button: func(id, label string, run func()) *ui.TouchButton {
		b := ui.NewButton(label, run)
		buttons[id] = b
		return b
	}}
	var target string
	summary := SkillSummary{Detail: func(domain.Feedback) string { return "Needs practice" }, OnSubskills: func(person, skill, version string) { target = person + ":" + skill + ":" + version }}
	scroll := newFeedbackCards(surface, &summary, Snapshot{Person: domain.Enrollment{ID: "student"}, Entries: []domain.Feedback{entry}})
	scroll.Resize(fyne.NewSize(600, 400))
	attempts := buttons["feedback-attempts:skill:pinned"]
	subskills := buttons["feedback-subskills:skill:pinned"]
	if subskills == nil || subskills.Size().Height < 48 || subskills.Position().X <= attempts.Position().X {
		t.Fatal("subskills action must sit beside attempts")
	}
	subskills.OnTap()
	if target != "student:skill:pinned" {
		t.Fatalf("subskills target: %s", target)
	}
	if len(window.Canvas().Overlays().List()) != 0 {
		t.Fatal("subskills must navigate without a popup")
	}
}
