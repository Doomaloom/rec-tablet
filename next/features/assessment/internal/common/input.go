// Package common contains assessment snapshots and shared presentation helpers.
package common

import (
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// Input is the lesson snapshot supplied by the parent to a mode component.
type Input struct {
	State      navigation.State
	Frame      navigation.Frame
	Course     domain.Course
	Roster     []domain.Enrollment
	Attendance map[string]string
}

// PresentRoster excludes absent and unmarked students for this lesson.
func (input Input) PresentRoster() []domain.Enrollment {
	var roster []domain.Enrollment
	for _, student := range input.Roster {
		if student.Active && input.Attendance[student.ID] == "present" {
			roster = append(roster, student)
		}
	}
	return roster
}
