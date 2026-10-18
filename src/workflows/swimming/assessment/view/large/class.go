package large

import (
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *View) ComposeLargeClassBody(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := Split(surface, .28)
		swimmers := c.ClassRoster()
		c.largePagedGrid(left, "large-class-roster", "Class", swimmers, 1, 5, c.Pages.Class, c.SelectClassPerson)
		c.RenderLargeChecklist(right, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}
