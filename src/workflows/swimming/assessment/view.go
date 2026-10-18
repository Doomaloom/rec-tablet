package assessment

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if state.HeaderAction != "" {
		state.HeaderWidth = 134
	}
	switch c.Frame().Screen {
	case navigation.Skills:
		state.ComposeBody = c.sharedView().ComposeSkillListBody(*state)
	case navigation.SkillSwimmers:
		if _, timed := benchmarkFor(c.Frame().Skill); timed {
			state.ComposeBody = c.composeTimedSelectionBody(*state)
		} else {
			state.ComposeBody = c.composeSwimmerListBody(*state)
		}
	case navigation.TimedLaps:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.largeView().ComposeLargeTimedLapsBody(*state)
		} else {
			state.ComposeBody = c.smallView().ComposeTimedLapsBody(*state)
		}
	case navigation.TimedHeat:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.largeView().ComposeLargeTimedHeatBody(*state)
		} else {
			state.ComposeBody = c.smallView().ComposeTimedHeatBody(*state)
		}
	case navigation.Attempts:
		state.ComposeBody = c.sharedView().ComposeAttemptListBody(*state)
	case navigation.Assessment:
		if state.FooterAction == "" {
			state.ComposeBody = c.smallView().ComposeSubskillChecklistBody(*state)
		} else {
			state.ComposeBody = c.smallView().ComposeAssessmentBody(*state)
		}
	}
}

func (c *Controller) composeSwimmerListBody(state shell.State) ui.BodyComposer {
	if state.DisplayMode.IsLarge() {
		return c.largeView().ComposeLargeSwimmerListBody(state)
	}
	return c.sharedView().ComposeAssessmentListBody(state)
}

func (c *Controller) composeTimedSelectionBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		rows := append([]ui.Item(nil), snapshot.Rows...)
		for i := range rows {
			if rows[i].Status == "selected" {
				rows[i].Fill = ui.Green
				rows[i].Selected = true
			}
		}
		if surface.DisplayMode.IsLarge() {
			c.largeView().RenderLargeTimedSelection(surface, rows)
		} else {
			c.smallView().RenderSmallTimedSelection(surface, rows)
		}
	}
}
