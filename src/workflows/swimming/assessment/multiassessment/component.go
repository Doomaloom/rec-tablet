package multiassessment

import (
	"fmt"

	"fyne.io/fyne/v2/theme"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Component owns multi-group assessment state and presentation through explicit data and action callbacks.
type Component struct {
	navigation common.Navigation
	State      State
	storage    Storage
	actions    Actions
	paginate   func(*navigation.State, []navigation.Row)
	small      *small.View
	large      *large.View
	shared     *shared.View
}

// State holds mode session and paging state.
type State struct {
	Session                                      *Session
	Edit                                         *StudentEdit
	RosterPage, GroupPage, MemberPage, TablePage int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	Attendance           func(string) (map[string]string, error)
	DiscardMultiAttempts func([]domain.MultiAttempt) error
	Draft                func(string) (map[string]string, error)
	GetCourse            func(string) (domain.Course, error)
	ListRoster           func(string) ([]domain.Enrollment, error)
	OpenMultiAttempts    func([]string, string, string) ([]domain.MultiAttempt, error)
	ReplaceDraftAnswers  func(string, map[string]string) error
	Skill                func(string, string) (curriculum.Skill, error)
	SubmitAttempts       func([]string) error
}

type Actions struct {
	EnterSingleMode func() error
	ModeSwitch      func(curriculum.Skill, bool) []shell.HeaderSwitchOption
}

// Options provides storage and navigation operations without a parent-component dependency.
type Options struct {
	Navigation  common.Navigation
	Storage     Storage
	Actions     Actions
	ViewActions shared.Actions
	Paginate    func(*navigation.State, []navigation.Row)
}

func New(options Options) *Component {
	c := &Component{navigation: options.Navigation, storage: options.Storage, actions: options.Actions, paginate: options.Paginate}
	c.small = &small.View{Actions: options.ViewActions}
	c.shared = &shared.View{Actions: options.ViewActions, AdjustGroups: c.AdjustGroupCount, PageGroup: c.PageGroup, PassStudent: c.PassMultiStudent, EditStudent: c.EditMultiStudent}
	c.large = &large.View{Actions: options.ViewActions,
		Shared: c.shared, Pages: large.Pages{Group: &c.State.GroupPage, Member: &c.State.MemberPage, MultiRoster: &c.State.RosterPage, Table: &c.State.TablePage}, MultiRows: c.largeMultiRows, SelectGroup: c.SelectGroup,
	}
	return c
}

func (c *Component) BuildGroupCount(input common.Input) (navigation.State, error) {
	state := input.State
	if c.State.Session == nil {
		return state, fmt.Errorf("group setup is no longer available")
	}
	state.Title = "Number of Groups"
	state.PrimaryInfo = fmt.Sprintf("%s · %s · %d students present", c.State.Session.CourseInfo, c.State.Session.Skill.Name, len(c.State.Session.Roster))
	state.HeaderSwitch = c.actions.ModeSwitch(c.State.Session.Skill, true)
	state.Empty = "No students marked present. Take attendance first."
	state.BackRailAction = &shell.RailAction{Label: "Back", Run: c.actions.EnterSingleMode, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.HideFooter = true
	state.CanSubmit = len(c.State.Session.Roster) > 0
	state.Progress = SplitLabel(len(c.State.Session.Roster), c.State.Session.ProposedGroupCount)
	state.ProgressInBody = true
	state.ComposeBody = c.shared.ComposeMultiGroupCountBody(state, c.State.Session.ProposedGroupCount, len(c.State.Session.Roster))
	return state, nil
}

func (c *Component) BuildSelection(input common.Input) (navigation.State, error) {
	state := input.State
	if c.State.Session == nil || len(c.State.Session.Groups) == 0 {
		return state, fmt.Errorf("choose the number of groups first")
	}
	group := c.State.Session.Groups[c.State.Session.AssignmentGroup]
	state.Title = fmt.Sprintf("Assign Group %d of %d", group.Number, c.State.Session.GroupCount)
	selectionInfo := fmt.Sprintf("%d of %d selected · %s", len(group.Members), group.Capacity, c.State.Session.Skill.Name)
	state.PrimaryInfo = c.State.Session.CourseInfo + " · " + selectionInfo
	state.HeaderActions = []shell.HeaderAction{{Label: "Groups", Run: c.ChangeGroupCount}}
	state.HomeRailAction = common.DiscardRailAction(c.actions.EnterSingleMode)
	state.HomeRailAction.Label = "Cancel setup"
	state.BackRailAction = &shell.RailAction{Label: "Group count", Run: c.ChangeGroupCount, Icon: theme.NavigateBackIcon(), Fill: ui.Rail, Ink: ui.Blue}
	state.FooterAction, state.CanSubmit = "Start marking", c.Ready()
	assigned := 0
	for _, g := range c.State.Session.Groups {
		assigned += len(g.Members)
	}
	state.Progress = fmt.Sprintf("%d/%d assigned", assigned, len(c.State.Session.Roster))
	var rows []navigation.Row
	for _, student := range c.State.Session.Roster {
		row := navigation.Row{ID: student.ID, Label: student.Name + " · Unassigned"}
		for i, other := range c.State.Session.Groups {
			if common.Contains(other.Members, student.ID) {
				row.Label = fmt.Sprintf("%s · Group %d", student.Name, other.Number)
				row.Selected = i == c.State.Session.AssignmentGroup
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
		state.ComposeBody = c.large.ComposeLargeMultiSelectionBody(state, c.State.Session.AssignmentGroup, len(c.State.Session.Groups))
	} else {
		state.ComposeBody = c.shared.ComposeMultiSelectionBody(state, c.State.Session.AssignmentGroup, len(c.State.Session.Groups))
	}
	return state, nil
}

func (c *Component) BuildAssessment(input common.Input) (navigation.State, error) {
	state, f := input.State, input.Frame
	if c.State.Session == nil || len(c.State.Session.Groups) == 0 {
		return state, fmt.Errorf("no pending groups")
	}
	state.PageCount = len(c.State.Session.Groups)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	group, skill := c.State.Session.Groups[state.Page], c.State.Session.Skill
	label := fmt.Sprintf("Group %d of %d", group.Number, c.State.Session.GroupCount)
	state.Title = label
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: label}
	state.PrimaryInfo = fmt.Sprintf("%s · %d students", c.State.Session.CourseInfo, len(group.Attempts))
	state.HideFooter = true
	state.LeftRailActions = c.railActions()
	students := make([]studentSnapshot, 0, len(group.Attempts))
	for _, attempt := range group.Attempts {
		answers, err := c.storage.Draft(attempt.AttemptID)
		if err != nil {
			return state, err
		}
		student := studentSnapshot{PersonID: attempt.PersonID, Name: attempt.Name, Status: "unmarked", Total: len(skill.Subskills)}
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

func (c *Component) BuildStudent(input common.Input) (navigation.State, error) {
	state, f := input.State, input.Frame
	if c.State.Session == nil || c.State.Edit == nil || c.State.Edit.Attempt.AttemptID != f.Attempt {
		return state, fmt.Errorf("no student selected")
	}
	skill, edit := c.State.Session.Skill, c.State.Edit
	state.Title = edit.Attempt.Name
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: edit.Attempt.Name}
	state.PrimaryInfo = c.State.Session.CourseInfo + " · " + skill.Number + ". " + skill.Name
	state.HideFooter = true
	state.HomeRailAction = common.SaveRailAction(c.SaveMultiStudent)
	state.BackRailAction = common.DiscardRailAction(c.DiscardMultiStudent)
	// Checklist edits stay local until Save; leaving via Escape also abandons them.
	state.OnLeave = func() { c.State.Edit = nil }
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

func (c *Component) Ready() bool {
	if c.State.Session == nil || len(c.State.Session.Groups) == 0 {
		return false
	}
	eligible := make(map[string]bool, len(c.State.Session.Roster))
	for _, student := range c.State.Session.Roster {
		eligible[student.ID] = true
	}
	seen := make(map[string]bool, len(eligible))
	for _, group := range c.State.Session.Groups {
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

func (c *Component) ToggleMember(person string) error {
	if c.State.Session == nil || len(c.State.Session.Groups) == 0 {
		return nil
	}
	eligible := false
	for _, student := range c.State.Session.Roster {
		if student.ID == person {
			eligible = true
			break
		}
	}
	if !eligible {
		return nil
	}
	group := &c.State.Session.Groups[c.State.Session.AssignmentGroup]
	if common.Contains(group.Members, person) {
		group.Members = common.Remove(group.Members, person)
		return nil
	}
	if len(group.Members) >= group.Capacity {
		return nil
	}
	for _, other := range c.State.Session.Groups {
		if common.Contains(other.Members, person) {
			return nil
		}
	}
	group.Members = append(group.Members, person)
	return nil
}

func (c *Component) ToggleSubskill(id string) error {
	if c.State.Edit == nil {
		return nil
	}
	for _, subskill := range c.State.Session.Skill.Subskills {
		if subskill.ID == id {
			if c.State.Edit.Answers[id] == "successful" {
				delete(c.State.Edit.Answers, id)
			} else {
				c.State.Edit.Answers[id] = "successful"
			}
			break
		}
	}
	return nil
}

func (c *Component) largeMultiRows(group, count int) ([]ui.Item, []ui.Item) {
	var groups, members []ui.Item
	for i := 0; i < count; i++ {
		g := c.State.Session.Groups[i]
		row := ui.Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("Group %d · %d/%d", g.Number, len(g.Members), g.Capacity), Selected: i == group}
		if row.Selected {
			row.Fill = ui.Green
		}
		groups = append(groups, row)
	}
	for _, student := range c.State.Session.Roster {
		if common.Contains(c.State.Session.Groups[group].Members, student.ID) {
			members = append(members, ui.Item{ID: student.ID, Label: student.Name, Selected: true, Fill: ui.Green})
		}
	}
	return groups, members
}
