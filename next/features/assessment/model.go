package assessment

import (
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// Repository lists only the data operations this feature consumes.
type Repository interface {
	navigation.LessonReader
	Attendance(string) (map[string]string, error)
	ReplaceDraftAnswers(string, map[string]string) error
	GetEnrollment(string) (domain.Enrollment, error)
	CatalogFor(string) (curriculum.Catalog, error)
	Skill(string, string) (curriculum.Skill, error)
	Progress(string, string, string) (domain.Progress, error)
	ListAttempts(string, string, string) ([]domain.Attempt, error)
	GetAttempt(string) (domain.Attempt, error)
	Draft(string) (map[string]string, error)
	NewAttempt(string, string, string) (string, error)
	EditAttempt(string) error
	Mark(string, string, string) error
	Submit(string) (bool, error)
	DiscardAttempt(string) error
	DiscardAttempts([]string) error
	RemoveLatestSuccess(string, string, string) error
	OpenMultiAttempts([]string, string, string) ([]domain.MultiAttempt, error)
	SubmitAttempts([]string) error
	DiscardMultiAttempts([]domain.MultiAttempt) error
	NewTimedHeat([]string, string, string, int) ([]domain.TimedParticipant, error)
	RecordTimedLap(string) (int, error)
	ResetTimedHeat([]string) error
	SaveTimedHeat([]domain.TimedResult) error
	DiscardTimedHeat([]string) error
}
