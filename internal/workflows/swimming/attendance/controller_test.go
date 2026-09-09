package attendance

import (
	"errors"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
)

type attendanceStore struct {
	err                    error
	person, lesson, status string
}

func (s *attendanceStore) GetCourse(string) (domain.Course, error) {
	return domain.Course{Name: "Splash 1"}, nil
}
func (s *attendanceStore) ListRoster(string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{{ID: "active", Name: "A", Active: true}, {ID: "former", Name: "B"}}, nil
}
func (s *attendanceStore) GetEnrollment(string) (domain.Enrollment, error) {
	return domain.Enrollment{Name: "A"}, nil
}
func (s *attendanceStore) Attendance(string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (s *attendanceStore) SetAttendance(person, lesson, status string) error {
	s.person, s.lesson, s.status = person, lesson, status
	return s.err
}

func TestAttendanceControllerSaveAndFailure(t *testing.T) {
	nav := navigation.New()
	s := &attendanceStore{}
	c := New(nav, s)
	nav.Features[navigation.Attendance] = c
	nav.Features[navigation.AttendanceMark] = c
	nav.Push(navigation.Attendance, func(f *navigation.Frame) { f.Course = "course"; f.Lesson = "lesson" })
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	if len(nav.State.Rows) != 1 || nav.State.Rows[0].ID != "active" {
		t.Fatal("former enrollment was selectable")
	}
	if err := nav.Select("active"); err != nil {
		t.Fatal(err)
	}
	s.err = errors.New("disk full")
	if nav.Select("present") == nil || nav.Frame().Screen != navigation.AttendanceMark || !nav.State.Error {
		t.Fatal("failed save navigated away or hid the error")
	}
	s.err = nil
	if err := nav.Select("present"); err != nil {
		t.Fatal(err)
	}
	if s.person != "active" || s.lesson != "lesson" || s.status != "present" || nav.Frame().Screen != navigation.Attendance {
		t.Fatal("wrong attendance save or destination")
	}
}
