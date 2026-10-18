package assessment

import (
	"fmt"

	"fyne.io/fyne/v2/theme"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/navigation"
)

// MultiAssessment owns its mode state and presentation. Data and actions arrive as callbacks.
type MultiAssessment struct {
	session                                      *multiSession
	edit                                         *multiStudentEdit
	rosterPage, groupPage, memberPage, tablePage int
	data                                         multiAssessmentData
	actions                                      multiAssessmentActions
	paginate                                     func(*navigation.State, []navigation.Row)
	small                                        *small.View
	large                                        *large.View
	shared                                       *shared.View
}

type multiAssessmentData struct {
	Draft func(string) (map[string]string, error)
}

type multiAssessmentActions struct {
	DiscardMultiStudent func() error
	EnterSingleMode     func() error
	SaveMultiStudent    func() error
	changeMultiCount    func() error
	multiRailActions    func() []shell.RailAction
	skillModeSwitch     func(curriculum.Skill, bool) []shell.HeaderSwitchOption
}

func (c *MultiAssessment) buildGroupCount(input modeInput) (navigation.State, error) {
	state := input.state
	if c.session == nil {
		return state, fmt.Errorf("group setup is no longer available")
	}
	state.Title = "Number of Groups"
	state.PrimaryInfo = fmt.Sprintf("%s · %s · %d students present", c.session.CourseInfo, c.session.Skill.Name, len(c.session.Roster))
	state.HeaderSwitch = c.actions.skillModeSwitch(c.session.Skill, true)
	state.Empty = "No students marked present. Take attendance first."
	state.BackRailAction = &shell.RailAction{Label: "Back", Run: c.actions.EnterSingleMode, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.HideFooter = true
	state.CanSubmit = len(c.session.Roster) > 0
	state.Progress = multiSplitLabel(len(c.session.Roster), c.session.ProposedGroupCount)
	state.ProgressInBody = true
	state.ComposeBody = c.shared.ComposeMultiGroupCountBody(state, c.session.ProposedGroupCount, len(c.session.Roster))
	return state, nil
}

func (c *MultiAssessment) buildSelection(input modeInput) (navigation.State, error) {
	state := input.state
	if c.session == nil || len(c.session.Groups) == 0 {
		return state, fmt.Errorf("choose the number of groups first")
	}
	group := c.session.Groups[c.session.AssignmentGroup]
	state.Title = fmt.Sprintf("Assign Group %d of %d", group.Number, c.session.GroupCount)
	selectionInfo := fmt.Sprintf("%d of %d selected · %s", len(group.Members), group.Capacity, c.session.Skill.Name)
	state.PrimaryInfo = c.session.CourseInfo + " · " + selectionInfo
	state.HeaderActions = []shell.HeaderAction{{Label: "Groups", Run: c.actions.changeMultiCount}}
	state.HomeRailAction = discardRailAction(c.actions.EnterSingleMode)
	state.HomeRailAction.Label = "Cancel setup"
	state.BackRailAction = &shell.RailAction{Label: "Group count", Run: c.actions.changeMultiCount, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.FooterAction, state.CanSubmit = "Start marking", c.multiReady()
	assigned := 0
	for _, g := range c.session.Groups {
		assigned += len(g.Members)
	}
	state.Progress = fmt.Sprintf("%d/%d assigned", assigned, len(c.session.Roster))
	var rows []navigation.Row
	for _, student := range c.session.Roster {
		row := navigation.Row{ID: student.ID, Label: student.Name + " · Unassigned"}
		for i, other := range c.session.Groups {
			if containsString(other.Members, student.ID) {
				row.Label = fmt.Sprintf("%s · Group %d", student.Name, other.Number)
				row.Selected = i == c.session.AssignmentGroup
				row.Disabled = !row.Selected
				if row.Selected {
					row.Fill = ui.Green
				}
			}
		}
		if !row.Selected && len(group.Members) == group.Capacity {
			row.Disabled = true
		}
		rows = append(rows, row)
	}
	if state.DisplayMode.IsLarge() {
		state.PageSize = max(1, len(rows))
	}
	c.paginate(&state, rows)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.large.ComposeLargeMultiSelectionBody(state, c.session.AssignmentGroup, len(c.session.Groups))
	} else {
		state.ComposeBody = c.shared.ComposeMultiSelectionBody(state, c.session.AssignmentGroup, len(c.session.Groups))
	}
	return state, nil
}

func (c *MultiAssessment) buildAssessment(input modeInput) (navigation.State, error) {
	state, f := input.state, input.frame
	if c.session == nil || len(c.session.Groups) == 0 {
		return state, fmt.Errorf("no pending groups")
	}
	state.PageCount = len(c.session.Groups)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	group, skill := c.session.Groups[state.Page], c.session.Skill
	label := fmt.Sprintf("Group %d of %d", group.Number, c.session.GroupCount)
	state.Title = label
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: label}
	state.PrimaryInfo = fmt.Sprintf("%s · %d students", c.session.CourseInfo, len(group.Attempts))
	state.HideFooter = true
	state.LeftRailActions = c.actions.multiRailActions()
	students := make([]multiStudentSnapshot, 0, len(group.Attempts))
	for _, attempt := range group.Attempts {
		answers, err := c.data.Draft(attempt.AttemptID)
		if err != nil {
			return state, err
		}
		student := multiStudentSnapshot{PersonID: attempt.PersonID, Name: attempt.Name, Status: "unmarked", Total: len(skill.Subskills)}
		for _, subskill := range skill.Subskills {
			if answers[subskill.ID] == "successful" {
				student.Passed++
			}
		}
		if len(answers) > 0 {
			student.Status = "not_yet"
		}
		if student.Total > 0 && student.Passed == student.Total {
			student.Status = "successful"
		}
		students = append(students, student)
		state.Rows = append(state.Rows, navigation.Row{ID: student.PersonID, Label: student.Name, Status: student.Status})
	}
	state.RowCount, state.PageSize = len(students), len(students)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.large.ComposeLargeMultiAssessmentBody(students)
	} else {
		state.ComposeBody = c.shared.ComposeMultiAssessmentBody(students)
	}
	return state, nil
}

