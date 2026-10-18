package classes

import (
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	Dates() ([]string, error)
	ListCourses(string) ([]domain.Course, error)
	navigation.LessonReader
}
