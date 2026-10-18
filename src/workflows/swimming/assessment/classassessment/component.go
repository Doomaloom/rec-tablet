package classassessment

import (
	"fmt"

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

// Component owns class assessment state and presentation through explicit data and action callbacks.
type Component struct {
	navigation common.Navigation
	State      State
	storage    Storage
	actions    Actions
	paginate   func(*navigation.State, []navigation.Row)
	small      *small.View
	large      *large.View
}

// State holds mode session and paging state.
type State struct {
	Attempts, Roster []domain.MultiAttempt
	Status           map[string]string
	RosterPage       int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	DiscardAttempts   func([]string) error
	Draft             func(string) (map[string]string, error)
	ListRoster        func(string) ([]domain.Enrollment, error)
	Mark              func(string, string, string) error
	OpenMultiAttempts func([]string, string, string) ([]domain.MultiAttempt, error)
	Skill             func(string, string) (curriculum.Skill, error)
	SubmitAttempts    func([]string) error
}

type Actions struct {
	BeforeEnter func()
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
	c.large = &large.View{Actions: options.ViewActions,
		Pages: large.Pages{Class: &c.State.RosterPage}, ClassRoster: func() []ui.Item { return c.rosterRows(c.navigation.Frame().Page) }, SelectClassPerson: c.SelectPerson,
	}
	return c
}

func (c *Component) Build(input common.Input) (navigation.State, error) {
	state, course, f := input.State, input.Course, input.Frame
	if len(c.State.Attempts) == 0 {
		return state, fmt.Errorf("no pending class attempts")
	}
	skill, err := c.storage.Skill(f.Version, f.Skill)
	if err != nil {
		return state, err
	}
	state.PageCount = len(c.State.Attempts)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	student := c.State.Attempts[state.Page]
	answers, err := c.storage.Draft(student.AttemptID)
	if err != nil {
		return state, err
	}
	state.Title = student.Name
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: student.Name}
	state.PrimaryInfo = fmt.Sprintf("%s · %s", course.Name, course.Time)
	if state.DisplayMode.IsLarge() {
		ConfigureLargeHeader(&state, skill, course)
	} else {
		navigation.AppendPrimaryInfo(&state, fmt.Sprintf("%s. %s", skill.Number, skill.Name))
		lineHeight := ui.WrappedTextHeight("Ag", 680, 20, true)
		if ui.WrappedTextHeight(state.PrimaryInfo, 680, 20, true) > 2*lineHeight+0.01 {
			state.PrimaryInfoFontSize = 14
		}
	}
	state.HideFooter = true
	state.LeftRailActions = c.railActions()
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	state.ComposeBody = c.Compose(state.Rows, state.DisplayMode)
	return state, nil
}

func ConfigureLargeHeader(state *navigation.State, skill curriculum.Skill, course domain.Course) {
	full := skill.Number + ". " + skill.Name
	if ui.WrappedTextHeight(full, 450, 16, true) <= shell.SplitHeaderHeight {
		state.SplitHeader.Left = full
		return
	}
	state.SplitHeader.Left = skill.Number + ". " + curriculum.CompactTitle(skill)
	state.PrimaryInfo = full + " · " + course.Name + " · " + course.Time
	state.PrimaryInfoFontSize = 16
}

func (c *Component) rosterRows(page int) []ui.Item {
	var rows []ui.Item
	current := c.State.Attempts[page].PersonID
	for _, student := range c.State.Roster {
		status := c.State.Status[student.PersonID]
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

// Compose renders the current checklist in the selected display layout.
func (c *Component) Compose(rows []navigation.Row, mode ui.DisplayMode) ui.BodyComposer {
	if mode.IsLarge() {
		return c.large.ComposeLargeClassBody(rows)
	}
	return c.small.ComposeClassChecklist(rows)
}
