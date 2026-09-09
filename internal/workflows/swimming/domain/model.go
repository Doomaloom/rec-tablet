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
	Remaining []curriculum.Subskill
}
