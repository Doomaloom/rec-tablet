package classes

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if c.Frame().Screen == navigation.Classes {
		state.HoldTitle = func() { _ = c.Maintenance() }
	}
	switch c.Frame().Screen {
	case navigation.Classes:
		state.ComposeBody = c.composeClassListBody(*state)
	case navigation.Dates:
		state.ComposeBody = c.composeDateListBody(*state)
	case navigation.Class:
		state.ComposeBody = c.composeLessonMenuBody(*state)
	}
}

func (c *Controller) composeClassListBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeGridBody(state, 2, 3)
	}
	return c.composeRowListBody(state)
}

func (c *Controller) composeDateListBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeDateCalendarBody(state)
	}
	return c.composeRowListBody(state)
}

func (c *Controller) composeLessonMenuBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeGridBody(state, 3, 1)
	}
	return c.composeRowListBody(state)
}
