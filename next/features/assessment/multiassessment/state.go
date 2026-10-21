package multiassessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

type Group struct {
	Number, Capacity int
	Members          []string
	Attempts         []domain.MultiAttempt
}

type Session struct {
	Roster             []domain.Enrollment
	Skill              curriculum.Skill
	CourseInfo         string
	Groups             []Group
	AssignmentGroup    int
	GroupCount         int
	ProposedGroupCount int
}

type StudentEdit struct {
	Attempt domain.MultiAttempt
	Answers map[string]string
}

type studentSnapshot = common.MultiStudentSnapshot
