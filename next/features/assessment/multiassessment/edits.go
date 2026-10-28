package multiassessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func (c *Component) PrepareEdit(id string) (navigation.Edit, bool) {
	if c.navigation.Frame().Screen != navigation.MultiStudent || c.State.Edit == nil {
		return navigation.Edit{}, false
	}
	for i, row := range c.navigation.Rows() {
		if row.ID != id || row.Disabled {
			continue
		}
		edit := common.ChecklistEdit(row, "criterion:"+id, false, c.Layout.Panels, nil)
		if !c.Layout.Panels {
			edit.Button.ID = "row:" + id
			edit.Button.Selected = false
			if edit.Row.Status != "successful" {
				edit.Button.Fill = ui.Blue
				if i%2 == 1 {
					edit.Button.Fill = ui.Navy
				}
			}
		}
		// Multi's checklist remains an in-memory edit until Save student, as before.
		_ = c.ToggleSubskill(id)
		edit.Save = nil
		return edit, true
	}
	return navigation.Edit{}, false
}
