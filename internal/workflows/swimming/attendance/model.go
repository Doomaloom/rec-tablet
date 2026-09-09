package attendance

import (
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	navigation.LessonReader
	GetEnrollment(string) (domain.Enrollment, error)
	Attendance(string) (map[string]string, error)
	SetAttendance(string, string, string) error
}
