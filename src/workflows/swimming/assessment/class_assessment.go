package assessment

import (
	"fmt"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// ClassAssessment owns its mode state and presentation. Data and actions arrive as callbacks.
type ClassAssessment struct {
	attempts, roster []domain.MultiAttempt
	status           map[string]string
	rosterPage       int
	data             classAssessmentData
	actions          classAssessmentActions
	small            *small.View
	large            *large.View
}

type classAssessmentData struct {
	Draft func(string) (map[string]string, error)
	Skill func(string, string) (curriculum.Skill, error)
}

type classAssessmentActions struct {
	classRailActions func() []shell.RailAction
}

func (c *ClassAssessment) build(input modeInput) (navigation.State, error) {
	state, course, f := input.state, input.course, input.frame
	if len(c.attempts) == 0 {
		return state, fmt.Errorf("no pending class attempts")
	}
	skill, err := c.data.Skill(f.Version, f.Skill)
	if err != nil {
		return state, err
	}
	state.PageCount = len(c.attempts)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	student := c.attempts[state.Page]
	answers, err := c.data.Draft(student.AttemptID)
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
	state.LeftRailActions = c.actions.classRailActions()
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.large.ComposeLargeClassBody(state.Rows)
	} else {
		state.ComposeBody = c.small.ComposeClassChecklist(state.Rows)
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

func (c *ClassAssessment) rosterRows(page int) []ui.Item {
	var rows []ui.Item
	current := c.attempts[page].PersonID
	for _, student := range c.roster {
		status := c.status[student.PersonID]
		if status == "" {
			status = "Pending"
		}
		row := ui.Item{ID: student.PersonID, Label: student.Name + " · " + status, Selected: student.PersonID == current, Disabled: status != "Pending"}
		if row.Selected {
			row.Fill = ui.Green
		}
		rows = append(rows, row)
	}
	return rows
}
