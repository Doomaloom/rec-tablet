package classassessment

import (
	"fmt"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// Component owns class assessment state and presentation through explicit data and action callbacks.
type Component struct {
	Layout     Layout
	navigation common.Navigation
	State      State
	storage    Storage
	actions    Actions
	paginate   func(*navigation.State, []navigation.Row)
	small      *CompactTemplate
	large      *LargeTemplate
}

// State holds mode session and paging state.
type State struct {
	Attempts, Roster []domain.MultiAttempt
	Status           map[string]string
	Successes        map[string]int
	RosterPage       int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	DiscardAttempts   func([]string) error
	Draft             func(string) (map[string]string, error)
	ListRoster        func(string) ([]domain.Enrollment, error)
	Mark              func(string, string, string) error
	OpenMultiAttempts func([]string, string, string) ([]domain.MultiAttempt, error)
	Progress          func(string, string, string) (domain.Progress, error)
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
	ViewActions common.Actions
	Paginate    func(*navigation.State, []navigation.Row)
}

func New(options Options) *Component {
	c := &Component{navigation: options.Navigation, storage: options.Storage, actions: options.Actions, paginate: options.Paginate}
	c.small = &CompactTemplate{Actions: options.ViewActions}
	c.large = &LargeTemplate{Actions: options.ViewActions,
		Pages: Pages{Class: &c.State.RosterPage}, ClassRoster: func() []ui.Item { return c.rosterRows(c.navigation.Frame().Page) }, SelectClassPerson: c.SelectPerson,
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
	state.SplitHeader = &contracts.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: student.Name}
	state.PrimaryInfo = fmt.Sprintf("%s · %s", course.Name, course.Time)
	if c.Layout.Panels {
		successes := make(map[string]int, len(c.State.Roster))
		for _, person := range c.State.Roster {
			progress, err := c.storage.Progress(person.PersonID, f.Skill, f.Version)
			if err != nil {
				return state, err
			}
			successes[person.PersonID] = min(3, max(0, progress.Successes))
		}
		c.State.Successes = successes
		ConfigureLargeHeader(&state, skill)
	} else {
		navigation.AppendPrimaryInfo(&state, fmt.Sprintf("%s. %s", skill.Number, skill.Name))
		lineHeight := ui.WrappedTextHeight("Ag", 680, 20, true)
		if ui.WrappedTextHeight(state.PrimaryInfo, 680, 20, true) > 2*lineHeight+0.01 {
			state.PrimaryInfoFontSize = 14
		}
	}
	state.HideFooter = true
	if c.Layout.Panels {
		state.HideLeftRail = true
		state.RightRailActions = c.classRailActions()
	} else {
		state.LeftRailActions = c.railActions()
	}
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	state.ComposeBody = c.Compose(state.Rows)
	return state, nil
}

func (c *Component) rosterRows(page int) []ui.Item {
	var rows []ui.Item
	current := c.State.Attempts[page].PersonID
	labels := studentLabels(c.State.Roster)
	for i, student := range c.State.Roster {
		row := ui.Item{ID: student.PersonID, Label: labels[i], Selected: student.PersonID == current, Disabled: c.State.Status[student.PersonID] != ""}
		row.SuccessChecks = c.State.Successes[student.PersonID]
		if row.Selected {
			row.Fill = ui.Green
		}
		rows = append(rows, row)
	}
	return rows
}

// Layout is chosen by the assessment root.
type Layout struct{ Panels bool }
