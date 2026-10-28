// Package domain owns swimming concepts and rules, independently of SQL and Fyne.
package domain

import "rectablet/src/workflows/swimming/curriculum"

type Course struct{ ID, Name, Level, Time, Location, LessonID, CurriculumVersion string }
type Enrollment struct {
	ID, Name, Level, CourseID, CurriculumVersion string
	Active                                       bool
}
type Attempt struct {
	ID, EnrollmentID, LessonID, SkillID, CurriculumVersion, CreatedAt, Date string
	SubmittedAt                                                             *string
	Success, Editing                                                        bool
}

// MultiAttempt describes one attempt opened for a multi-student assessment.
// OriginalAnswers is used to restore an existing unfinished draft on discard.
type MultiAttempt struct {
	AttemptID, PersonID, Name string
	OriginalAnswers           map[string]string
	Created                   bool
}
type Progress struct {
	Assessed, Successes int
	Complete            bool
}

// SubskillMiss counts saved attempts whose current result is not yet successful.
type SubskillMiss struct {
	Subskill curriculum.Subskill
	Count    int
}

type Feedback struct {
	Skill   curriculum.Skill
	Version string
	Progress
	MissedSubskills []SubskillMiss
	Remaining       []curriculum.Subskill
	Timed           bool
	TimedElapsedMS  int64
	TimedLaps       int
	TimedRequired   int
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
