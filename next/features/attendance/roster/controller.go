// Package roster presents active swimmers and their attendance status.
package roster

import (
	"strings"

	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

// Component receives loaded roster data and emits swimmer selections.
type Component struct {
	Layout   Layout
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

// Configure prepares active swimmers for pagination by the parent feature.
func (c *Component) Configure(state *contracts.State, swimmers []domain.Enrollment, marks map[string]string) []contracts.Row {
	state.Title = "Attendance"
	state.Empty = "No enrolled swimmers."
	if c.Layout.Columns > 0 && c.Layout.Rows > 0 {
		state.PageSize = c.Layout.Columns * c.Layout.Rows
	}
	rows := make([]contracts.Row, 0, len(swimmers))
	for _, swimmer := range swimmers {
		if !swimmer.Active {
			continue
		}
		status := marks[swimmer.ID]
		if status == "" {
			status = "unmarked"
		}
		row := contracts.Row{
			ID:     swimmer.ID,
			Label:  swimmer.Name + " · " + strings.ToUpper(status[:1]) + status[1:],
			Status: status,
		}
		switch status {
		case "present":
			row.Fill = ui.Green
		case "absent":
			row.Fill = ui.Amber
		}
		rows = append(rows, row)
	}
	return rows
}

// Layout is selected by the feature root, independently of child rendering.
type Layout struct{ Columns, Rows int }
