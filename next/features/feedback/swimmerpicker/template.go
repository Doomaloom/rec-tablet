package swimmerpicker

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func (c *SwimmerPicker) Compose(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.Grid(ui.SelectableSurface(surface), state.Rows, 1, max(1, state.PageSize), c.OnSelect)
	}
}

// RenderSidebar receives the panel owned by the parent's combined layout.
func (c *SwimmerPicker) RenderSidebar(surface *ui.Surface, state contracts.State) {
	sidebar := *surface
	sidebar.Button = func(id, label string, run func()) *ui.TouchButton {
		button := surface.Button(id, label, run)
		button.Selectable = true
		button.FontSize = 18
		button.MinimumFont = 16
		return button
	}
	ui.Grid(&sidebar, state.Rows, 1, 12, c.OnSelect)
}
