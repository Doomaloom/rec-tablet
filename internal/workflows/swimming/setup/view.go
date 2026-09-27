package setup

import (
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
			} else if strings.HasPrefix(id, "profile:") {
				_ = c.Select(id)
			}
		})
	}
}
