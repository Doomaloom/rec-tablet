package feedback

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2/container"

	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// SkillDetail presents remaining criteria and timed results for a pinned skill version.
type SkillDetail struct{}

func (c *SkillDetail) Configure(state *shell.State, person domain.Enrollment, entries []domain.Feedback, skill, version string) []shell.Row {
	state.Title = person.Name
	state.HeaderAction = "Attempts"
	state.HeaderWidth = 120
	state.Empty = "This skill is now complete."
	state.PageSize = 1
	rows := []shell.Row{}
	for _, entry := range entries {
		if entry.Skill.ID != skill || entry.Version != version {
			continue
		}
		state.Progress = navigation.ProgressText(entry.Progress)
		navigation.AppendPrimaryInfo(state, entry.Skill.Name)
		state.Empty = fmt.Sprintf("Latest attempt successful. Needs %d more successful observation(s).", 3-entry.Successes)
		if entry.Timed {
			rows = append(rows, shell.Row{Label: fmt.Sprintf("Timed heat · %d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatTimedElapsed(entry.TimedElapsedMS)), Disabled: true})
		}
		for _, subskill := range entry.Remaining {
			rows = append(rows, shell.Row{Label: subskill.Text, Disabled: true})
		}
	}
	return rows
}

func (c *SkillDetail) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		renderDetail(surface, state.Rows[0])
	}
}

func (c *SkillDetail) Text(entry domain.Feedback) string {
	parts := make([]string, 0, len(entry.Remaining)+2)
	if entry.Timed {
		parts = append(parts, fmt.Sprintf("Latest timed heat · %d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatTimedElapsed(entry.TimedElapsedMS)))
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

func formatTimedElapsed(milliseconds int64) string {
	if milliseconds < 0 {
		milliseconds = 0
	}
	minutes := milliseconds / 60000
	seconds := (milliseconds / 1000) % 60
	tenths := (milliseconds % 1000) / 100
	return fmt.Sprintf("%02d:%02d.%d", minutes, seconds, tenths)
}

func renderDetail(surface *ui.Surface, row ui.Item) {

	box := card.New(container.NewWithoutLayout(), card.Panel)
	surface.AddNamed("feedback-detail-card", box, surface.BodyBounds)
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("feedback-detail-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		height -= 2 * ui.ButtonPadding
		return x + ui.ButtonPadding, y + ui.ButtonPadding, width - 2*ui.ButtonPadding, height
	})
}
