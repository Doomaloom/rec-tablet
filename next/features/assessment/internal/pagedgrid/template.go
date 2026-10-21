package pagedgrid

import (
	"fmt"

	"rectablet/src/shared/ui"
)

func (c Controller) Render(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	capacity := columns * rowCount
	pageCount := max(1, (len(rows)+capacity-1)/capacity)
	*page = max(0, min(*page, pageCount-1))
	if pageCount > 1 {
		for direction, caption := range map[int]string{-1: "‹", 1: "›"} {
			button := surface.Button(fmt.Sprintf("%s-page-%d", key, direction), caption, func() {
				_ = c.Run(func() error { *page = max(0, min(*page+direction, pageCount-1)); return nil })
			})
			button.Fill, button.Ink, button.FontSize = ui.Rail, ui.Blue, 26
			surface.AddNamed(fmt.Sprintf("%s-page-button-%d", key, direction), button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := surface.BodyBounds(w, h)
				if direction < 0 {
					return x, y + height - 44, 64, 44
				}
				return x + width - 64, y + height - 44, 64, 44
			})
		}
	}
	grid := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		bottom := float32(0)
		if pageCount > 1 {
			bottom = 48
		}
		return x, y + ui.Gap, width, height - ui.Gap - bottom
	})
	start := *page * capacity
	visible := rows[start:min(start+capacity, len(rows))]
	if len(visible) == 0 {
		ui.Empty(grid, "No swimmers available.")
		return
	}
	buttonFactory := grid.Button
	grid.Button = func(id, label string, selectRow func()) *ui.TouchButton {
		button := buttonFactory(id, label, selectRow)
		button.FontSize, button.MinimumFont = 18, 16
		button.Compact = true
		return button
	}
	ui.Grid(grid, visible, columns, rowCount, selectRow)
}
