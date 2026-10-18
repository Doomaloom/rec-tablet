// Package roster presents active swimmers and their attendance status.
package roster

import (
	"strings"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
)

// Component receives loaded roster data and emits swimmer selections.
type Component struct {
	onSelect func(string)
}

func New(onSelect func(string)) *Component {
	return &Component{onSelect: onSelect}
}

// Configure prepares active swimmers for pagination by the parent feature.
func (c *Component) Configure(state *shell.State, swimmers []domain.Enrollment, marks map[string]string) []shell.Row {
	state.Title = "Attendance"
	state.Empty = "No enrolled swimmers."
	if state.DisplayMode.IsLarge() {
		state.PageSize = 12
	}
	rows := make([]shell.Row, 0, len(swimmers))
	for _, swimmer := range swimmers {
		if !swimmer.Active {
			continue
		}
		status := marks[swimmer.ID]
		if status == "" {
			status = "unmarked"
		}
		row := shell.Row{
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
