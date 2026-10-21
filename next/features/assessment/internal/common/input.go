// Package common contains assessment snapshots and shared presentation helpers.
package common

import (
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// Input is the lesson snapshot supplied by the parent to a mode component.
type Input struct {
	State  navigation.State
	Frame  navigation.Frame
	Course domain.Course
	Roster []domain.Enrollment
}
