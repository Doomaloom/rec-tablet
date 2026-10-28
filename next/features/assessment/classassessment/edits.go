package classassessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func (c *Component) PrepareEdit(id string) (navigation.Edit, bool) {
	f := c.navigation.Frame()
	if f.Screen != navigation.ClassAssessment || f.Page < 0 || f.Page >= len(c.State.Attempts) {
		return navigation.Edit{}, false
	}
	attempt := c.State.Attempts[f.Page].AttemptID
	for i, row := range c.navigation.Rows() {
		if row.ID != id || row.Disabled {
			continue
		}
		buttonID := "criterion:" + id
		if !c.Layout.Panels {
			buttonID = "row:" + id
		}
		edit := common.ChecklistEdit(row, buttonID, true, c.Layout.Panels, func(status string) error { return c.storage.Mark(attempt, id, status) })
		if !c.Layout.Panels {
			edit.Button.Selected = false
			if edit.Row.Status != "successful" {
				edit.Button.Fill = ui.Blue
				if i%2 == 1 {
					edit.Button.Fill = ui.Navy
				}
			}
		}
		return edit, true
	}
	return navigation.Edit{}, false
}
