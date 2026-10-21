package swimmerpicker

import (
	"fmt"

	"rectablet/next/contracts"
	"rectablet/src/workflows/swimming/domain"
)

// SwimmerPicker presents current and former swimmers and emits selections.
type SwimmerPicker struct {
	Layout   Layout
	OnSelect func(string)
}

func (c *SwimmerPicker) Configure(state *contracts.State, roster []domain.Enrollment, feedback map[string][]domain.Feedback, selected string) []contracts.Row {
	state.Title = "Feedback"
	state.Empty = "No swimmers loaded."
	if c.Layout.Sidebar {
		state.PageSize = 6
	}
	rows := make([]contracts.Row, 0, len(roster))
	for _, person := range roster {
		unit := "skills"
		if len(feedback[person.ID]) == 1 {
			unit = "skill"
		}
		label := fmt.Sprintf("%s · %d %s in progress", person.Name, len(feedback[person.ID]), unit)
		if c.Layout.Sidebar {
			label = person.Name
		}
		if !person.Active {
			label += " · Former enrollment"
		}
		rows = append(rows, contracts.Row{ID: person.ID, Label: label, Selected: c.Layout.Sidebar && person.ID == selected})
	}
	return rows
}

type Layout struct{ Sidebar bool }
