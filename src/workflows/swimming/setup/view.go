package setup

import (
	"fyne.io/fyne/v2/container"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"strings"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	state.ComposeBody = c.composeSetupBody(*state)
}

func (c *Controller) composeSetupBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) {
			v := &dialogs{Controller: c, Window: surface.Window}
			if id == "import" {
				v.importDialog()
			} else if id == "share" {
				v.shareDialog()
			} else if id == "server" {
				v.shareServerDialog()
			} else if id == "export" {
				v.exportDialog()
			} else if id == "display" {
				var chooser *modalshell.Modal
				choose := func(mode ui.DisplayMode) { chooser.Hide(); _ = c.SetDisplayMode(mode) }
				chooser = modalshell.New(surface.Window, container.NewVBox(
					actionbutton.New(ui.DisplayCompact.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayCompact) }),
					actionbutton.New(ui.DisplayLarge.Label(), actionbutton.Options{Variant: actionbutton.Outline}, func() { choose(ui.DisplayLarge) }),
				), modalshell.Options{Title: "Display", MaxWidth: 600})
				chooser.Show()
			} else if strings.HasPrefix(id, "profile:") {
				_ = c.Select(id)
			}
		})
	}
}
