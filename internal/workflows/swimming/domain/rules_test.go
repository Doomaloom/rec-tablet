package domain

import (
	"rectablet/internal/workflows/swimming/curriculum"
	"testing"
)

func TestAssessmentRequiresEveryCriterionTogether(t *testing.T) {
	skill := curriculum.Skill{Subskills: []curriculum.Subskill{{ID: "a"}, {ID: "b"}}}
	for _, tt := range []struct {
		name             string
		answers          map[string]string
		success, invalid bool
	}{
		{"incomplete", map[string]string{"a": "successful"}, false, true},
		{"wrong criterion", map[string]string{"a": "successful", "other": "successful"}, false, true},
		{"unmarked", map[string]string{"a": "successful", "b": "unmarked"}, false, true},
		{"failed", map[string]string{"a": "successful", "b": "not_yet"}, false, false},
		{"successful", map[string]string{"a": "successful", "b": "successful"}, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			success, err := Assess(skill, tt.answers)
			if success != tt.success || (err != nil) != tt.invalid {
				t.Fatalf("got %v, %v", success, err)
			}
		})
	}
}
func TestProgressAndFeedbackRules(t *testing.T) {
	for successes := 0; successes <= 4; successes++ {
		p := CalculateProgress(7, successes)
		if p.Complete != (successes >= 3) || p.Assessed != 7 || p.Successes != successes {
			t.Fatalf("bad progress: %+v", p)
		}
	}
	skill := curriculum.Skill{Subskills: []curriculum.Subskill{{ID: "a"}, {ID: "b"}}}
	remaining := Remaining(skill, map[string]string{"a": "successful", "b": "not_yet"})
	if len(remaining) != 1 || remaining[0].ID != "b" {
		t.Fatalf("remaining: %+v", remaining)
	}
}