func (c *MultiAssessment) buildStudent(input modeInput) (navigation.State, error) {
	state, f := input.state, input.frame
	if c.session == nil || c.edit == nil || c.edit.Attempt.AttemptID != f.Attempt {
		return state, fmt.Errorf("no student selected")
	}
	skill, edit := c.session.Skill, c.edit
	state.Title = edit.Attempt.Name
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: edit.Attempt.Name}
	state.PrimaryInfo = c.session.CourseInfo + " · " + skill.Number + ". " + skill.Name
	state.HideFooter = true
	state.HomeRailAction = saveRailAction(c.actions.SaveMultiStudent)
	state.BackRailAction = discardRailAction(c.actions.DiscardMultiStudent)
	// Checklist edits stay local until Save; leaving via Escape also abandons them.
	state.OnLeave = func() { c.edit = nil }
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: edit.Answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.large.ComposeLargeChecklistBody(state.Rows, func(id string) { _ = c.large.Actions.Select(id) })
	} else {
		state.ComposeBody = c.small.ComposeClassChecklist(state.Rows)
	}
	return state, nil
}

func (c *MultiAssessment) multiReady() bool {
	if c.session == nil || len(c.session.Groups) == 0 {
		return false
	}
	eligible := make(map[string]bool, len(c.session.Roster))
	for _, student := range c.session.Roster {
		eligible[student.ID] = true
	}
	seen := make(map[string]bool, len(eligible))
	for _, group := range c.session.Groups {
		if len(group.Members) != group.Capacity {
			return false
		}
		for _, person := range group.Members {
			if !eligible[person] || seen[person] {
				return false
			}
			seen[person] = true
		}
	}
	return len(seen) == len(eligible)
}

func (c *MultiAssessment) toggleMultiMember(person string) error {
	if c.session == nil || len(c.session.Groups) == 0 {
		return nil
	}
	eligible := false
	for _, student := range c.session.Roster {
		if student.ID == person {
			eligible = true
			break
		}
	}
	if !eligible {
		return nil
	}
	group := &c.session.Groups[c.session.AssignmentGroup]
	if containsString(group.Members, person) {
		group.Members = removeString(group.Members, person)
		return nil
	}
	if len(group.Members) >= group.Capacity {
		return nil
	}
	for _, other := range c.session.Groups {
		if containsString(other.Members, person) {
			return nil
		}
	}
	group.Members = append(group.Members, person)
	return nil
}

func (c *MultiAssessment) toggleMultiSubskill(id string) error {
	if c.edit == nil {
		return nil
	}
	for _, subskill := range c.session.Skill.Subskills {
		if subskill.ID == id {
			if c.edit.Answers[id] == "successful" {
				delete(c.edit.Answers, id)
			} else {
				c.edit.Answers[id] = "successful"
			}
			break
		}
	}
	return nil
}

func (c *MultiAssessment) largeMultiRows(group, count int) ([]ui.Item, []ui.Item) {
	var groups, members []ui.Item
	for i := 0; i < count; i++ {
		g := c.session.Groups[i]
		row := ui.Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("Group %d · %d/%d", g.Number, len(g.Members), g.Capacity), Selected: i == group}
		if row.Selected {
			row.Fill = ui.Green
		}
		groups = append(groups, row)
	}
	for _, student := range c.session.Roster {
		if containsString(c.session.Groups[group].Members, student.ID) {
			members = append(members, ui.Item{ID: student.ID, Label: student.Name, Selected: true, Fill: ui.Green})
		}
	}
	return groups, members
}
