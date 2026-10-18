package attendance

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
)

func (c *Controller) renderLargeAttendanceRoster(surface *ui.Surface, state shell.State) {
	ui.Grid(surface, state.Rows, 2, 6, func(id string) { _ = c.Controller.Select(id) })
}
