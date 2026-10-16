package classes

import (
	"fmt"
	"time"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"
)

func (c *Controller) composeFeatureBody(state *shell.State) {
	if c.Frame().Screen == navigation.Classes {
		state.HoldTitle = func() { _ = c.Maintenance() }
	}
	switch c.Frame().Screen {
	case navigation.Classes:
		state.ComposeBody = c.composeClassListBody(*state)
	case navigation.Dates:
		state.ComposeBody = c.composeDateListBody(*state)
	case navigation.Class:
		state.ComposeBody = c.composeLessonMenuBody(*state)
	}
}

func (c *Controller) composeClassListBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeGridBody(state, 2, 3)
	}
	return c.composeRowListBody(state)
}

func (c *Controller) composeDateListBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeDateCalendarBody(state)
	}
	return c.composeRowListBody(state)
}

func (c *Controller) composeLessonMenuBody(state shell.State) ui.BodyComposer {
	if c.DisplayMode.IsLarge() {
		return c.composeGridBody(state, 3, 1)
	}
	return c.composeRowListBody(state)
}

func (c *Controller) composeGridBody(state shell.State, columns, rows int) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.Grid(surface, snapshot.Rows, columns, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeRowListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeDateCalendarBody(state shell.State) ui.BodyComposer {
	snapshot := state
	month := c.calendarMonth
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		if month.IsZero() {
			month, _ = time.Parse("2006-01-02", c.Date)
		}
		month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())

		controls := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x, y, width, 48
		})
		previous := controls.Button("calendar-previous", "‹", func() { _ = c.moveCalendarMonth(-1) })
		controls.Add(previous, func(w, h float32) (float32, float32, float32, float32) {
			x, y, _, height := controls.BodyBounds(w, h)
			return x, y, 88, height
		})
		next := controls.Button("calendar-next", "›", func() { _ = c.moveCalendarMonth(1) })
		controls.Add(next, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := controls.BodyBounds(w, h)
			return x + width - 88, y, 88, height
		})
		label := ui.NewText(month.Format("January 2006"), 24, ui.Navy)
		label.Bold, label.Center = true, true
		controls.Add(label, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := controls.BodyBounds(w, h)
			return x + 92, y, width - 184, height
		})

		calendar := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x, y + 52, width, height - 52
		})
		weekday := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		items := make([]ui.Item, 0, 49)
		for i, day := range weekday {
			items = append(items, ui.Item{ID: fmt.Sprintf("weekday-%d", i), Label: day, Disabled: true, Selected: true})
		}
		scheduled := map[string]ui.Item{}
		for _, row := range snapshot.Rows {
			scheduled[row.ID] = row
		}
		start := -int(month.Weekday())
		daysInMonth := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location()).Day()
		for offset := 0; offset < 42; offset++ {
			dayNumber := start + offset + 1
			if dayNumber < 1 || dayNumber > daysInMonth {
				items = append(items, ui.Item{ID: fmt.Sprintf("empty-%d", offset), Disabled: true})
				continue
			}
			date := time.Date(month.Year(), month.Month(), dayNumber, 0, 0, 0, 0, month.Location()).Format("2006-01-02")
			if row, ok := scheduled[date]; ok {
				row.Label = fmt.Sprintf("%d", dayNumber)
				items = append(items, row)
			} else {
				items = append(items, ui.Item{ID: fmt.Sprintf("unscheduled-%s", date), Label: fmt.Sprintf("%d", dayNumber), Disabled: true})
			}
		}
		ui.Grid(calendar, items, 7, 7, func(id string) { _ = c.Controller.Select(id) })
	}
}
