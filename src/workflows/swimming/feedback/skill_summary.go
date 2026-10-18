package feedback

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// SkillSummary owns the selected swimmer snapshot and presents incomplete skills.
type SkillSummary struct {
	selected   summaryData
	detail     *SkillDetail
	onSelect   func(string)
	onAttempts func(person, skill, version string)
}

type summaryData struct {
	person  domain.Enrollment
	entries []domain.Feedback
}

func (c *SkillSummary) Configure(state *shell.State, person domain.Enrollment, entries []domain.Feedback) []shell.Row {
	state.Title = person.Name
	state.Empty = "No assessed skills remain incomplete. Unassessed skills are not included here."
	rows := make([]shell.Row, 0, len(entries))
	for _, entry := range entries {
		label := entry.Skill.Name + " · " + navigation.ProgressText(entry.Progress)
		if entry.Timed {
			label += " · Last heat " + formatTimedElapsed(entry.TimedElapsedMS)
		}
		rows = append(rows, shell.Row{ID: entry.Skill.ID, Label: label})
	}
	return rows
}

func (c *SkillSummary) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.List(surface, state.Rows, c.onSelect)
	}
}
