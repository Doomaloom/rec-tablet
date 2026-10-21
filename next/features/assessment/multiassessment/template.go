package multiassessment

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/checklist"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/features/assessment/internal/panels"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

type LargeTemplate struct {
	Actions     common.Actions
	Pages       Pages
	Shared      *ListTemplate
	MultiRows   func(int, int) ([]ui.Item, []ui.Item)
	SelectGroup func(string)
}

func (c *LargeTemplate) ComposeLargeChecklistBody(rows []navigation.Row, selectRow func(string)) ui.BodyComposer {
	return func(surface *ui.Surface) { c.RenderLargeChecklist(surface, rows, selectRow) }
}
func (c *LargeTemplate) ComposeLargeMultiAssessmentBody(students []common.MultiStudentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		pageCount := max(1, (len(students)+5)/6)
		*c.Pages.Table = max(0, min(*c.Pages.Table, pageCount-1))
		if pageCount > 1 {
			for direction, caption := range map[int]string{-1: "‹ Previous swimmers", 1: "Next swimmers ›"} {
				direction, caption := direction, caption
				button := surface.Button(fmt.Sprintf("multi-table-page-%d", direction), caption, func() {
					_ = c.Actions.Run(func() error { *c.Pages.Table = max(0, min(*c.Pages.Table+direction, pageCount-1)); return nil })
				})
				button.Fill, button.Ink = ui.Rail, ui.Blue
				surface.AddNamed(fmt.Sprintf("multi-table-navigation-%d", direction), button, func(w, h float32) (float32, float32, float32, float32) {
					x, y, width, _ := surface.BodyBounds(w, h)
					if direction < 0 {
						return x, y, (width - ui.Gap) / 2, 44
					}
					return x + (width+ui.Gap)/2, y, (width - ui.Gap) / 2, 44
				})
			}
		}
		table := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			if pageCount > 1 {
				return x, y + 48, width, height - 48
			}
			return x, y, width, height
		})
		start := *c.Pages.Table * 6
		c.Shared.ComposeMultiAssessmentBody(students[start:min(start+6, len(students))])(table)
	}
}
func (c *LargeTemplate) ComposeLargeMultiSelectionBody(state navigation.State, group, count int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := panels.Split(surface, .45)
		groupsPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			return x, y, width, (height - ui.Gap) * .47
		})
		membersPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			upper := (height - ui.Gap) * .47
			return x, y + upper + ui.Gap, width, height - upper - ui.Gap
		})
		groups, members := c.MultiRows(group, count)
		c.largePagedGrid(groupsPanel, "multi-groups", groups, 1, 2, c.Pages.Group, c.SelectGroup)
		c.largePagedGrid(membersPanel, "multi-group-members", members, 1, 3, c.Pages.Member, func(id string) { _ = c.Actions.Select(id) })
		c.largePagedGrid(right, "multi-assign-roster", state.Rows, 2, 5, c.Pages.MultiRoster, func(id string) { _ = c.Actions.Select(id) })
	}
}

func (c *LargeTemplate) RenderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	checklist.Controller{}.RenderGrid(surface, rows, selectRow)
}
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Group, Member, MultiRoster, Table *int }

type CompactTemplate struct {
	Actions common.Actions
}

func (c *CompactTemplate) ComposeClassChecklist(rows []navigation.Row) ui.BodyComposer {
	return checklist.Controller{Actions: c.Actions}.ComposeCompact(rows)
}
func multiStudentBounds(count, index int, surface *ui.Surface, w, h float32) (float32, float32, float32, float32) {
	x, y, width, height := surface.BodyBounds(w, h)
	columns := 1
	if count > 8 {
		columns = 2
	}
	rows := (count + columns - 1) / columns
	rowHeight := min(float32(80), (height-float32(rows-1)*ui.Gap)/float32(rows))
	columnWidth := (width - float32(columns-1)*ui.Gap) / float32(columns)
	column, row := index/rows, index%rows
	return x + float32(column)*(columnWidth+ui.Gap), y + float32(row)*(rowHeight+ui.Gap), columnWidth, rowHeight
}

type ListTemplate struct {
	Actions      common.Actions
	AdjustGroups func(int) error
	PageGroup    func(int) error
	PassStudent  func(string) error
	EditStudent  func(string) error
}

