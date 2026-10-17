package assessment

import (
	"fmt"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/navigation"
)

func configureLargeAssessmentHeader(state *navigation.State, skill curriculum.Skill) {
	if ui.WrappedTextHeight(skill.Name, 450, 16, true) <= shell.SplitHeaderHeight {
		return
	}
	state.SplitHeader.Right = curriculum.CompactTitle(skill)
	state.PrimaryInfo = skill.Number + ". " + skill.Name
	state.PrimaryInfoFontSize = 16
}

// largeSplit keeps feature panels inside the shell's body rectangle.
func largeSplit(surface *ui.Surface, leftFraction float32) (*ui.Surface, *ui.Surface) {
	left := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, y, (width - ui.Gap) * leftFraction, height
	})
	right := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		leftWidth := (width - ui.Gap) * leftFraction
		return x + leftWidth + ui.Gap, y, width - leftWidth - ui.Gap, height
	})
	return left, right
}

func (c *Controller) largePagedGrid(surface *ui.Surface, key, title string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	capacity := columns * rowCount
	pageCount := max(1, (len(rows)+capacity-1)/capacity)
	*page = max(0, min(*page, pageCount-1))
	label := ui.NewText(fmt.Sprintf("%s · %d/%d", title, *page+1, pageCount), 19, ui.Navy)
	label.Bold, label.MaxLines, label.MinimumFont = true, -1, 16
	titleHeight := float32(34)
	if len(title) > 32 {
		titleHeight = 62
	}
	surface.AddNamed(key+"-title", label, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, _ := surface.BodyBounds(w, h)
		return x + ui.ButtonPadding, y, width - 2*ui.ButtonPadding, titleHeight
	})
	if pageCount > 1 {
		for direction, caption := range map[int]string{-1: "‹", 1: "›"} {
			direction, caption := direction, caption
			button := surface.Button(fmt.Sprintf("%s-page-%d", key, direction), caption, func() {
				_ = c.Run(func() error { *page = max(0, min(*page+direction, pageCount-1)); return nil })
			})
			button.Fill, button.Ink, button.FontSize = ui.Rail, ui.Blue, 26
			surface.AddNamed(fmt.Sprintf("%s-page-button-%d", key, direction), button, func(w, h float32) (float32, float32, float32, float32) {
				x, y, width, height := surface.BodyBounds(w, h)
				if direction < 0 {
					return x, y + height - 44, 64, 44
				}
				return x + width - 64, y + height - 44, 64, 44
			})
		}
	}
	grid := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		bottom := float32(0)
		if pageCount > 1 {
			bottom = 48
		}
		return x, y + titleHeight + ui.Gap, width, height - titleHeight - ui.Gap - bottom
	})
	start := *page * capacity
	visible := rows[start:min(start+capacity, len(rows))]
	if len(visible) == 0 {
		ui.Empty(grid, "No swimmers available.")
		return
	}
	buttonFactory := grid.Button
	grid.Button = func(id, label string, selectRow func()) *ui.TouchButton {
		button := buttonFactory(id, label, selectRow)
		button.FontSize, button.MinimumFont = 18, 16
		button.Compact = true
		return button
	}
	ui.Grid(grid, visible, columns, rowCount, selectRow)
}

func (c *Controller) composeLargeSkillsBody(skills []navigation.Row, selected navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .46)
		if len(skills) == 0 {
			ui.Empty(surface, "No skills available.")
			return
		}
		skillRows := append([]ui.Item(nil), skills...)
		for i := range skillRows {
			if skillRows[i].ID == selected.ID {
				skillRows[i].Selected = true
			}
		}
		c.largePagedGrid(ui.SelectableSurface(left), "large-skills", "Skills", skillRows, 1, 6, &c.largeSkillPage, func(id string) {
			if _, timed := benchmarkFor(id); timed {
				_ = c.openLargeTimedSkill(id)
				return
			}
			_ = c.Run(func() error { c.largeSkillID = id; c.largeRosterPage = 0; return nil })
		})
		c.largePagedGrid(right, "large-skill-swimmers", selected.Label, swimmers, 2, 6, &c.largeRosterPage, func(id string) {
			_ = c.openLargeSkillSwimmer(selected, id)
		})
	}
}

