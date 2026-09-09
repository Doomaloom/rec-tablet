package feedback

import (
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
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
