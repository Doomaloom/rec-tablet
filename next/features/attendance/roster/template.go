package roster

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func (c *Component) Compose(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if state.RowCount == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		ui.Grid(surface, state.Rows, max(1, c.Layout.Columns), max(1, state.PageSize/max(1, c.Layout.Columns)), c.onSelect)
	}
}
