package classassessment

import (
	"rectablet/next/features/assessment/internal/checklist"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/features/assessment/internal/panels"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
)

type LargeTemplate struct {
	Actions           common.Actions
	Pages             Pages
	ClassRoster       func() []ui.Item
	SelectClassPerson func(string)
}

const studentColumnFraction float32 = .28

func (c *LargeTemplate) ComposeLargeClassBody(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		main := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			reclaimed := max(float32(0), ui.ContentInset-x)
			return x + reclaimed, y, width - reclaimed, height
		})
		left, right := panels.Split(main, studentColumnFraction)
		roster := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			start, _, _, _ := surface.BodyBounds(w, h)
			return start, ui.OuterMargin, width + x - start, height + y - ui.OuterMargin
		})
		roster = ui.SelectableSurface(roster)
		swimmers := c.ClassRoster()
		c.largePagedGrid(roster, "large-class-roster", swimmers, 1, 12, c.Pages.Class, c.SelectClassPerson)
		c.RenderLargeChecklist(right, rows, func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *LargeTemplate) RenderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	checklist.Controller{Columns: 1, Fill: ui.Blue, Ink: ui.White}.RenderGrid(surface, rows, selectRow)
}
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Class *int }
type CompactTemplate struct {
	Actions common.Actions
}

func (c *CompactTemplate) ComposeClassChecklist(rows []navigation.Row) ui.BodyComposer {
	return checklist.Controller{Actions: c.Actions}.ComposeCompact(rows)
}
func ConfigureLargeHeader(state *navigation.State, skill curriculum.Skill) {
	state.SplitHeader = nil
	state.Title = skill.Number + ". " + skill.Name
	state.PrimaryInfo = ""
	state.HeaderStartFraction = studentColumnFraction
	state.HeaderTitleFontSize = 22
	state.FitHeader = true
}

// Compose renders the current checklist in the selected display layout.
func (c *Component) Compose(rows []navigation.Row) ui.BodyComposer {
	if c.Layout.Panels {
		return c.large.ComposeLargeClassBody(rows)
	}
	return c.small.ComposeClassChecklist(rows)
}
