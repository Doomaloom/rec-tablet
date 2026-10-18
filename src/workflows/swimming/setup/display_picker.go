package setup

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

// DisplayPicker presents device layouts and emits a chosen display mode.
type DisplayPicker struct {
	onSelect func(ui.DisplayMode) error
}

func (c *DisplayPicker) Row(mode ui.DisplayMode) shell.Row {
	return shell.Row{ID: "display", Label: "Display · " + mode.Label()}
}

func (c *DisplayPicker) Open(window fyne.Window) {
	var chooser *modalshell.Modal
	choose := func(mode ui.DisplayMode) { chooser.Hide(); _ = c.onSelect(mode) }
	chooser = modalshell.New(window, container.NewVBox(
		actionbutton.New(ui.DisplayCompact.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayCompact) }),
		actionbutton.New(ui.DisplayLarge.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayLarge) }),
	), modalshell.Options{Title: "Display", MaxWidth: 600})
	chooser.Show()
}
