package feedback

import (
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	navigation.LessonReader
	GetEnrollment(string) (domain.Enrollment, error)
	Feedback(string) ([]domain.Feedback, error)
}
