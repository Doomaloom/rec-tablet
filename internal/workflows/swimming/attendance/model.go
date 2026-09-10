package attendance

import (
	"rectablet/internal/workflows/swimming/navigation"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	navigation.LessonReader
	Attendance(string) (map[string]string, error)
	SetAttendance(string, string, string) error
	MarkPastAttendanceAbsent(string, string) error
}
