package classes

import (
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	Dates() ([]string, error)
	ListCourses(string) ([]domain.Course, error)
	navigation.LessonReader
}
