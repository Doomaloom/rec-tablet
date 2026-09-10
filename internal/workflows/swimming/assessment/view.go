package assessment

import (
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if state.HeaderAction != "" {
		state.HeaderWidth = 134
	}
	switch c.Frame().Screen {
	case navigation.Skills:
		state.ComposeBody = c.composeSkillListBody(*state)
	case navigation.SkillSwimmers:
		state.ComposeBody = c.composeSwimmerListBody(*state)
	case navigation.Attempts:
		state.ComposeBody = c.composeAttemptListBody(*state)
	case navigation.Assessment:
		if state.FooterAction == "" {
			state.ComposeBody = c.composeSubskillChecklistBody(*state)
		} else {
			state.ComposeBody = c.composeAssessmentBody(*state)
		}
	}
}

func (c *Controller) composeSkillListBody(state shell.State) ui.BodyComposer {
	return c.composeAssessmentListBody(state)
}

func (c *Controller) composeSwimmerListBody(state shell.State) ui.BodyComposer {
	return c.composeAssessmentListBody(state)
}

func (c *Controller) composeAttemptListBody(state shell.State) ui.BodyComposer {
	return c.composeAssessmentListBody(state)
}

func (c *Controller) composeAssessmentListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeAssessmentBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		renderCriterion(surface, snapshot.Rows[0], func(status string) { _ = c.Controller.Mark(status) })
	}
}

func (c *Controller) composeSubskillChecklistBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		rows := append([]ui.Item(nil), snapshot.Rows...)
		for i := range rows {
			if rows[i].Status == "successful" {
				rows[i].Fill = ui.Green
				if rows[i].Detail == "" {
					rows[i].Detail = "Complete"
				} else {
					rows[i].Detail = "Complete · " + rows[i].Detail
				}
			}
		}
		ui.List(surface, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}
