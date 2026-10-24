package setup

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func (c *Component) composeFeatureBody(state *contracts.State) {
	state.ComposeBody = c.composeSetupBody(*state)
}

func (c *Component) composeSetupBody(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.Grid(surface, state.Rows, 1, max(1, state.PageSize), func(id string) {
			if c.navigation.Closed {
				return
			}
			switch id {
			case "export":
				c.fileTransfer.OpenExport(surface.Window)
			case "server":
				c.sessionSharing.OpenServer(surface.Window)
			case "display":
				c.displayPicker.Open(surface.Window)
			}
		})
	}
}