func (c *Controller) openLargeSkillSwimmer(skill navigation.Row, person string) error {
	params := navigation.Destination(skill)
	if _, timed := benchmarkFor(skill.ID); timed {
		if err := c.openLargeTimedSkill(skill.ID); err != nil {
			return err
		}
		return c.Controller.Select(person)
	}
	return c.Run(func() error {
		progress, err := c.Store.Progress(person, skill.ID, params.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.Store.NewAttempt(person, c.Frame().Lesson, skill.ID)
		if err != nil {
			return err
		}
		c.Push(navigation.Assessment, func(f *navigation.Frame) {
			f.Person, f.Attempt, f.Skill, f.Level, f.Version = person, attempt, skill.ID, params.Level, params.Version
		})
		return nil
	})
}

func (c *Controller) openLargeTimedSkill(id string) error {
	for _, row := range c.State.Rows {
		if row.ID != id {
			continue
		}
		params := navigation.Destination(row)
		return c.Run(func() error {
			c.largeSkillID = id
			c.timedSelection = nil
			c.timedLapsInput = ""
			c.Push(navigation.SkillSwimmers, func(f *navigation.Frame) { f.Skill, f.Level, f.Version = id, params.Level, params.Version })
			return nil
		})
	}
	return nil
}

func (c *Controller) openLargeSelectedMode(mode string) error {
	if c.Frame().Screen != navigation.Skills {
		return nil
	}
	if mode == "single" {
		return nil
	}
	var chosen navigation.Row
	for _, row := range c.State.Rows {
		if row.ID == c.largeSkillID {
			chosen = row
			break
		}
	}
	if chosen.ID == "" {
		return nil
	}
	params := navigation.Destination(chosen)
	if err := c.Run(func() error {
		c.Push(navigation.SkillSwimmers, func(f *navigation.Frame) {
			f.Skill, f.Level, f.Version = chosen.ID, params.Level, params.Version
		})
		return nil
	}); err != nil {
		return err
	}
	switch mode {
	case "multi":
		return c.EnterMultiMode()
	case "class":
		return c.EnterClassMode()
	}
	return nil
}

func (c *Controller) composeLargeSingleBody(rows []navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .28)
		c.largePagedGrid(left, "large-single-roster", "Swimmers", swimmers, 1, 6, &c.largeRosterPage, c.switchLargeAssessmentPerson)
		c.renderLargeChecklist(right, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) switchLargeAssessmentPerson(person string) {
	if person == c.Frame().Person {
		return
	}
	_ = c.Run(func() error {
		f := &c.Frames[len(c.Frames)-1]
		progress, err := c.Store.Progress(person, f.Skill, f.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.Store.NewAttempt(person, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		f.Person, f.Attempt, f.Page = person, attempt, 0
		return nil
	})
}

func (c *Controller) toggleLargeCorrectionSubskill(id string) error {
	if c.Frame().Screen != navigation.Assessment || !c.Frame().Review {
		return nil
	}
	status := "successful"
	for _, row := range c.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "not_yet"
			break
		}
	}
	return c.Run(func() error { return c.Store.Mark(c.Frame().Attempt, id, status) })
}

func (c *Controller) composeLargeChecklistBody(rows []navigation.Row, selectRow func(string)) ui.BodyComposer {
	return func(surface *ui.Surface) { c.renderLargeChecklist(surface, rows, selectRow) }
}

// Every checklist criterion is visible at once. Twelve criteria use two columns
// of six, leaving at least a 48 pixel target in the large shell body.
func (c *Controller) renderLargeChecklist(surface *ui.Surface, rows []navigation.Row, selectRow func(string)) {
	if len(rows) == 0 {
		ui.Empty(surface, "No criteria available.")
		return
	}
	columns := 1
	if len(rows) > 6 {
		columns = 2
	}
	visualRows := (len(rows) + columns - 1) / columns
	var lastWidth, lastHeight, sharedFont float32
	for i, row := range rows {
		i, row := i, row
		button := surface.Button("criterion:"+row.ID, row.Label, func() {
			if selectRow != nil {
				selectRow(row.ID)
			}
		})
		button.Compact, button.MaxLabelLines = true, -1
		button.Fill = ui.Rail
		button.Ink = ui.Navy
		if row.Status == "successful" {
			button.Fill, button.Ink, button.Selected = ui.Green, ui.White, true
		}
		if row.Status == "not_yet" {
			button.Fill, button.Ink = ui.Amber, ui.White
		}
		surface.AddNamed("large-criterion-"+row.ID, button, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			cellWidth := (width - float32(columns-1)*ui.Gap) / float32(columns)
			cellHeight := (height - float32(visualRows-1)*ui.Gap) / float32(visualRows)
			if width != lastWidth || height != lastHeight {
				sharedFont = fitLargeChecklistFont(rows, cellWidth, cellHeight)
				lastWidth, lastHeight = width, height
			}
			button.FontSize = sharedFont
			return x + float32(i%columns)*(cellWidth+ui.Gap), y + float32(i/columns)*(cellHeight+ui.Gap), cellWidth, cellHeight
		})
	}
}

