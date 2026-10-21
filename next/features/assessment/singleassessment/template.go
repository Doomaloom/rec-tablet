package singleassessment

import (
	"image/color"

	"fyne.io/fyne/v2/container"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/checklist"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/features/assessment/internal/panels"
	"rectablet/next/navigation"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
)

type LargeTemplate struct {
	Actions      common.Actions
	Pages        Pages
	SwitchPerson func(string)
}

func (c *LargeTemplate) ComposeLargeChecklistBody(rows []navigation.Row, selectRow func(string)) ui.BodyComposer {
	return func(surface *ui.Surface) { c.RenderLargeChecklist(surface, rows, selectRow) }
}
func (c *LargeTemplate) ComposeLargeSingleBody(rows []navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := panels.Split(surface, .28)
		c.largePagedGrid(left, "large-single-roster", swimmers, 1, 6, c.Pages.Roster, c.SwitchPerson)
		c.RenderLargeChecklist(right, rows, func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *LargeTemplate) RenderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	checklist.Controller{}.RenderGrid(surface, rows, selectRow)
}
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Roster *int }

func renderCriterion(surface *ui.Surface, row ui.Item, mark func(string)) {

	box := card.New(container.NewWithoutLayout(), card.Panel)
	surface.AddNamed("assessment-card", box, surface.BodyBounds)
	criterion := ui.NewText(row.Label, 25, ui.Navy)
	criterion.Bold = true
	criterion.MinimumFont = 18
	surface.Criterion = criterion
	surface.AddNamed("assessment-criterion", criterion, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		height -= 2 * ui.ButtonPadding
		if mark != nil {
			height -= ui.HeaderHeight + ui.ButtonPadding
		}
		return x + ui.ButtonPadding, y + ui.ButtonPadding, width - 2*ui.ButtonPadding, height
	})
	if mark != nil {
		yes := surface.Button("successful", "Successful", func() { mark("successful") })
		yes.Fill = color.NRGBA{R: 229, G: 241, B: 238, A: 255}
		yes.Ink = ui.Green
		if row.Status == "successful" {
			yes.Fill = ui.Green
			yes.Ink = ui.White
			yes.Selected = true
		}
		surface.AddNamed("assessment-successful", yes, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x + ui.ButtonPadding, y + height - ui.ButtonPadding - ui.HeaderHeight, (width - 3*ui.ButtonPadding) / 2, ui.HeaderHeight
		})
		no := surface.Button("not_yet", "Not yet", func() { mark("not_yet") })
		no.Fill = color.NRGBA{R: 248, G: 237, B: 223, A: 255}
		no.Ink = ui.Amber
		if row.Status == "not_yet" {
			no.Fill = ui.Amber
			no.Ink = ui.White
			no.Selected = true
		}
		surface.AddNamed("assessment-not-yet", no, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			buttonWidth := (width - 3*ui.ButtonPadding) / 2
			return x + 2*ui.ButtonPadding + buttonWidth, y + height - ui.ButtonPadding - ui.HeaderHeight, buttonWidth, ui.HeaderHeight
		})
	}
}

type CompactTemplate struct {
	Actions common.Actions
}

func (c *CompactTemplate) ComposeAssessmentBody(state contracts.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		renderCriterion(surface, snapshot.Rows[0], func(status string) { _ = c.Actions.Mark(status) })
	}
}
func (c *CompactTemplate) ComposeSubskillChecklistBody(state contracts.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		rows := append([]ui.Item(nil), snapshot.Rows...)
		for i := range rows {
			if rows[i].Status == "successful" {
				rows[i].Fill = ui.Green
				rows[i].Selected = true
			}
		}
		ui.Grid(surface, rows, 1, 3, func(id string) { _ = c.Actions.Select(id) })
	}
}

type ListTemplate struct {
	Actions common.Actions
}

func (c *ListTemplate) ComposeAssessmentListBody(state contracts.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.Grid(surface, snapshot.Rows, 1, max(1, snapshot.PageSize), func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *ListTemplate) ComposeAttemptListBody(state contracts.State) ui.BodyComposer {
	return c.ComposeAssessmentListBody(state)
}
func ConfigureLargeHeader(state *navigation.State, skill curriculum.Skill) {
	if ui.WrappedTextHeight(skill.Name, 450, 16, true) <= contracts.SplitHeaderHeight {
		return
	}
	state.SplitHeader.Right = curriculum.CompactTitle(skill)
	state.PrimaryInfo = skill.Number + ". " + skill.Name
	state.PrimaryInfoFontSize = 16
}
func (c *Component) compose(state *contracts.State, f navigation.Frame) {
	if state.HeaderAction != "" {
		state.HeaderWidth = 134
	}
	switch f.Screen {
	case navigation.Attempts:
		state.ComposeBody = c.shared.ComposeAttemptListBody(*state)
	case navigation.Assessment:
		if state.FooterAction == "" {
			state.ComposeBody = c.small.ComposeSubskillChecklistBody(*state)
		} else {
			state.ComposeBody = c.small.ComposeAssessmentBody(*state)
		}
	}
}
