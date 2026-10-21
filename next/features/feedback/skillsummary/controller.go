package skillsummary

import (
	"rectablet/next/contracts"
	"rectablet/next/features/feedback/internal/formatting"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// SkillSummary owns the selected swimmer snapshot and presents incomplete skills.
type SkillSummary struct {
	Selected   Snapshot
	Detail     func(domain.Feedback) string
	OnSelect   func(string)
	OnAttempts func(person, skill, version string)
}
type Snapshot struct {
	Person  domain.Enrollment
	Entries []domain.Feedback
}

func (c *SkillSummary) Configure(state *contracts.State, person domain.Enrollment, entries []domain.Feedback) []contracts.Row {
	state.Title = person.Name
	state.Empty = "No assessed skills remain incomplete. Unassessed skills are not included here."
	rows := make([]contracts.Row, 0, len(entries))
	for _, entry := range entries {
		label := entry.Skill.Name + " · " + navigation.ProgressText(entry.Progress)
		if entry.Timed {
			label += " · Last heat " + formatting.TimedElapsed(entry.TimedElapsedMS)
		}
		rows = append(rows, contracts.Row{ID: entry.Skill.ID, Label: label})
	}
	return rows
}