func fitLargeChecklistFont(rows []navigation.Row, cellWidth, cellHeight float32) float32 {
	for font := float32(20); font >= 14; font-- {
		fits := true
		for _, row := range rows {
			if ui.WrappedTextHeight(row.Label, cellWidth-2*ui.ButtonPadding, font, true)+4 > cellHeight {
				fits = false
				break
			}
		}
		if fits {
			return font
		}
	}
	return 14
}

func (c *Controller) composeLargeClassBody(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .28)
		var swimmers []ui.Item
		current := c.classAttempts[c.Frame().Page].PersonID
		for _, student := range c.classRoster {
			status := c.classStatus[student.PersonID]
			if status == "" {
				status = "Pending"
			}
			row := ui.Item{ID: student.PersonID, Label: student.Name + " · " + status, Selected: student.PersonID == current, Disabled: status != "Pending"}
			if row.Selected {
				row.Fill = ui.Green
			}
			swimmers = append(swimmers, row)
		}
		c.largePagedGrid(left, "large-class-roster", "Class", swimmers, 1, 5, &c.largeClassPage, func(id string) {
			_ = c.Run(func() error {
				for i, student := range c.classAttempts {
					if student.PersonID == id {
						c.Frames[len(c.Frames)-1].Page = i
						break
					}
				}
				return nil
			})
		})
		c.renderLargeChecklist(right, rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeLargeMultiSelectionBody(state navigation.State, group, count int) ui.BodyComposer {
	return func(surface *ui.Surface) {
		left, right := largeSplit(surface, .45)
		groupsPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			return x, y, width, (height - ui.Gap) * .47
		})
		membersPanel := ui.Panel(left, func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := left.BodyBounds(w, h)
			upper := (height - ui.Gap) * .47
			return x, y + upper + ui.Gap, width, height - upper - ui.Gap
		})
		var groups []ui.Item
		for i := 0; i < count; i++ {
			g := c.multi.Groups[i]
			row := ui.Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("Group %d · %d/%d", g.Number, len(g.Members), g.Capacity), Selected: i == group}
			if row.Selected {
				row.Fill = ui.Green
			}
			groups = append(groups, row)
		}
		c.largePagedGrid(groupsPanel, "multi-groups", "Groups", groups, 1, 2, &c.largeGroupPage, func(id string) {
			var index int
			if _, err := fmt.Sscan(id, &index); err != nil {
				return
			}
			_ = c.Run(func() error { c.multi.AssignmentGroup = index; c.largeMemberPage = 0; return nil })
		})
		var members []ui.Item
		for _, student := range c.multi.Roster {
			if containsString(c.multi.Groups[group].Members, student.ID) {
				members = append(members, ui.Item{ID: student.ID, Label: student.Name, Selected: true, Fill: ui.Green})
			}
		}
		c.largePagedGrid(membersPanel, "multi-group-members", "Group members", members, 1, 3, &c.largeMemberPage, func(id string) { _ = c.Controller.Select(id) })
		c.largePagedGrid(right, "multi-assign-roster", "Assign swimmers", state.Rows, 2, 5, &c.largeMultiRosterPage, func(id string) { _ = c.Controller.Select(id) })
	}
}

func (c *Controller) composeLargeMultiAssessmentBody(students []multiStudentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		pageCount := max(1, (len(students)+5)/6)
		c.largeTablePage = max(0, min(c.largeTablePage, pageCount-1))
		if pageCount > 1 {
			for direction, caption := range map[int]string{-1: "‹ Previous swimmers", 1: "Next swimmers ›"} {
				direction, caption := direction, caption
				button := surface.Button(fmt.Sprintf("multi-table-page-%d", direction), caption, func() {
					_ = c.Run(func() error { c.largeTablePage = max(0, min(c.largeTablePage+direction, pageCount-1)); return nil })
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
		start := c.largeTablePage * 6
		c.composeMultiAssessmentBody(students[start:min(start+6, len(students))])(table)
	}
}
