package classassessment

import (
	"errors"
	"testing"

	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestClassRosterShowsSavedSuccessCountsForCurrentSkill(t *testing.T) {
	frame := navigation.Frame{Screen: navigation.ClassAssessment, Parameters: navigation.Parameters{Skill: "skill", Version: "version"}}
	counts := map[string]int{"a": 0, "b": 1, "c": 2, "d": 3, "e": 5, "f": -1}
	wantError := errors.New("progress unavailable")
	var progressError error
	c := New(Options{
		Navigation: common.Navigation{Frame: func() navigation.Frame { return frame }},
		Storage: Storage{
			Skill: func(string, string) (curriculum.Skill, error) { return curriculum.Skill{ID: "skill"}, nil },
			Draft: func(string) (map[string]string, error) { return nil, nil },
			Progress: func(person, skill, version string) (domain.Progress, error) {
				if skill != "skill" || version != "version" {
					t.Fatalf("wrong progress scope: %q %q", skill, version)
				}
				return domain.Progress{Successes: counts[person]}, progressError
			},
		},
	})
	c.Layout.Panels = true
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		c.State.Roster = append(c.State.Roster, domain.MultiAttempt{PersonID: id, Name: id})
	}
	c.State.Attempts = c.State.Roster
	c.State.Status = map[string]string{"e": "Saved"}
	input := common.Input{Frame: frame}
	if _, err := c.Build(input); err != nil {
		t.Fatal(err)
	}
	rows := c.rosterRows(0)
	for i, want := range []int{0, 1, 2, 3, 3, 0} {
		if rows[i].SuccessChecks != want {
			t.Fatalf("student %s: checks = %d, want %d", rows[i].ID, rows[i].SuccessChecks, want)
		}
	}
	if !rows[0].Selected || !rows[4].Disabled {
		t.Fatal("progress changed selection or saved status")
	}
	counts["a"] = 2
	if _, err := c.Build(input); err != nil {
		t.Fatal(err)
	}
	if c.rosterRows(0)[0].SuccessChecks != 2 {
		t.Fatal("progress did not refresh")
	}
	progressError = wantError
	if _, err := c.Build(input); !errors.Is(err, wantError) {
		t.Fatalf("progress error = %v, want %v", err, wantError)
	}
}
