package assessment

import (
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	navigation.LessonReader
	GetEnrollment(string) (domain.Enrollment, error)
	CatalogFor(string) (curriculum.Catalog, error)
	Skill(string, string) (curriculum.Skill, error)
	Progress(string, string, string) (domain.Progress, error)
	ListAttempts(string, string, string) ([]domain.Attempt, error)
	GetAttempt(string) (domain.Attempt, error)
	GetLesson(string) (domain.Lesson, error)
	Draft(string) (map[string]string, error)
	NewAttempt(string, string, string) (string, error)
	EditAttempt(string) error
	Mark(string, string, string) error
	Submit(string) (bool, error)
	RemoveLatestSuccess(string, string, string) error
}
