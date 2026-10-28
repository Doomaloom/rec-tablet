package skilldetail

import (
	"fmt"
	"strings"

	"rectablet/next/contracts"
	"rectablet/next/features/feedback/internal/formatting"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// SkillDetail presents remaining criteria and timed results for a pinned skill version.
type SkillDetail struct{}

func (c *SkillDetail) Configure(state *contracts.State, person domain.Enrollment, entries []domain.Feedback, skill, version string) []contracts.Row {
	state.Title = person.Name
	state.HeaderAction = "Attempts"
	state.HeaderWidth = 120
	state.Empty = "This skill is now complete."
	state.PageSize = 1
	rows := []contracts.Row{}
	for _, entry := range entries {
		if entry.Skill.ID != skill || entry.Version != version {
			continue
		}
		state.Progress = navigation.ProgressText(entry.Progress)
		navigation.AppendPrimaryInfo(state, entry.Skill.Name)
		state.Empty = fmt.Sprintf("Latest attempt successful. Needs %d more successful observation(s).", 3-entry.Successes)
		if entry.Timed {
			rows = append(rows, contracts.Row{Label: fmt.Sprintf("Timed heat · %d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatting.TimedElapsed(entry.TimedElapsedMS)), Disabled: true})
		}
		for _, subskill := range entry.Remaining {
			rows = append(rows, contracts.Row{Label: subskill.Text, Disabled: true})
		}
	}
	return rows
}
func (c *SkillDetail) Text(entry domain.Feedback) string {
	if entry.Progress.Complete {
		return "Complete · Three successful observations recorded."
	}
	parts := make([]string, 0, len(entry.Remaining)+2)
	if entry.Timed {
		parts = append(parts, fmt.Sprintf("Latest timed heat · %d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatting.TimedElapsed(entry.TimedElapsedMS)))
	}
	if len(entry.Remaining) > 0 {
		parts = append(parts, "Needs practice:")
		for _, criterion := range entry.Remaining {
			parts = append(parts, "• "+criterion.Text)
		}
	} else if !entry.Timed {
		parts = append(parts, fmt.Sprintf("Latest attempt successful. Needs %d more successful observation(s).", max(0, 3-entry.Successes)))
	}
	return strings.Join(parts, "\n")
}
