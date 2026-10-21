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
			switch id {
			case "import":
				c.fileTransfer.OpenImport(surface.Window)
			case "export":
				c.fileTransfer.OpenExport(surface.Window)
			case "share":
				c.sessionSharing.Open(surface.Window)
			case "server":
				c.sessionSharing.OpenServer(surface.Window)
			case "display":
				c.displayPicker.Open(surface.Window)
			default:
				_ = c.profilePicker.Select(id)
			}
		})
	}
}
