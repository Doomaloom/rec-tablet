package attendance

import (
	"rectablet/next/contracts"
	"rectablet/next/features/attendance/roster"
	"rectablet/src/shared/ui"
)

func (c *Component) configureLayout(mode ui.DisplayMode) {
	c.roster.Layout = roster.Layout{Columns: 1, Rows: 3}
	if mode.IsLarge() {
		c.roster.Layout = roster.Layout{Columns: 2, Rows: 6}
	}
}
func (c *Component) composeFeatureBody(state *contracts.State) {
	state.ComposeBody = c.roster.Compose(*state)
}
