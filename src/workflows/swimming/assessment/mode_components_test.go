package assessment

import (
	"testing"

	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// This component can prepare a checklist and emit actions without a repository
// implementation or a navigation controller.
func TestSingleAssessmentWithExplicitDataAndActions(t *testing.T) {
	var saved, discarded bool
	component := SingleAssessment{
		data: singleAssessmentData{
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
		actions: singleAssessmentActions{
			SaveAttempt:    func() error { saved = true; return nil },
			DiscardAttempt: func() error { discarded = true; return nil },
		},
		paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows; state.RowCount = len(rows) },
		small:    &small.View{},
	}
	state, err := component.build(modeInput{frame: navigation.Frame{Screen: navigation.Assessment, Parameters: navigation.Parameters{Attempt: "draft"}}})
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

func TestTimedHeatOwnsTickerCleanup(t *testing.T) {
	stop := make(chan struct{})
	component := TimedHeat{
		heat:     &timedHeatState{SkillID: "Splash7:10", stop: stop},
		paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows; state.RowCount = len(rows) },
		small:    &small.View{},
	}
	state, err := component.build(modeInput{frame: navigation.Frame{Screen: navigation.TimedHeat}})
	if err != nil {
		t.Fatal(err)
	}
	if state.OnLeave == nil {
		t.Fatal("heat has no cleanup callback")
	}
	state.OnLeave()
	select {
	case <-stop:
	default:
		t.Fatal("leaving heat did not stop its ticker")
	}
	state.OnLeave() // Leaving again is harmless.
}
