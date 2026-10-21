package displaypicker

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

// DisplayPicker presents device layouts and emits a chosen display mode.
type DisplayPicker struct {
	OnSelect func(ui.DisplayMode) error
}

func (c *DisplayPicker) Row(mode ui.DisplayMode) contracts.Row {
	return contracts.Row{ID: "display", Label: "Display · " + mode.Label()}
}
