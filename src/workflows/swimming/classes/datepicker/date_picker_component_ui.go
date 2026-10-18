package datepicker

import (
	"fmt"
	"time"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Component) Compose(state shell.State) ui.BodyComposer {
	if state.DisplayMode.IsLarge() {
		return c.composeCalendar(state)
	}
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.List(surface, state.Rows, c.onSelect)
	}
}

func (c *Component) composeCalendar(state shell.State) ui.BodyComposer {
	snapshot := state
	month := c.month
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}

		controls := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, _ := surface.BodyBounds(w, h)
			return x, y, width, 48
		})
		previous := controls.Button("calendar-previous", "‹", func() { c.onMoveMonth(-1) })
		controls.Add(previous, func(w, h float32) (float32, float32, float32, float32) {
			x, y, _, height := controls.BodyBounds(w, h)
			return x, y, 88, height
		})
		next := controls.Button("calendar-next", "›", func() { c.onMoveMonth(1) })
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
		ui.Grid(calendar, items, 7, 7, c.onSelect)
	}
}
