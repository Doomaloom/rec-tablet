package displaypicker

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/ui"
)

func (c *DisplayPicker) Open(window fyne.Window) {
	var chooser *modalshell.Modal
	choose := func(mode ui.DisplayMode) { chooser.Hide(); _ = c.OnSelect(mode) }
	chooser = modalshell.New(window, container.NewVBox(
		actionbutton.New(ui.DisplayCompact.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayCompact) }),
		actionbutton.New(ui.DisplayLarge.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayLarge) }),
	), modalshell.Options{Title: "Display", MaxWidth: 600})
	chooser.Show()
}
