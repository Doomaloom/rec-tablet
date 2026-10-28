package feedback

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	host "rectablet/next/app"
	"rectablet/next/features/feedback/skilldetail"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

type feedbackStore struct{}

func (feedbackStore) GetCourse(string) (domain.Course, error) {
	return domain.Course{Name: "Splash 1"}, nil
}
func (feedbackStore) ListRoster(string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{{ID: "former", Name: "Swimmer"}}, nil
}
func (feedbackStore) GetEnrollment(string) (domain.Enrollment, error) {
	return domain.Enrollment{ID: "former", Name: "Swimmer"}, nil
}
func (feedbackStore) Feedback(string) ([]domain.Feedback, error) {
	return []domain.Feedback{{Skill: curriculum.Skill{ID: "skill", Name: "Skill"}, Version: "pinned", Progress: domain.CalculateProgress(2, 1), Remaining: []curriculum.Subskill{{ID: "criterion", Text: "Practice breathing"}}}}, nil
}

func TestFeedbackIncludesFormerEnrollmentsAndPreservesVersion(t *testing.T) {
	nav := navigation.New()
	c := New(nav, feedbackStore{})
	for _, route := range []navigation.Route{navigation.FeedbackSwimmers, navigation.Feedback, navigation.FeedbackDetail} {
		nav.Features[route] = c
	}
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 1 || nav.State.Rows[0].Disabled {
		t.Fatal("former enrollment unavailable for feedback")
	}
	if err := nav.Select("former"); err != nil {
		t.Fatal(err)
	}
	if err := nav.Select("skill"); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Version != "pinned" || nav.State.PageSize != 1 || nav.State.Rows[0].Label != "Practice breathing" {
		t.Fatal("feedback lost pinned version or detail")
	}
	if err := c.Header(); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.Attempts || nav.Frame().Version != "pinned" {
		t.Fatal("attempt review lost context")
	}
}

type manyFeedbackStore struct{}

func (manyFeedbackStore) GetCourse(string) (domain.Course, error) {
	return domain.Course{Name: "Splash 1"}, nil
}
func (manyFeedbackStore) ListRoster(string) ([]domain.Enrollment, error) {
	people := make([]domain.Enrollment, 12)
	for i := range people {
		people[i] = domain.Enrollment{ID: fmt.Sprint(i), Name: fmt.Sprintf("Swimmer %d", i), Active: i != 11}
	}
	return people, nil
}
func (manyFeedbackStore) GetEnrollment(id string) (domain.Enrollment, error) {
	return domain.Enrollment{ID: id, Name: "Swimmer " + id}, nil
}
func (manyFeedbackStore) Feedback(id string) ([]domain.Feedback, error) {
	return []domain.Feedback{{Skill: curriculum.Skill{ID: "skill-" + id, Name: "Skill " + id}, Version: "pinned", Progress: domain.CalculateProgress(2, 1), Remaining: []curriculum.Subskill{{Text: "Keep the head low"}, {Text: "Breathe to the side"}}}, {Skill: curriculum.Skill{ID: "second-" + id, Name: "Second skill"}, Version: "pinned"}, {Skill: curriculum.Skill{ID: "third-" + id, Name: "Third skill"}, Version: "pinned"}}, nil
}

func TestLargeFeedbackPagesSkillsAndKeepsRosterVisible(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	c := New(nav, manyFeedbackStore{})
	nav.Features[navigation.FeedbackSwimmers] = c
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 12 || nav.State.PageCount != 2 || nav.Frame().Person != "0" || !nav.State.Rows[0].Selected {
		t.Fatalf("initial state: %+v", nav.State)
	}
	if err := nav.Page(1); err != nil {
		t.Fatal(err)
	}
	if nav.State.Page != 1 || len(nav.State.Rows) != 12 || nav.Frame().Person != "0" {
		t.Fatal("feedback paging changed roster or swimmer")
	}
	if err := nav.Select("11"); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Person != "11" || nav.State.Page != 0 || !nav.State.Rows[11].Selected || c.skillSummary.Selected.Person.Active {
		t.Fatal("selection did not reset feedback page or retain former swimmer")
	}
}

