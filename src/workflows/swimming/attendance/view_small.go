package attendance

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Controller) renderSmallAttendanceRoster(surface *ui.Surface, state shell.State) {
	ui.List(surface, state.Rows, func(id string) { _ = c.Controller.Select(id) })
}
