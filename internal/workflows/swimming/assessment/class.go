package assessment

import (
	"fmt"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
)

// EnterClassMode opens one pending attempt for every active student at this level.
// Existing drafts are reused, just as they are when entering Multi mode.
func (c *Controller) EnterClassMode() error {
	if c.Frame().Screen != navigation.SkillSwimmers && c.Frame().Screen != navigation.MultiSelection && c.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		roster, err := c.Store.ListRoster(f.Course)
		if err != nil {
			return err
		}
		var people []string
		for _, student := range roster {
			if student.Active && student.Level == f.Level {
				people = append(people, student.ID)
			}
		}
		if len(people) == 0 {
			return nil
		}
		attempts, err := c.Store.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		c.leaveMultiSetup()
		c.multi, c.multiEdit = nil, nil
		c.classAttempts = attempts
		c.classRoster = append([]domain.MultiAttempt(nil), attempts...)
		c.classStatus = make(map[string]string, len(attempts))
		c.Push(navigation.ClassAssessment, nil)
		return nil
	})
}

func (c *Controller) buildClassAssessmentState(course domain.Course, f navigation.Frame) (navigation.State, error) {
	state := c.InitialState()
	if len(c.classAttempts) == 0 {
		return state, fmt.Errorf("no pending class attempts")
	}
	skill, err := c.Store.Skill(f.Version, f.Skill)
	if err != nil {
		return state, err
	}
	state.PageCount = len(c.classAttempts)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	c.Frames[len(c.Frames)-1].Page = state.Page
	student := c.classAttempts[state.Page]
	answers, err := c.Store.Draft(student.AttemptID)
	if err != nil {
		return state, err
	}
	state.Title = student.Name
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: student.Name}
	state.PrimaryInfo = fmt.Sprintf("%s · %s", course.Name, course.Time)
	if state.DisplayMode.IsLarge() {
		configureLargeClassHeader(&state, skill, course)
	} else {
		navigation.AppendPrimaryInfo(&state, fmt.Sprintf("%s. %s", skill.Number, skill.Name))
		lineHeight := ui.WrappedTextHeight("Ag", 680, 20, true)
		if ui.WrappedTextHeight(state.PrimaryInfo, 680, 20, true) > 2*lineHeight+0.01 {
			state.PrimaryInfoFontSize = 14
		}
	}
	state.HideFooter = true
	state.LeftRailActions = c.classRailActions()
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.composeLargeClassBody(state.Rows)
	} else {
		state.ComposeBody = c.composeClassChecklist(state.Rows)
	}
	return state, nil
}

func configureLargeClassHeader(state *navigation.State, skill curriculum.Skill, course domain.Course) {
	full := skill.Number + ". " + skill.Name
	if ui.WrappedTextHeight(full, 450, 16, true) <= shell.SplitHeaderHeight {
		state.SplitHeader.Left = full
		return
	}
	state.SplitHeader.Left = skill.Number + ". " + curriculum.CompactTitle(skill)
	state.PrimaryInfo = full + " · " + course.Name + " · " + course.Time
	state.PrimaryInfoFontSize = 16
}

func (c *Controller) classRailActions() []shell.RailAction {
	save, discard := *saveRailAction(c.SaveClassStudent), *discardRailAction(c.DiscardClassStudent)
	save.ID, save.Label = "save-student", "Save current student"
	discard.ID, discard.Label = "discard-student", "Discard current attempt"
	discardAll, saveAll := *discardRailAction(c.DiscardClass), *saveRailAction(c.SaveClass)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending attempts", discardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending attempts", saveAllActionIcon
	return []shell.RailAction{save, discard, discardAll, saveAll}
}

func (c *Controller) toggleClassSubskill(id string) error {
	index := c.Frame().Page
	if index < 0 || index >= len(c.classAttempts) {
		return nil
	}
	status := "successful"
	for _, row := range c.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
		}
	}
	if err := c.Store.Mark(c.classAttempts[index].AttemptID, id, status); err != nil {
		return err
	}
	c.Saved("")
	return nil
}

func (c *Controller) SaveClassStudent() error    { return c.finishClassAttempts(false, true) }
func (c *Controller) DiscardClassStudent() error { return c.finishClassAttempts(false, false) }
func (c *Controller) SaveClass() error           { return c.finishClassAttempts(true, true) }
func (c *Controller) DiscardClass() error        { return c.finishClassAttempts(true, false) }

// Only remove students from the queue after the complete storage operation succeeds.
func (c *Controller) finishClassAttempts(all, save bool) error {
	if c.Frame().Screen != navigation.ClassAssessment || len(c.classAttempts) == 0 {
		return nil
	}
	return c.Run(func() error {
		index := c.Frame().Page
		pending := c.classAttempts
		if !all {
			pending = pending[index : index+1]
		}
		ids := make([]string, len(pending))
		for i, student := range pending {
			ids[i] = student.AttemptID
		}
		var err error
		if save {
			err = c.Store.SubmitAttempts(ids)
		} else {
			err = c.Store.DiscardAttempts(ids)
		}
		if err != nil {
			return err
		}
		for _, student := range pending {
			if save {
				c.classStatus[student.PersonID] = "Saved"
			} else {
				c.classStatus[student.PersonID] = "Discarded"
			}
		}
		if all {
			c.classAttempts = nil
		} else {
			c.classAttempts = append(c.classAttempts[:index], c.classAttempts[index+1:]...)
		}
		if len(c.classAttempts) == 0 {
			c.Pop()
		} else {
			c.Frames[len(c.Frames)-1].Page = min(index, len(c.classAttempts)-1)
		}
		c.Saved("")
		return nil
	})
}

type checklistLayout struct {
	height float32
	font   float32
}

const (
	classChecklistGap    float32 = 2
	classChecklistBottom float32 = ui.OuterMargin
)

// Divide the page into equal rows and use the largest shared font that fits every label.
func fitClassChecklist(rows []navigation.Row, width, height float32) checklistLayout {
	if len(rows) == 0 {
		return checklistLayout{}
	}
	available := height - float32(len(rows)-1)*classChecklistGap
	layout := checklistLayout{height: available / float32(len(rows)), font: 22}
	for _, row := range rows {
		// Allow compact button padding plus one pixel for fractional-scale rounding.
		for layout.font > 1 && ui.WrappedTextHeight(row.Label, width-2*ui.ButtonPadding, layout.font, true)+5 > layout.height {
			layout.font--
		}
	}
	return layout
}

func (c *Controller) composeClassChecklist(rows []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		var layout checklistLayout
		var lastWidth, lastHeight float32
		for i, row := range rows {
			index, row := i, row
			button := surface.Button("row:"+row.ID, row.Label, func() { _ = c.Controller.Select(row.ID) })
			button.Compact, button.MaxLabelLines = true, -1
			if index%2 == 1 {
				button.Fill = ui.Navy
			}
			if row.Status == "successful" {
				button.Fill = ui.Green
			}
			surface.AddNamed(fmt.Sprintf("class-subskill-%d", index), button, func(width, height float32) (float32, float32, float32, float32) {
				x, top, bodyWidth, bodyHeight := surface.BodyBounds(width, height)
				if width != lastWidth || height != lastHeight {
					layout = fitClassChecklist(rows, bodyWidth, bodyHeight)
					lastWidth, lastHeight = width, height
				}
				button.FontSize = layout.font
				y := top + float32(index)*(layout.height+classChecklistGap)
				return x, y, bodyWidth, layout.height
			})
		}
	}
}
