package common

import (
	"rectablet/next/contracts"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

// ChecklistEdit keeps the normal and optimistic rendering styles identical.
func ChecklistEdit(row navigation.Row, buttonID string, class, panels bool, save func(string) error) navigation.Edit {
	status := "successful"
	if row.Status == status {
		status = "unmarked"
	}
	row.Status = status
	update := contracts.ButtonUpdate{ID: buttonID, Text: row.Label, Fill: ui.Blue, Ink: ui.White}
	if panels && !class {
		update.Fill, update.Ink = ui.Rail, ui.Navy
	}
	if !panels && class {
		update.Fill = ui.Navy
	}
	if status == "successful" {
		update.Fill, update.Ink, update.Selected = ui.Green, ui.White, true
	}
	return navigation.Edit{Row: row, Button: update, Save: func() error { return save(status) }}
}
