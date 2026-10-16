package setup

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
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
				var chooser dialog.Dialog
				choose := func(mode ui.DisplayMode) { chooser.Hide(); _ = c.SetDisplayMode(mode) }
				chooser = dialog.NewCustom("Display", "Cancel", container.NewVBox(
					widget.NewButton(ui.DisplayCompact.Label(), func() { choose(ui.DisplayCompact) }),
					widget.NewButton(ui.DisplayLarge.Label(), func() { choose(ui.DisplayLarge) }),
				), surface.Window)
				chooser.Show()
			} else if strings.HasPrefix(id, "profile:") {
				_ = c.Select(id)
			}
		})
	}
}
