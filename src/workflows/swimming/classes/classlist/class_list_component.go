// Package classlist renders the scheduled class list.
package classlist

import (
	"fmt"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
)

// Component has a selection callback and receives a snapshot to render.
type Component struct {
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

func (c *Component) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		if state.DisplayMode.IsLarge() {
			ui.Grid(surface, state.Rows, 2, 3, c.onSelect)
			return
		}
		ui.List(surface, state.Rows, c.onSelect)
	}
}

// Configure builds the class-list presentation from loaded course and date data.
func (c *Component) Configure(state *shell.State, courses []domain.Course, dates []string, selected string) []shell.Row {
	state.HeaderAction = "Date"
	if state.DisplayMode.IsLarge() {
		state.PageSize = 6
	}
	for i, date := range dates {
		if date == selected {
			state.SplitHeader = &shell.SplitHeader{Left: state.Title, Right: fmt.Sprintf("Class %d/%d", i+1, len(dates))}
			break
		}
	}
	state.Empty = "No classes scheduled. Choose Date to view other lessons."
	if len(dates) == 0 {
		state.Empty = "No classes loaded. Hold My Classes for 3 seconds to import a class package."
	}
	rows := make([]shell.Row, 0, len(courses))
	for _, course := range courses {
		rows = append(rows, shell.Row{ID: course.ID, Label: course.Name + " · " + course.Time})
	}
	return rows
}
