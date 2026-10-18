package feedback

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Controller) composeFeedbackListBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}
