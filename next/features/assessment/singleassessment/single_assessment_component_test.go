package singleassessment

import (
	"testing"

	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// This component can prepare a checklist and emit actions without a repository
// implementation or a navigation controller.
func TestSingleAssessmentWithExplicitDataAndActions(t *testing.T) {
	var saved, discarded bool
	component := Component{
		storage: Storage{
			Submit:         func(string) (bool, error) { saved = true; return false, nil },
			DiscardAttempt: func(string) error { discarded = true; return nil },
			GetAttempt: func(id string) (domain.Attempt, error) {
				if id != "draft" {
					t.Fatalf("attempt requested = %q", id)
				}
				return domain.Attempt{EnrollmentID: "student", SkillID: "skill", CurriculumVersion: "v1"}, nil
			},
			GetEnrollment: func(string) (domain.Enrollment, error) {
				return domain.Enrollment{Name: "Alex"}, nil
			},
			Skill: func(version, skill string) (curriculum.Skill, error) {
				if version != "v1" || skill != "skill" {
					t.Fatalf("skill requested = %q / %q", version, skill)
				}
				return curriculum.Skill{Name: "Float", Subskills: []curriculum.Subskill{{ID: "a", Text: "First criterion"}, {ID: "b", Text: "Second criterion"}}}, nil
			},
			Draft: func(string) (map[string]string, error) {
				return map[string]string{"a": "successful"}, nil
			},
		},
		navigation: common.Navigation{
			Frame: func() navigation.Frame {
				return navigation.Frame{Screen: navigation.Assessment, Parameters: navigation.Parameters{Attempt: "draft"}}
			},
			Run: func(run func() error) error { return run() },
			Pop: func() {},
		},
		paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows; state.RowCount = len(rows) },
		small:    &CompactTemplate{},
	}
	state, err := component.Build(common.Input{Frame: navigation.Frame{Screen: navigation.Assessment, Parameters: navigation.Parameters{Attempt: "draft"}}})
	if err != nil {
		t.Fatal(err)
	}
	if state.Title != "Alex" || len(state.Rows) != 2 || state.Rows[0].Status != "successful" || state.Rows[1].Status != "unmarked" || state.ComposeBody == nil {
		t.Fatalf("checklist = %+v", state)
	}
	if err := state.HomeRailAction.Run(); err != nil {
		t.Fatal(err)
	}
	if err := state.BackRailAction.Run(); err != nil {
		t.Fatal(err)
	}
	if !saved || !discarded {
		t.Fatalf("actions: saved=%v discarded=%v", saved, discarded)
	}
}
