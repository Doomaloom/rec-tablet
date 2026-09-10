// Package domain owns swimming concepts and rules, independently of SQL and Fyne.
package domain

import "rectablet/internal/workflows/swimming/curriculum"

type Course struct{ ID, Name, Time, Location, LessonID, CurriculumVersion string }
type Enrollment struct {
	ID, Name, Level, CourseID, CurriculumVersion string
	Active                                       bool
}
type Attempt struct {
	ID, EnrollmentID, LessonID, SkillID, CurriculumVersion, CreatedAt, Date string
	SubmittedAt                                                             *string
	Success, Editing                                                        bool
}
type Lesson struct{ ID, Date string }
type Progress struct {
	Assessed, Successes int
	Complete            bool
}
type Feedback struct {
	Skill   curriculum.Skill
	Version string
	Progress
	Remaining      []curriculum.Subskill
	Timed          bool
	TimedElapsedMS int64
	TimedLaps      int
	TimedRequired  int
}

// TimedParticipant is one swimmer in a shared timed heat.
type TimedParticipant struct {
	AttemptID, EnrollmentID, PersonID, Name string
	LapsRequired, LapsCompleted             int
	ElapsedMS                               int64
}

// TimedResult is the final measurement saved for one heat participant.
type TimedResult struct {
	AttemptID                   string
	LapsRequired, LapsCompleted int
	ElapsedMS                   int64
	Success                     bool
}
