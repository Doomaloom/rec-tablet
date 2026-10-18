// Package datepicker owns lesson-date presentation and calendar browsing state.
package datepicker

import (
	"rectablet/src/shell"
	"time"
)

type Options struct {
	OnSelect    func(string)
	OnMoveMonth func(int)
}

type Component struct {
	month       time.Time
	onSelect    func(string)
	onMoveMonth func(int)
}

func New(options Options) *Component {
	return &Component{onSelect: options.OnSelect, onMoveMonth: options.OnMoveMonth}
}

// Configure prepares date choices and initializes the large-display calendar.
func (c *Component) Configure(state *shell.State, dates []string, selected string) []shell.Row {
	state.Title = "Lesson Date"
	state.Empty = "Import a class package to add lesson dates."
	rows := make([]shell.Row, 0, len(dates))
	for _, date := range dates {
		parsed, _ := time.Parse("2006-01-02", date)
		rows = append(rows, shell.Row{ID: date, Label: parsed.Format("Monday, Jan 02, 2006"), Selected: date == selected})
	}
	if state.DisplayMode.IsLarge() {
		c.ensureMonth(selected)
		// The calendar needs all scheduled dates, rather than a page of dates.
		state.PageSize = max(1, len(rows))
	}
	return rows
}

func (c *Component) Month() time.Time { return c.month }
func (c *Component) Reset()           { c.month = time.Time{} }
func (c *Component) MoveMonth(delta int, selected string) {
	c.ensureMonth(selected)
	c.month = c.month.AddDate(0, delta, 0)
}

func (c *Component) ensureMonth(selected string) {
	if !c.month.IsZero() {
		return
	}
	date, err := time.Parse("2006-01-02", selected)
	if err != nil {
		date = time.Now()
	}
	c.month = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}
