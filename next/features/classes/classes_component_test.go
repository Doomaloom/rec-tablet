package classes

import (
	"errors"
	"fmt"
	"testing"

	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

type classStore struct {
	Repository
	err     error
	courses []domain.Course
}

func (s *classStore) Meta(string) (string, error) { return "Instructor", nil }
func (s *classStore) Dates() ([]string, error)    { return []string{"2026-09-04", "2026-09-11"}, nil }
func (s *classStore) GetCourse(id string) (domain.Course, error) {
	for _, course := range s.courses {
		if course.ID == id {
			return course, nil
		}
	}
	return domain.Course{Name: "Splash 1", Time: "9:00 AM"}, nil
}
func (s *classStore) ListRoster(string) ([]domain.Enrollment, error) { return nil, nil }
func (s *classStore) ListCourses(string) ([]domain.Course, error) {
	if s.courses != nil {
		return s.courses, s.err
	}
	return []domain.Course{{ID: "course", LessonID: "lesson", Name: "Splash 1"}}, s.err
}

func TestLargeClassCardsAndLessonMenuPagination(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = "large"
	nav.Date = "2026-09-04"
	courses := make([]domain.Course, 7)
	for i := range courses {
		courses[i] = domain.Course{ID: fmt.Sprintf("course-%d", i), LessonID: fmt.Sprintf("lesson-%d", i), Name: fmt.Sprintf("Splash %d", i+1), Time: "9:00 AM"}
	}
	s := &classStore{courses: courses}
	c := New(nav, s)
	nav.Features[navigation.Classes] = c
	nav.Features[navigation.Class] = c
	nav.Features[navigation.Dates] = c
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if nav.State.PageSize != 6 || nav.State.PageCount != 2 || len(nav.State.Rows) != 6 || nav.State.Rows[0].Label != "Splash 1 · 9:00 AM" {
		t.Fatalf("large class card page = %d rows, size %d, pages %d", len(nav.State.Rows), nav.State.PageSize, nav.State.PageCount)
	}
	if err := nav.Page(1); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 1 || nav.State.Rows[0].ID != "course-6" {
		t.Fatalf("last class card page = %+v", nav.State.Rows)
	}
	if err := nav.Select("course-6"); err != nil {
		t.Fatal(err)
	}
	if nav.State.PageSize != 4 || len(nav.State.Rows) != 4 || nav.State.PageCount != 1 {
		t.Fatalf("large lesson menu = %d rows, size %d, pages %d", len(nav.State.Rows), nav.State.PageSize, nav.State.PageCount)
	}
}

func TestLargeDateCalendarTracksSelectedMonthAndScheduledDates(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = "large"
	nav.Date = "2026-09-04"
	s := &classStore{courses: []domain.Course{{ID: "course", LessonID: "lesson", Name: "Splash 1"}}}
	c := New(nav, s)
	nav.Features[navigation.Classes] = c
	nav.Features[navigation.Dates] = c
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if err := nav.Header(); err != nil {
		t.Fatal(err)
	}
	if c.datePicker.Month().Format("2006-01") != "2026-09" || nav.State.Rows[0].ID != "2026-09-04" || !nav.State.Rows[0].Selected {
		t.Fatalf("calendar did not start on selected lesson month/date: month %s, rows %+v", c.datePicker.Month().Format("2006-01"), nav.State.Rows)
	}
	if err := c.moveCalendarMonth(1); err != nil {
		t.Fatal(err)
	}
	if c.datePicker.Month().Format("2006-01") != "2026-10" {
		t.Fatalf("next month moved calendar to %s", c.datePicker.Month().Format("2006-01"))
	}
	if err := nav.Select("2026-09-11"); err != nil {
		t.Fatal(err)
	}
	if nav.Date != "2026-09-11" || nav.Frame().Screen != navigation.Classes {
		t.Fatalf("scheduled date selection did not return to classes: date %s, screen %s", nav.Date, nav.Frame().Screen)
	}
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
