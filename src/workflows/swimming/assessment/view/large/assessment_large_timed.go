package large

import (
	"strconv"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) RenderLargeTimedSelection(surface *ui.Surface, rows []ui.Item) {
	c.largePagedGrid(surface, "timed-selection", rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
}

func (c *View) ComposeLargeTimedSetup(state navigation.State, rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .5)
		c.largePagedGrid(left, "timed-setup-swimmers", rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
		c.largeLapKeypad(right)
	}
}

func (c *View) largeLapKeypad(surface *ui.Surface) {
	display := ui.NewText("Laps needed: "+c.LapsInput(), 24, ui.Navy)
	display.Bold, display.Center = true, true
	surface.AddNamed("timed-laps-display", display, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, _ := surface.BodyBounds(w, h)
		return x, y, width, 48
	})
	for index, digit := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		index, digit := index, digit
		button := surface.Button("timed-digit:"+digit, digit, func() {
			c.AddDigit(digit)
		})
		button.FontSize = 28
		surface.AddNamed("timed-digit-"+digit, button, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			cellWidth := (width - 2*ui.Gap) / 3
			cellHeight := (height - 2*52 - 3*ui.Gap) / 4
			return x + float32(index%3)*(cellWidth+ui.Gap), y + 52 + float32(index/3)*(cellHeight+ui.Gap), cellWidth, cellHeight
		})
	}
	clear := surface.Button("timed-clear", "Clear", c.ClearLaps)
	clear.Fill, clear.Ink = ui.Rail, ui.Blue
	surface.AddNamed("timed-clear", clear, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		cellWidth := (width - 2*ui.Gap) / 3
		cellHeight := (height - 2*52 - 3*ui.Gap) / 4
		return x + cellWidth + ui.Gap, y + 52 + 3*(cellHeight+ui.Gap), 2*cellWidth + ui.Gap, cellHeight
	})
	start := surface.Button("timed-create-heat", "Continue to heat", func() { _ = c.StartHeat() })
	laps, _ := strconv.Atoi(c.LapsInput())
	start.Disabled = c.SelectionCount() == 0 || laps < 1 || laps > 999
	start.Fill = ui.Green
	surface.AddNamed("timed-create-heat", start, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, y + height - 48, width, 48
	})
}

// ComposeLargeTimedLapsBody supports drafts opened in compact mode.
func (c *View) ComposeLargeTimedLapsBody(state navigation.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .5)
		swimmers := c.SelectedSwimmers()
		c.largePagedGrid(left, "timed-setup-swimmers", swimmers, 1, 6, c.Pages.Roster, func(string) {})
		c.largeLapKeypad(right)
	}
}

func (c *View) ComposeLargeTimedHeatBody(state navigation.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		rows := append([]ui.Item(nil), state.Rows...)
		for i := range rows {
			if rows[i].Status == "successful" {
				rows[i].Fill, rows[i].Selected = ui.Green, true
			}
		}
		c.largePagedGrid(surface, "timed-heat-swimmers", rows, 2, 6, c.Pages.Heat, func(id string) { _ = c.Actions.Select(id) })
	}
}
