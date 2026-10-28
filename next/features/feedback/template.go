package feedback

import (
	"rectablet/next/contracts"
	"rectablet/next/features/feedback/skillsummary"
	"rectablet/next/features/feedback/swimmerpicker"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func (c *Component) composeFeatureBody(state *contracts.State) {
	switch c.navigation.Frame().Screen {
	case navigation.FeedbackSwimmers:
		if c.combined {
			state.ComposeBody = c.composeLargeFeedbackBody(*state, c.skillSummary.Selected)
		} else {
			state.ComposeBody = c.swimmerPicker.Compose(*state)
		}
	case navigation.Feedback:
		state.ComposeBody = c.skillSummary.Compose(*state)
	case navigation.FeedbackDetail:
		state.ComposeBody = c.skillDetail.Compose(*state)
	}
}

const feedbackSidebarFraction float32 = .28

func feedbackSidebarBounds(surface *ui.Surface) ui.Bounds {
	return func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, ui.OuterMargin, width * feedbackSidebarFraction, height + y - ui.OuterMargin
	}
}

func feedbackDetailBounds(surface *ui.Surface) ui.Bounds {
	return func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		left := width * feedbackSidebarFraction
		return x + left + 2*ui.Gap, y, width - left - 2*ui.Gap, height
	}
}

func (c *Component) composeLargeFeedbackBody(state contracts.State, data skillsummary.Snapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		left := ui.Panel(surface, feedbackSidebarBounds(surface))
		c.swimmerPicker.RenderSidebar(left, state)
		right := ui.Panel(surface, feedbackDetailBounds(surface))
		c.skillSummary.ComposeLarge(data, state.Page)(right)
	}
}

func (c *Component) configureLayout(mode ui.DisplayMode) {
	c.combined = mode.IsLarge()
	c.swimmerPicker.Layout = swimmerpicker.Layout{Sidebar: c.combined}
}
