package small

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/shared"
)

func (c *View) ComposeTimedLapsBody(state shell.State) ui.BodyComposer {
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
			button := surface.Button("timed-digit:"+digit, digit, func() { _ = c.Actions.Select("timed-digit:" + digit) })
			button.FontSize = 26
			surface.AddNamed("timed-digit-"+digit, button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, _ := surface.BodyBounds(w, h)
				digitWidth := (width - 4*ui.Gap) / 5
				return x + float32(column)*(digitWidth+ui.Gap), y + 58 + ui.Gap + float32(row)*(58+ui.Gap), digitWidth, 58
			})
		}
		clear := surface.Button("timed-clear", "Clear", func() { _ = c.Actions.Select("timed-clear") })
		clear.Fill = ui.Rail
		clear.Ink = ui.Blue
		clear.FontSize = 18
		surface.AddNamed("timed-clear", clear, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x, y + 3*(58+ui.Gap), width, ui.HeaderHeight
		})
	}
}

func (c *View) ComposeTimedHeatBody(state shell.State) ui.BodyComposer {
	return (&shared.View{Actions: c.Actions}).ComposeAssessmentListBody(state)
}

func (c *View) RenderSmallTimedSelection(surface *ui.Surface, rows []ui.Item) {
	ui.List(surface, rows, func(id string) { _ = c.Actions.Select(id) })
}
