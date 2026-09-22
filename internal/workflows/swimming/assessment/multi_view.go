package assessment

import (
	"fmt"
	"strings"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (c *Controller) composeMultiGroupCountBody(state shell.State, count, students int) ui.BodyComposer {
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
		minus := surface.Button("multi-count-minus", "Fewer groups", func() { _ = c.adjustMultiGroupCount(-1) })
		plus := surface.Button("multi-count-plus", "More groups", func() { _ = c.adjustMultiGroupCount(1) })
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
		summary.Center = true
		surface.AddNamed("multi-count-summary", summary, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x, y + height*0.52, width, height * 0.25
		})
		next := surface.Button("submit", "Continue", func() { _ = c.Controller.Submit() })
		surface.AddNamed("multi-count-continue", next, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			return x + (width-240)/2, y + height - 56, 240, 56
		})
	}
}

func (c *Controller) composeMultiSelectionBody(state shell.State, group, count int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		previous := surface.Button("multi-group-previous", "‹ Previous group", func() { _ = c.pageMultiGroup(-1) })
		next := surface.Button("multi-group-next", "Next group ›", func() { _ = c.pageMultiGroup(1) })
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
		ui.List(&listSurface, state.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
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

func (c *Controller) composeMultiAssessmentBody(students []multiStudentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		for index, student := range students {
			index, student := index, student
			nameLabel := student.Name
			if len(students) > 8 {
				nameLabel = student.ShortName
			}
			name := surface.Button("multi-name:"+student.PersonID, nameLabel, func() {
				label := widget.NewLabel(student.Name)
				label.Wrapping = fyne.TextWrapWord
				details := dialog.NewCustom("Student", "Close", container.NewVScroll(label), surface.Window)
				details.Resize(fyne.NewSize(520, 200))
				details.Show()
			})
			name.Fill, name.Ink, name.FontSize = ui.Rail, ui.Navy, 22
			name.Compact, name.MaxLabelLines = true, 1
			name.Detail = "Unmarked"
			switch student.Status {
			case "successful":
				name.Detail = "Passed"
			case "not_yet":
				name.Detail = fmt.Sprintf("Incomplete · %d/%d passed", student.Passed, student.Total)
			}
			statusDetail := name.Detail
			surface.AddNamed("multi-name:"+student.PersonID, name, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := multiStudentBounds(len(students), index, surface, w, h)
				actionWidth := min(float32(64), height)
				// Compact rows omit the second line; the selected result button still
				// distinguishes assessed students from unmarked students.
				name.Detail = statusDetail
				if height < 48 {
					name.Detail = ""
				}
				return x, y, width - 2*(actionWidth+ui.Gap), height
			})
			pass := surface.Button("multi-pass:"+student.PersonID, "Pass", func() { _ = c.PassMultiStudent(student.PersonID) })
			fail := surface.Button("multi-fail:"+student.PersonID, "Incomplete: choose passed subskills", func() { _ = c.EditMultiStudent(student.PersonID) })
			pass.Icon, fail.Icon = saveActionIcon, discardActionIcon
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

var (
	multiCountMinusIcon = fyne.NewStaticResource("group-count-minus.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 32h44" fill="none" stroke="#fff" stroke-width="8"/></svg>`))
	multiCountPlusIcon  = fyne.NewStaticResource("group-count-plus.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 32h44M32 10v44" fill="none" stroke="#fff" stroke-width="8"/></svg>`))
	multiPassIcon       = fyne.NewStaticResource("multi-pass.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 34l14 14 30-32" fill="none" stroke="#17675d" stroke-width="8" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
	multiFailIcon       = fyne.NewStaticResource("multi-fail.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M14 14l36 36M50 14L14 50" fill="none" stroke="#a43633" stroke-width="8" stroke-linecap="round"/></svg>`))
)
