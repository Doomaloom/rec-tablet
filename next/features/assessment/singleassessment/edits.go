package singleassessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
)

func (c *Component) PrepareEdit(id string) (navigation.Edit, bool) {
	f := c.navigation.Frame()
	if f.Screen != navigation.Assessment || f.Review {
		return navigation.Edit{}, false
	}
	for _, row := range c.navigation.Rows() {
		if row.ID != id || row.Disabled {
			continue
		}
		buttonID := "row:" + id
		if c.Layout.Panels {
			buttonID = "criterion:" + id
		}
		return common.ChecklistEdit(row, buttonID, false, c.Layout.Panels, func(status string) error { return c.storage.Mark(f.Attempt, id, status) }), true
	}
	return navigation.Edit{}, false
}
