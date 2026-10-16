package attendance

import (
	"errors"
	"fmt"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
)

type attendanceStore struct {
	err                    error
	person, lesson, status string
	marks                  map[string]string
	finalized              bool
	roster                 []domain.Enrollment
}

func (s *attendanceStore) GetCourse(string) (domain.Course, error) {
	return domain.Course{Name: "Splash 1"}, nil
}
func (s *attendanceStore) ListRoster(string) ([]domain.Enrollment, error) {
	if s.roster != nil {
		return s.roster, nil
	}
	return []domain.Enrollment{{ID: "active", Name: "A", Active: true}, {ID: "former", Name: "B"}}, nil
}
func (s *attendanceStore) Attendance(string) (map[string]string, error) {
	out := map[string]string{}
	for person, status := range s.marks {
		out[person] = status
	}
	return out, nil
}

func TestLargeAttendanceUsesTwoColumnsAndTwelveStudentPages(t *testing.T) {
	nav := navigation.New()
	nav.DisplayMode = "large"
	s := &attendanceStore{marks: map[string]string{}}
	for i := 0; i < 13; i++ {
		s.roster = append(s.roster, domain.Enrollment{ID: fmt.Sprintf("person-%02d", i), Name: fmt.Sprintf("Swimmer %02d", i), Active: true})
	}
	c := New(nav, s)
	nav.Features[navigation.Attendance] = c
	nav.Push(navigation.Attendance, func(f *navigation.Frame) { f.Course = "course"; f.Lesson = "lesson" })
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if nav.State.PageSize != 12 || nav.State.PageCount != 2 || len(nav.State.Rows) != 12 {
		t.Fatalf("large attendance page = %d rows, size %d, pages %d", len(nav.State.Rows), nav.State.PageSize, nav.State.PageCount)
	}
	if err := nav.Page(1); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 1 || nav.State.Rows[0].ID != "person-12" {
		t.Fatalf("last attendance page = %+v", nav.State.Rows)
	}
	if err := nav.Select("person-12"); err != nil || s.status != "present" {
		t.Fatalf("last student tap did not mark attendance: %v, %q", err, s.status)
	}
}
func (s *attendanceStore) SetAttendance(person, lesson, status string) error {
	s.person, s.lesson, s.status = person, lesson, status
	if s.err != nil {
		return s.err
	}
	if s.marks == nil {
		s.marks = map[string]string{}
	}
	s.marks[person] = status
	return nil
}
func (s *attendanceStore) MarkPastAttendanceAbsent(string, string) error {
	s.finalized = true
	return nil
}

func TestAttendanceControllerToggleAndFailure(t *testing.T) {
	nav := navigation.New()
	s := &attendanceStore{marks: map[string]string{}}
	c := New(nav, s)
	nav.Features[navigation.Attendance] = c
	nav.Push(navigation.Attendance, func(f *navigation.Frame) { f.Course = "course"; f.Lesson = "lesson" })
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 1 || nav.State.Rows[0].ID != "active" {
		t.Fatal("former enrollment was selectable")
	}
	s.err = errors.New("disk full")
	if nav.Select("active") != nil || nav.Frame().Screen != navigation.Attendance || nav.State.Error {
		t.Fatal("failed save was not handled quietly")
	}
	s.err = nil
	if err := nav.Select("active"); err != nil {
		t.Fatal(err)
	}
	if s.person != "active" || s.lesson != "lesson" || s.status != "present" || nav.Frame().Screen != navigation.Attendance {
		t.Fatal("wrong attendance save or destination")
	}
	if err := nav.Select("active"); err != nil {
		t.Fatal(err)
	}
	if s.status != "absent" || nav.State.Rows[0].Status != "absent" {
		t.Fatal("attendance toggle did not switch to absent")
	}
	if !s.finalized {
		t.Fatal("past attendance was not finalized")
	}
}
