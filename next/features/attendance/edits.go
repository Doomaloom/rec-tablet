package attendance

import (
	"strings"

	"rectablet/next/contracts"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func (c *Component) PrepareEdit(id string) (navigation.Edit, bool) {
	f := c.navigation.Frame()
	if f.Screen != navigation.Attendance {
		return navigation.Edit{}, false
	}
	for _, row := range c.navigation.State.Rows {
		if row.ID != id || row.Disabled {
			continue
		}
		wasPresent := row.Status == "present"
		row.Status = "present"
		row.Fill = ui.Green
		if wasPresent {
			row.Status = "absent"
			row.Fill = ui.Amber
		}
		name := row.Label
		if i := strings.LastIndex(name, " · "); i >= 0 {
			name = name[:i]
		}
		row.Label = name + " · " + strings.ToUpper(row.Status[:1]) + row.Status[1:]
		status := row.Status
		return navigation.Edit{Row: row, Button: contracts.ButtonUpdate{ID: "row:" + id, Text: row.Label, Fill: row.Fill, Ink: ui.White}, Save: func() error { return c.store.SetAttendance(id, f.Lesson, status) }}, true
	}
	return navigation.Edit{}, false
}
