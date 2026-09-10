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
		if _, timed := benchmarkFor(c.Frame().Skill); timed {
			state.ComposeBody = c.composeTimedSelectionBody(*state)
		} else {
			state.ComposeBody = c.composeSwimmerListBody(*state)
		}
	case navigation.TimedLaps:
		state.ComposeBody = c.composeTimedLapsBody(*state)
	case navigation.TimedHeat:
		state.ComposeBody = c.composeTimedHeatBody(*state)
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
				rows[i].Detail = "Selected for heat"
			}
		}
		ui.List(surface, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeTimedLapsBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		display := ui.NewText(snapshot.Progress, 28, ui.Navy)
		display.Bold = true
		display.Center = true
		surface.AddNamed("timed-laps-display", display, func(w, h float32) (float32, float32, float32, float32) { return 100, 110, w - 200, 58 })
		digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
		for i, digit := range digits {
			digit, column, row := digit, i%5, i/5
			button := surface.Button("timed-digit:"+digit, digit, func() { _ = c.Controller.Select("timed-digit:" + digit) })
			button.FontSize = 26
			surface.AddNamed("timed-digit-"+digit, button, func(w, h float32) (float32, float32, float32, float32) {
				return 100 + float32(column)*((w-220)/5+10), 180 + float32(row)*68, (w - 260) / 5, 58
			})
		}
		clear := surface.Button("timed-clear", "Clear", func() { _ = c.Controller.Select("timed-clear") })
		clear.Fill = ui.Rail
		clear.Ink = ui.Blue
		clear.FontSize = 18
		surface.AddNamed("timed-clear", clear, func(w, h float32) (float32, float32, float32, float32) { return 100, 322, w - 200, 52 })
	}
}

func (c *Controller) composeTimedHeatBody(state shell.State) ui.BodyComposer {
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
