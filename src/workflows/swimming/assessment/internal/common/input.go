// Package common contains assessment snapshots and shared presentation helpers.
package common

import (
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Input is the lesson snapshot supplied by the parent to a mode component.
type Input struct {
	State  navigation.State
	Frame  navigation.Frame
	Course domain.Course
	Roster []domain.Enrollment
}
