package classassessment

import (
	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/checklist"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/features/assessment/internal/panels"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

type LargeTemplate struct {
	Actions           common.Actions
	Pages             Pages
	ClassRoster       func() []ui.Item
	SelectClassPerson func(string)
}

func (c *LargeTemplate) ComposeLargeClassBody(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := panels.Split(surface, .28)
		swimmers := c.ClassRoster()
		c.largePagedGrid(left, "large-class-roster", "Class", swimmers, 1, 5, c.Pages.Class, c.SelectClassPerson)
		c.RenderLargeChecklist(right, rows, func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *LargeTemplate) RenderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	checklist.Controller{}.RenderGrid(surface, rows, selectRow)
}
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key, title string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, title, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Class *int }
type CompactTemplate struct {
	Actions common.Actions
}

func (c *CompactTemplate) ComposeClassChecklist(rows []navigation.Row) ui.BodyComposer {
	return checklist.Controller{Actions: c.Actions}.ComposeCompact(rows)
}
func ConfigureLargeHeader(state *navigation.State, skill curriculum.Skill, course domain.Course) {
	full := skill.Number + ". " + skill.Name
	if ui.WrappedTextHeight(full, 450, 16, true) <= contracts.SplitHeaderHeight {
		state.SplitHeader.Left = full
		return
	}
	state.SplitHeader.Left = skill.Number + ". " + curriculum.CompactTitle(skill)
	state.PrimaryInfo = full + " · " + course.Name + " · " + course.Time
	state.PrimaryInfoFontSize = 16
}

// Compose renders the current checklist in the selected display layout.
func (c *Component) Compose(rows []navigation.Row) ui.BodyComposer {
	if c.Layout.Panels {
		return c.large.ComposeLargeClassBody(rows)
	}
	return c.small.ComposeClassChecklist(rows)
}
