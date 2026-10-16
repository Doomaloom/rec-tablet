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
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.composeLargeTimedLapsBody(*state)
		} else {
			state.ComposeBody = c.composeTimedLapsBody(*state)
		}
	case navigation.TimedHeat:
		if state.DisplayMode.IsLarge() {
			state.ComposeBody = c.composeLargeTimedHeatBody(*state)
		} else {
			state.ComposeBody = c.composeTimedHeatBody(*state)
		}
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
	if state.DisplayMode.IsLarge() {
		return func(surface *ui.Surface) {
			c.largePagedGrid(surface, "large-return-swimmers", "Swimmers", state.Rows, 2, 6, &c.largeRosterPage, func(id string) { _ = c.Controller.Select(id) })
		}
	}
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
				rows[i].Selected = true
			}
		}
		if surface.DisplayMode.IsLarge() {
			c.largePagedGrid(surface, "timed-selection", "Swimmers in heat", rows, 2, 6, &c.largeRosterPage, func(id string) { _ = c.Controller.Select(id) })
		} else {
			ui.List(surface, rows, func(id string) { _ = c.Controller.Select(id) })
		}
	}
}

func (c *Controller) composeTimedLapsBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		display := ui.NewText(snapshot.Progress, 28, ui.Navy)
		display.Bold = true
		display.Center = true
		surface.AddNamed("timed-laps-display", display, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x, y, width, 58
		})
		digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
		for i, digit := range digits {
			digit, column, row := digit, i%5, i/5
			button := surface.Button("timed-digit:"+digit, digit, func() { _ = c.Controller.Select("timed-digit:" + digit) })
			button.FontSize = 26
			surface.AddNamed("timed-digit-"+digit, button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, _ := surface.BodyBounds(w, h)
				digitWidth := (width - 4*ui.Gap) / 5
				return x + float32(column)*(digitWidth+ui.Gap), y + 58 + ui.Gap + float32(row)*(58+ui.Gap), digitWidth, 58
			})
		}
		clear := surface.Button("timed-clear", "Clear", func() { _ = c.Controller.Select("timed-clear") })
		clear.Fill = ui.Rail
		clear.Ink = ui.Blue
		clear.FontSize = 18
		surface.AddNamed("timed-clear", clear, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x, y + 3*(58+ui.Gap), width, ui.HeaderHeight
		})
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
				rows[i].Selected = true
			}
		}
		ui.List(surface, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}
