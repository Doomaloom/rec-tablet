package assessment

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/shared"
)

func saveRailAction(run func() error) *shell.RailAction {
	return &shell.RailAction{Label: "Save", Run: run, Fill: ui.Green, Ink: ui.White, FontSize: 34, Icon: shared.SaveActionIcon}
}

func discardRailAction(run func() error) *shell.RailAction {
	return &shell.RailAction{Label: "Discard", Run: run, Fill: ui.Red, Ink: ui.White, FontSize: 34, Icon: shared.DiscardActionIcon}
}