func (c *ListTemplate) ComposeMultiAssessmentBody(students []common.MultiStudentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		for index, student := range students {
			index, student := index, student
			nameLabel := student.Name
			statusLabel := "Unmarked"
			switch student.Status {
			case "successful":
				statusLabel = "Passed"
			case "not_yet":
				statusLabel = fmt.Sprintf("Incomplete · %d/%d passed", student.Passed, student.Total)
			}
			name := ui.NewText(nameLabel+" · "+statusLabel, 20, ui.Navy)
			name.Bold, name.Center, name.MinimumFont, name.MaxLines = true, true, 16, -1
			nameBounds := func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := multiStudentBounds(len(students), index, surface, w, h)
				actionWidth := min(float32(64), height)
				if height < 50 {
					name.Text = student.Name
					name.FontSize = 18
				} else {
					name.Text = nameLabel + " · " + statusLabel
				}
				return x, y, width - 2*(actionWidth+ui.Gap), height
			}
			background := canvas.NewRectangle(ui.Rail)
			surface.AddNamed("multi-name-background:"+student.PersonID, background, nameBounds)
			surface.AddNamed("multi-name:"+student.PersonID, name, nameBounds)
			pass := surface.Button("multi-pass:"+student.PersonID, "Pass", func() { _ = c.PassStudent(student.PersonID) })
			fail := surface.Button("multi-fail:"+student.PersonID, "Incomplete: choose passed subskills", func() { _ = c.EditStudent(student.PersonID) })
			pass.Icon, fail.Icon = common.SaveActionIcon, common.DiscardActionIcon
			pass.Selected, fail.Selected = student.Status == "successful", student.Status == "not_yet"
			for action, button := range []*ui.TouchButton{pass, fail} {
				action := action
				button.IconOnly, button.Outlined = true, true
				button.Fill, button.Ink = ui.Rail, ui.Navy
				// Use colored icons while unselected, a white icon on the selected fill.
				if action == 0 {
					if button.Selected {
						button.Fill = ui.Green
					} else {
						button.Icon = multiPassIcon
					}
				} else {
					if button.Selected {
						button.Fill = ui.Red
					} else {
						button.Icon = multiFailIcon
					}
				}
				surface.AddNamed(fmt.Sprintf("multi-result:%s:%d", student.PersonID, action), button, func(w, h float32) (float32, float32, float32, float32) {
					x, y, width, height := multiStudentBounds(len(students), index, surface, w, h)
					actionWidth := min(float32(64), height)
					return x + width - float32(2-action)*actionWidth - float32(1-action)*ui.Gap, y, actionWidth, height
				})
			}
		}
	}
}
func (c *ListTemplate) ComposeMultiGroupCountBody(state contracts.State, count, students int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if students == 0 {
			ui.Empty(surface, state.Empty)
			return
		}
		label := ui.NewText("Number of groups", 24, ui.Navy)
		label.Center, label.Bold = true, true
		surface.AddNamed("multi-count-label", label, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x, y + height*0.04, width, 36
		})
		value := ui.NewText(fmt.Sprint(count), 72, ui.Navy)
		value.Center, value.Bold = true, true
		surface.AddNamed("multi-count-value", value, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x + (width-160)/2, y + height*0.19, 160, 104
		})
		minus := surface.Button("multi-count-minus", "Fewer groups", func() { _ = c.AdjustGroups(-1) })
		plus := surface.Button("multi-count-plus", "More groups", func() { _ = c.AdjustGroups(1) })
		minus.Icon, plus.Icon = multiCountMinusIcon, multiCountPlusIcon
		minus.Disabled, plus.Disabled = count <= 1, count >= students
		for index, button := range []*ui.TouchButton{minus, plus} {
			index := index
			button.IconOnly = true
			if button.Disabled {
				button.Icon = theme.NewDisabledResource(button.Icon)
			}
			surface.AddNamed(fmt.Sprintf("multi-count-adjust-%d", index), button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := surface.BodyBounds(w, h)
				return x + width/2 - 184 + float32(index)*280, y + height*0.19 + 8, 88, 88
			})
		}
		summary := ui.NewText(strings.ReplaceAll(state.Progress, ", ", "\n"), 24, ui.Navy)
		summary.Center, summary.Bold = true, true
		surface.AddNamed("multi-count-summary", summary, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x, y + height*0.52, width, height * 0.25
		})
		next := surface.Button("submit", "Continue", func() { _ = c.Actions.Submit() })
		surface.AddNamed("multi-count-continue", next, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x + (width-240)/2, y + height - 56, 240, 56
		})
	}
}
func (c *ListTemplate) ComposeMultiSelectionBody(state contracts.State, group, count int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		previous := surface.Button("multi-group-previous", "‹ Previous group", func() { _ = c.PageGroup(-1) })
		next := surface.Button("multi-group-next", "Next group ›", func() { _ = c.PageGroup(1) })
		previous.Disabled, next.Disabled = group == 0, group == count-1
		for index, button := range []*ui.TouchButton{previous, next} {
			index := index
			button.FontSize, button.Fill, button.Ink = 18, ui.Rail, ui.Blue
			surface.AddNamed(fmt.Sprintf("multi-group-navigation-%d", index), button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, _ := surface.BodyBounds(w, h)
				buttonWidth := (width - ui.Gap) / 2
				return x + float32(index)*(buttonWidth+ui.Gap), y, buttonWidth, ui.HeaderHeight
			})
		}
		listSurface := *surface
		listSurface.Bounds = func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			offset := ui.HeaderHeight + ui.Gap
			return x, y + offset, width, height - offset
		}
		ui.Grid(&listSurface, state.Rows, 1, max(1, state.PageSize), func(id string) { _ = c.Actions.Select(id) })
	}
}

var (
	multiCountMinusIcon = fyne.NewStaticResource("group-count-minus.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 32h44" fill="none" stroke="#fff" stroke-width="8"/></svg>`))
	multiCountPlusIcon  = fyne.NewStaticResource("group-count-plus.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 32h44M32 10v44" fill="none" stroke="#fff" stroke-width="8"/></svg>`))
	multiPassIcon       = fyne.NewStaticResource("multi-pass.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 34l14 14 30-32" fill="none" stroke="#17675d" stroke-width="8" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
	multiFailIcon       = fyne.NewStaticResource("multi-fail.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M14 14l36 36M50 14L14 50" fill="none" stroke="#a43633" stroke-width="8" stroke-linecap="round"/></svg>`))
)
