package feedback

import (
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
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
