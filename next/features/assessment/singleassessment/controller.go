package singleassessment

import (
	"time"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// Component owns individual assessment state and presentation through explicit data and action callbacks.
type Component struct {
	Layout     Layout
	progress   common.ProgressActions
	navigation common.Navigation
	State      State
	storage    Storage
	paginate   func(*navigation.State, []navigation.Row)
	small      *CompactTemplate
	large      *LargeTemplate
	shared     *ListTemplate
}

// State holds mode session and paging state.
type State struct {
	RosterPage int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	DiscardAttempt      func(string) error
	Draft               func(string) (map[string]string, error)
	EditAttempt         func(string) error
	GetAttempt          func(string) (domain.Attempt, error)
	GetEnrollment       func(string) (domain.Enrollment, error)
	ListAttempts        func(string, string, string) ([]domain.Attempt, error)
	Mark                func(string, string, string) error
	NewAttempt          func(string, string, string) (string, error)
	Progress            func(string, string, string) (domain.Progress, error)
	RemoveLatestSuccess func(string, string, string) error
	Skill               func(string, string) (curriculum.Skill, error)
	Submit              func(string) (bool, error)
}

// Options provides storage and navigation operations without a parent-component dependency.
type Options struct {
	Navigation  common.Navigation
	Storage     Storage
	ViewActions common.Actions
	Paginate    func(*navigation.State, []navigation.Row)
}

func New(options Options) *Component {
	c := &Component{navigation: options.Navigation, storage: options.Storage, paginate: options.Paginate}
	c.progress = common.ProgressActions{Navigation: options.Navigation, Remove: options.Storage.RemoveLatestSuccess}
	c.small = &CompactTemplate{Actions: options.ViewActions}
	c.shared = &ListTemplate{Actions: options.ViewActions}
	c.large = &LargeTemplate{Actions: options.ViewActions,
		Pages: Pages{Roster: &c.State.RosterPage}, SwitchPerson: c.SwitchPerson,
	}
	return c
}

func (c *Component) Build(input common.Input) (navigation.State, error) {
	state := input.State
	f := input.Frame
	roster := input.PresentRoster()
	version := input.Course.CurriculumVersion
	rows := []navigation.Row{}
	s := c.storage
	switch f.Screen {
	case navigation.Attempts:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		navigation.AppendPrimaryInfo(&state, skill.Name)
		progress, e := s.Progress(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		state.Progress = navigation.ProgressText(progress)
		attempts, e := s.ListAttempts(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		for _, a := range attempts {
			label := "Not yet successful"
			if a.Success {
				label = "Successful"
			}
			if a.SubmittedAt == nil {
				label = "Unfinished attempt"
			} else if a.Editing {
				label += " · Edit saved"
			}
			created := a.CreatedAt
			clock := created
			if len(created) >= 19 {
				clock = created[11:19]
			}
			rows = append(rows, navigation.Row{ID: a.ID, Label: label + " · " + clock + " UTC"})
		}
		state.Empty = "No attempts recorded for this skill."
		if c.Layout.Panels {
			state.PageSize = 6
		}
	case navigation.Assessment:
		a, e := s.GetAttempt(f.Attempt)
		if e != nil {
			return state, e
		}
		p, e := s.GetEnrollment(a.EnrollmentID)
		if e != nil {
			return state, e
		}
		skill, e := s.Skill(a.CurriculumVersion, a.SkillID)
		if e != nil {
			return state, e
		}
		answers, e := s.Draft(f.Attempt)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.SplitHeader = &contracts.SplitHeader{Left: p.Name, Right: skill.Name}
		state.PrimaryInfo = ""
		state.Progress = ""
		if c.Layout.Panels {
			ConfigureLargeHeader(&state, skill)
		}
		for _, subskill := range skill.Subskills {
			status := answers[subskill.ID]
			if status == "" {
				status = "unmarked"
			}
			rows = append(rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: status})
		}
		if f.Review {
			state.FooterAction = "Save correction"
			state.CanSubmit = len(answers) == len(rows)
			state.PageSize = 1
		} else {
			state.HomeRailAction = common.SaveRailAction(c.SaveAttempt)
			state.BackRailAction = common.DiscardRailAction(c.DiscardAttempt)
			state.PageSize = 3
		}
		if c.Layout.Panels {
			state.PageSize = max(1, len(rows))
			if f.Review {
				state.ComposeBody = c.large.ComposeLargeChecklistBody(rows, func(id string) { _ = c.ToggleCorrection(id) })
			} else {
				var swimmers []navigation.Row
				for _, person := range roster {
					if !person.Active || person.Level != p.Level {
						continue
					}
					progress, err := s.Progress(person.ID, skill.ID, a.CurriculumVersion)
					if err != nil {
						return state, err
					}
					row := navigation.Row{ID: person.ID, Label: person.Name, ProgressSegments: min(3, max(0, progress.Successes)), Selected: person.ID == p.ID}
					if row.Selected {
						row.Fill = ui.Green
					}
					if progress.Successes > 0 {
						student := person.ID
						row.OnHold = func() { _ = c.progress.RemoveLatestSuccess(student, skill.ID, a.CurriculumVersion) }
						row.HoldDuration = 2 * time.Second
					}
					swimmers = append(swimmers, row)
				}
				state.ComposeBody = c.large.ComposeLargeSingleBody(rows, swimmers)
			}
		}
	}
	c.paginate(&state, rows)
	if state.ComposeBody == nil {
		c.compose(&state, f)
	}
	return state, nil
}

// Layout is chosen by the assessment root.
type Layout struct{ Panels bool }
