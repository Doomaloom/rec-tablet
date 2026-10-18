package feedback

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
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
	people := make([]domain.Enrollment, 8)
	for i := range people {
		people[i] = domain.Enrollment{ID: fmt.Sprint(i), Name: fmt.Sprintf("Swimmer %d", i), Active: i != 7}
	}
	return people, nil
}
func (manyFeedbackStore) GetEnrollment(id string) (domain.Enrollment, error) {
	return domain.Enrollment{ID: id, Name: "Swimmer " + id}, nil
}
func (manyFeedbackStore) Feedback(id string) ([]domain.Feedback, error) {
	return []domain.Feedback{{Skill: curriculum.Skill{ID: "skill-" + id, Name: "Skill " + id}, Version: "pinned", Progress: domain.CalculateProgress(2, 1), Remaining: []curriculum.Subskill{{Text: "Keep the head low"}, {Text: "Breathe to the side"}}}}, nil
}

func TestLargeFeedbackKeepsSidebarPageIndependentOfSelectedSwimmer(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = ui.DisplayLarge
	c := New(nav, manyFeedbackStore{})
	nav.Features[navigation.FeedbackSwimmers] = c
	nav.Push(navigation.FeedbackSwimmers, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if nav.State.PageSize != 6 || nav.State.RowCount != 8 || nav.State.PageCount != 2 || nav.Frame().Person != "0" {
		t.Fatalf("initial large feedback state: %+v, frame %+v", nav.State, nav.Frame())
	}
	if len(nav.State.Rows) != 6 || !nav.State.Rows[0].Selected {
		t.Fatal("first swimmer should be selected on entry")
	}
	if err := nav.Page(1); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Person != "0" || nav.State.Page != 1 {
		t.Fatal("paging changed selected swimmer")
	}
	if err := nav.Select("7"); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.FeedbackSwimmers || nav.Frame().Person != "7" || !nav.State.Rows[1].Selected {
		t.Fatalf("selecting swimmer should remain in combined view: %+v", nav.Frame())
	}
	if c.large.person.Active || c.large.entries[0].Skill.ID != "skill-7" {
		t.Fatal("former swimmer feedback was not loaded")
	}
	if err := nav.Page(-1); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Person != "7" || nav.State.Page != 0 {
		t.Fatal("sidebar page changed selected swimmer")
	}
}

func TestLargeFeedbackDetailIncludesEveryCriterionAndTimedResult(t *testing.T) {
	entry := domain.Feedback{Progress: domain.CalculateProgress(2, 1), Timed: true, TimedLaps: 3, TimedRequired: 4, TimedElapsedMS: 65400,
		Remaining: []curriculum.Subskill{{Text: "Keep the head low"}, {Text: "Breathe to the side"}}}
	text := feedbackDetailText(entry)
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
	data := c.large
	// Several cards must extend beyond one viewport without being clipped by the page limit.
	for i := 0; i < 4; i++ {
		data.entries = append(data.entries, data.entries[0])
	}
	surface := &ui.Surface{Button: func(_, label string, run func()) *ui.TouchButton { return ui.NewButton(label, run) }}
	scroll := newFeedbackScroll(surface, c, data)
	scroll.SetScale(1)
	scroll.Resize(fyne.NewSize(440, 180))
	if scroll.Content.MinSize().Height <= scroll.Size().Height {
		t.Fatal("feedback details should scroll beyond the viewport")
	}
	scroll.buttons[0].OnTap()
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
	view := shell.New(nav, window)
	t.Cleanup(func() { view.Close(); window.Close() })
	window.Resize(fyne.NewSize(1024, 600))
	window.Show()
	if view.Buttons["row:0"] == nil || view.Buttons["feedback-attempts:skill-0:pinned"] == nil {
		t.Fatal("large view did not expose swimmer selection and attempt access")
	}
	if view.Buttons["feedback-attempts:skill-0:pinned"].Size().Width <= 0 {
		t.Fatal("scrolling attempt action was not laid out")
	}
}

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
