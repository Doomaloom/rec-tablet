package timedheat

import (
	"strconv"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/features/assessment/internal/panels"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

type LargeTemplate struct {
	Actions          common.Actions
	Pages            Pages
	LapsInput        func() string
	SelectionCount   func() int
	AddDigit         func(string)
	ClearLaps        func()
	StartHeat        func() error
	SelectedSwimmers func() []ui.Item
}

func (c *LargeTemplate) ComposeLargeTimedHeatBody(state navigation.State) ui.BodyComposer {
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

// ComposeLargeTimedLapsBody supports drafts opened in compact mode.
func (c *LargeTemplate) ComposeLargeTimedLapsBody(state navigation.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := panels.Split(surface, .5)
		swimmers := c.SelectedSwimmers()
		c.largePagedGrid(left, "timed-setup-swimmers", swimmers, 1, 6, c.Pages.Roster, func(string) {})
		c.largeLapKeypad(right)
	}
}
func (c *LargeTemplate) ComposeLargeTimedSetup(state navigation.State, rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := panels.Split(surface, .5)
		c.largePagedGrid(left, "timed-setup-swimmers", rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
		c.largeLapKeypad(right)
	}
}
func (c *LargeTemplate) RenderLargeTimedSelection(surface *ui.Surface, rows []ui.Item) {
	c.largePagedGrid(surface, "timed-selection", rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
}
func (c *LargeTemplate) largeLapKeypad(surface *ui.Surface) {
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
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Roster, Heat *int }
type CompactTemplate struct {
	Actions common.Actions
}

func (c *CompactTemplate) ComposeTimedHeatBody(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		ui.Grid(surface, state.Rows, 1, max(1, state.PageSize), func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *CompactTemplate) ComposeTimedLapsBody(state contracts.State) ui.BodyComposer {
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
func (c *CompactTemplate) RenderSmallTimedSelection(surface *ui.Surface, rows []ui.Item) {
	ui.Grid(surface, rows, 1, 3, func(id string) { _ = c.Actions.Select(id) })
}
func (c *Component) compose(state *contracts.State, f navigation.Frame) {
	switch f.Screen {
	case navigation.SkillSwimmers:
		state.ComposeBody = c.composeTimedSelectionBody(*state)
	case navigation.TimedLaps:
		if c.Layout.Panels {
			state.ComposeBody = c.large.ComposeLargeTimedLapsBody(*state)
		} else {
			state.ComposeBody = c.small.ComposeTimedLapsBody(*state)
		}
	case navigation.TimedHeat:
		if c.Layout.Panels {
			state.ComposeBody = c.large.ComposeLargeTimedHeatBody(*state)
		} else {
			state.ComposeBody = c.small.ComposeTimedHeatBody(*state)
		}
	}
}
func (c *Component) composeTimedSelectionBody(state contracts.State) ui.BodyComposer {
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
		if c.Layout.Panels {
			c.large.RenderLargeTimedSelection(surface, rows)
		} else {
			c.small.RenderSmallTimedSelection(surface, rows)
		}
	}
}
func (c *Component) configureLargeTimedSetup(state *navigation.State, rows []navigation.Row) {
	state.Title = "Set Up Timed Heat"
	state.PageSize = max(1, len(rows))
	state.HomeRailAction = nil
	state.BackRailAction = common.DiscardRailAction(c.DiscardTimedSelection)
	state.Progress = c.Input()
	state.ProgressInBody = true
	state.ComposeBody = c.large.ComposeLargeTimedSetup(*state, rows)
}
