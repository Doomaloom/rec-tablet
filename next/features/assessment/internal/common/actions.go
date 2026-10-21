package common

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func SaveRailAction(run func() error) *contracts.RailAction {
	return &contracts.RailAction{Label: "Save", Run: run, Fill: ui.Green, Ink: ui.White, FontSize: 34, Icon: SaveActionIcon}
}

func DiscardRailAction(run func() error) *contracts.RailAction {
	return &contracts.RailAction{Label: "Discard", Run: run, Fill: ui.Red, Ink: ui.White, FontSize: 34, Icon: DiscardActionIcon}
}
