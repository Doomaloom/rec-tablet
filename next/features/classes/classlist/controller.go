// Package classlist renders the scheduled class list.
package classlist

import (
	"fmt"

	"rectablet/next/contracts"
	"rectablet/src/workflows/swimming/domain"
)

// Component has a selection callback and receives a snapshot to render.
type Component struct {
	Layout   Layout
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

// Configure builds the class-list presentation from loaded course and date data.
func (c *Component) Configure(state *contracts.State, courses []domain.Course, dates []string, selected string) []contracts.Row {
	state.HeaderAction = "Date"
	if c.Layout.Columns > 0 && c.Layout.Rows > 0 {
		state.PageSize = c.Layout.Columns * c.Layout.Rows
	}
	for i, date := range dates {
		if date == selected {
			state.SplitHeader = &contracts.SplitHeader{Left: state.Title, Right: fmt.Sprintf("Class %d/%d", i+1, len(dates))}
			break
		}
	}
	state.Empty = "No classes scheduled. Choose Date to view other lessons."
	if len(dates) == 0 {
		state.Empty = "No classes available for this instructor."
	}
	rows := make([]contracts.Row, 0, len(courses))
	for _, course := range courses {
		rows = append(rows, contracts.Row{ID: course.ID, Label: course.Name + " · " + course.Time})
	}
	return rows
}

// Layout is selected by the feature root, independently of child rendering.
type Layout struct{ Columns, Rows int }
