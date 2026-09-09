package classes

import (
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if c.Frame().Screen == navigation.Classes {
		state.HoldTitle = func() { _ = c.Maintenance() }
		state.TitleDetails = false
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
	return c.composeRowListBody(state)
}

func (c *Controller) composeDateListBody(state shell.State) ui.BodyComposer {
	return c.composeRowListBody(state)
}

func (c *Controller) composeLessonMenuBody(state shell.State) ui.BodyComposer {
	return c.composeRowListBody(state)
}

func (c *Controller) composeRowListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}
