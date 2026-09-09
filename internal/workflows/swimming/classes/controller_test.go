package classes

import (
	"errors"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
)

type classStore struct {
	Repository
	err error
}

func (s *classStore) Meta(string) (string, error) { return "Instructor", nil }
func (s *classStore) Dates() ([]string, error)    { return []string{"2026-09-04", "2026-09-11"}, nil }
func (s *classStore) ListCourses(string) ([]domain.Course, error) {
	return []domain.Course{{ID: "course", LessonID: "lesson", Name: "Splash 1"}}, s.err
}

func TestDateSelectionRestoresNavigationOnReadFailure(t *testing.T) {
	nav := navigation.New()
	nav.Date = "2026-09-04"
	s := &classStore{}
	c := New(nav, s)
	nav.Features[navigation.Classes] = c
	nav.Features[navigation.Dates] = c
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if err := nav.Header(); err != nil {
		t.Fatal(err)
	}
	s.err = errors.New("read failed")
	if nav.Select("2026-09-11") == nil || nav.Date != "2026-09-04" || nav.Frame().Screen != navigation.Dates {
		t.Fatal("failed date selection changed navigation")
	}
	s.err = nil
	if err := nav.Select("2026-09-11"); err != nil {
		t.Fatal(err)
	}
	if nav.Date != "2026-09-11" || nav.Frame().Screen != navigation.Classes {
		t.Fatal("date selection did not return home")
	}
	if navigation.Destination(nav.State.Rows[0]).Lesson != "lesson" {
		t.Fatal("typed lesson target lost")
	}
}
