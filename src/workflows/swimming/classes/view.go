package classes

import (
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) composeFeatureBody(state *shell.State) {
	switch c.navigation.Frame().Screen {
	case navigation.Classes:
		state.HoldTitle = func() { _ = c.navigation.Maintenance() }
		state.ComposeBody = c.classList.Compose(*state)
	case navigation.Dates:
		state.ComposeBody = c.datePicker.Compose(*state)
	case navigation.Class:
		state.ComposeBody = c.lessonMenu.Compose(*state)
	}
}
