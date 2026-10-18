package roster

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Component) Compose(state shell.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		if state.DisplayMode.IsLarge() {
			ui.Grid(surface, state.Rows, 2, 6, c.onSelect)
			return
		}
		ui.List(surface, state.Rows, c.onSelect)
	}
}
