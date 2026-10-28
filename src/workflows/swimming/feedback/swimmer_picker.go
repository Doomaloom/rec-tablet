package feedback

import (
	"fmt"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
)

// SwimmerPicker presents current and former swimmers and emits selections.
type SwimmerPicker struct {
	onSelect func(string)
}

func (c *SwimmerPicker) Configure(state *shell.State, roster []domain.Enrollment, feedback map[string][]domain.Feedback, selected string) []shell.Row {
	state.Title = "Feedback"
	state.Empty = "No swimmers loaded."
	rows := make([]shell.Row, 0, len(roster))
	for _, person := range roster {
		unit := "skills"
		if len(feedback[person.ID]) == 1 {
			unit = "skill"
		}
		label := fmt.Sprintf("%s · %d %s in progress", person.Name, len(feedback[person.ID]), unit)
		if state.DisplayMode.IsLarge() {
			label = person.Name
		}
		if !person.Active {
			label += " · Former enrollment"
		}
		rows = append(rows, shell.Row{ID: person.ID, Label: label, Selected: state.DisplayMode.IsLarge() && person.ID == selected})
	}
	return rows
}

func (c *SwimmerPicker) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.List(ui.SelectableSurface(surface), state.Rows, c.onSelect)
	}
}

// RenderSidebar receives the panel owned by the parent's combined layout.
func (c *SwimmerPicker) RenderSidebar(surface *ui.Surface, state shell.State) {
	sidebar := *surface
	sidebar.Button = func(id, label string, run func()) *ui.TouchButton {
		button := surface.Button(id, label, run)
		button.Selectable = true
		button.FontSize = 18
		button.MinimumFont = 16
		return button
	}
	ui.Grid(&sidebar, state.Rows, 1, 12, c.onSelect)
}