func TestLargeFeedbackDetailIncludesEveryCriterionAndTimedResult(t *testing.T) {
	entry := domain.Feedback{Progress: domain.CalculateProgress(2, 1), Timed: true, TimedLaps: 3, TimedRequired: 4, TimedElapsedMS: 65400,
		Remaining: []curriculum.Subskill{{Text: "Keep the head low"}, {Text: "Breathe to the side"}}}
	text := (&skilldetail.SkillDetail{}).Text(entry)
	for _, expected := range []string{"3/4 laps", "01:05.4", "Keep the head low", "Breathe to the side"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q in %q", expected, text)
		}
	}
}

func TestLargeFeedbackScrollAndAttemptAccess(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	c := New(nav, manyFeedbackStore{})
	nav.Features[navigation.FeedbackSwimmers] = c
	nav.Features[navigation.Attempts] = c
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	data := c.skillSummary.Selected
	// Several cards must extend beyond one viewport without being clipped by the page limit.
	for i := 0; i < 4; i++ {
		data.Entries = append(data.Entries, data.Entries[0])
	}
	var attempt *ui.TouchButton
	surface := &ui.Surface{
		Button: func(_, label string, run func()) *ui.TouchButton {
			b := ui.NewButton(label, run)
			if attempt == nil && label == "Attempts" {
				attempt = b
			}
			return b
		},
		AddNamed: func(string, fyne.CanvasObject, func(float32, float32) (float32, float32, float32, float32)) {},
	}
	c.skillSummary.ComposeLarge(data)(surface)
	if attempt == nil {
		t.Fatal("missing attempt action")
	}
	attempt.OnTap()
	if nav.Frame().Screen != navigation.Attempts || nav.Frame().Person != "0" || nav.Frame().Skill != "skill-0" || nav.Frame().Version != "pinned" {
		t.Fatalf("attempt access lost context: %+v", nav.Frame())
	}
}

func TestLargeFeedbackCompositionExposesSelectedDetails(t *testing.T) {
	app := test.NewApp()
	t.Cleanup(app.Quit)
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	c := New(nav, manyFeedbackStore{})
	nav.Features[navigation.FeedbackSwimmers] = c
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	window := app.NewWindow("Feedback")
	view := host.New(nav, window)
	t.Cleanup(func() { view.Close(); window.Close() })
	window.Resize(fyne.NewSize(1024, 600))
	window.Show()
	if view.Buttons["row:0"] == nil || view.Buttons["feedback-attempts:skill-0:pinned"] == nil {
		t.Fatal("large view did not expose swimmer selection and attempt access")
	}
	first, last := view.Buttons["row:0"], view.Buttons["row:11"]
	if first.Position().Y != ui.OuterMargin || last == nil || last.Position().Y+last.Size().Height > 600-ui.OuterMargin+1 {
		t.Fatal("all twelve students must fit in the full-height column")
	}
	if view.Buttons["feedback-attempts:skill-0:pinned"].Size().Width <= 0 {
		t.Fatal("scrolling attempt action was not laid out")
	}
}

func (manyFeedbackStore) CompletedFeedback(id string) ([]domain.Feedback, error) {
	return []domain.Feedback{{Skill: curriculum.Skill{ID: "complete-" + id, Name: "Completed skill"}, Version: "pinned", Progress: domain.CalculateProgress(3, 3)}}, nil
}

func TestCompletionSwitchResetsPageAndKeepsStudent(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	c := New(nav, manyFeedbackStore{})
	nav.Features[navigation.FeedbackSwimmers] = c
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if err := nav.Page(1); err != nil {
		t.Fatal(err)
	}
	c.setFilter(true)
	if nav.State.Page != 0 || nav.State.PageCount != 1 || nav.Frame().Person != "0" || len(nav.State.Rows) != 12 || c.skillSummary.Selected.Entries[0].Skill.ID != "complete-0" {
		t.Fatal("complete filter lost student, page, or completed skills")
	}
	c.setFilter(false)
	if nav.State.PageCount != 2 || c.skillSummary.Selected.Entries[0].Skill.ID != "skill-0" {
		t.Fatal("in progress filter did not restore skills")
	}
}

func (feedbackStore) CompletedFeedback(string) ([]domain.Feedback, error) { return nil, nil }
