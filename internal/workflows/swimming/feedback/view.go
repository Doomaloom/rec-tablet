package feedback

import (
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if state.HeaderAction != "" {
		state.HeaderWidth = 120
	}
	switch c.Frame().Screen {
	case navigation.FeedbackSwimmers:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.composeLargeFeedbackBody(*state, c.large)
		} else {
			state.ComposeBody = c.composeFeedbackSwimmerBody(*state)
		}
	case navigation.Feedback:
		state.ComposeBody = c.composeFeedbackSkillBody(*state)
	case navigation.FeedbackDetail:
		state.ComposeBody = c.composeFeedbackDetailBody(*state)
	}
}

func (c *Controller) composeFeedbackSwimmerBody(state shell.State) ui.BodyComposer {
	compose := c.composeFeedbackListBody(state)
	return func(surface *ui.Surface) { compose(ui.SelectableSurface(surface)) }
}

func (c *Controller) composeFeedbackSkillBody(state shell.State) ui.BodyComposer {
	return c.composeFeedbackListBody(state)
}

func (c *Controller) composeFeedbackListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeFeedbackDetailBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		renderDetail(surface, snapshot.Rows[0])
	}
}
