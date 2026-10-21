package skilldetail

import (
	"testing"

	"rectablet/next/contracts"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

func TestSkillDetailUsesOnlyPinnedVersion(t *testing.T) {
	detail := SkillDetail{}
	state := contracts.State{PrimaryInfo: "Swimming"}
	entries := []domain.Feedback{
		{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "older", Remaining: []curriculum.Subskill{{Text: "Older criterion"}}},
		{Skill: curriculum.Skill{ID: "skill", Name: "Float"}, Version: "pinned", Remaining: []curriculum.Subskill{{Text: "Pinned criterion"}}},
	}
	rows := detail.Configure(&state, domain.Enrollment{Name: "Alex"}, entries, "skill", "pinned")
	if len(rows) != 1 || rows[0].Label != "Pinned criterion" || !rows[0].Disabled {
		t.Fatalf("pinned criteria = %+v", rows)
	}
	if state.PageSize != 1 || state.Title != "Alex" || state.HeaderAction != "Attempts" {
		t.Fatalf("Detail state = %+v", state)
	}
}
