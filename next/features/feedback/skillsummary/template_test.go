package skillsummary

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

func TestLargeFeedbackLongSkillTitleExpandsCard(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	short := &feedbackCardLayout{title: "Front crawl", details: "Needs practice", width: 500, scale: 1}
	long := &feedbackCardLayout{title: strings.Repeat("Rescue safely while watching every swimmer in the pool ", 4), details: "Needs practice", width: 500, scale: 1}
	if long.titleHeight(500) <= short.titleHeight(500) || long.cardHeight() <= short.cardHeight() {
		t.Fatal("long skill title should wrap and increase card height")
	}
	if got := long.cardHeight() - long.titleHeight(500); got < 48+12 {
		t.Fatalf("card leaves %.1f pixels for details and action, want a 48-pixel touch target", got)
	}
}

func TestSkillCardsScrollBeyondViewport(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(a.Quit)
	data := Snapshot{Person: domain.Enrollment{Name: "Alex"}, Entries: make([]domain.Feedback, 5)}
	summary := SkillSummary{Detail: func(domain.Feedback) string { return "Needs practice" }}
	surface := &ui.Surface{Button: func(_, label string, run func()) *ui.TouchButton { return ui.NewButton(label, run) }}
	scroll := newFeedbackScroll(surface, &summary, data)
	scroll.SetScale(1)
	scroll.Resize(fyne.NewSize(440, 180))
	if scroll.Content.MinSize().Height <= scroll.Size().Height {
		t.Fatal("skill cards do not extend beyond viewport")
	}
}
