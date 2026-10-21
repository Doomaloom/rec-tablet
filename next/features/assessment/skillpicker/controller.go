package skillpicker

import (
	"time"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

// Component owns skill selection state and presentation through explicit data and action callbacks.
type Component struct {
	Layout     Layout
	progress   common.ProgressActions
	navigation common.Navigation
	State      State
	storage    Storage
	actions    Actions
	paginate   func(*navigation.State, []navigation.Row)
	large      *LargeTemplate
	shared     *ListTemplate
}

// State holds mode session and paging state.
type State struct {
	SelectedID            string
	SkillPage, RosterPage int
}

// Storage contains only the data operations consumed by this mode.
type Storage struct {
	CatalogFor          func(string) (curriculum.Catalog, error)
	Progress            func(string, string, string) (domain.Progress, error)
	RemoveLatestSuccess func(string, string, string) error
	Skill               func(string, string) (curriculum.Skill, error)
}

type Actions struct {
	OpenSelectedMode func(string) error
	ModeSwitch       func(curriculum.Skill, bool) []contracts.HeaderSwitchOption
	OpenTimedSkill   func(string) error
	OpenSkillSwimmer func(navigation.Row, string) error
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
	c.progress = common.ProgressActions{Navigation: options.Navigation, Remove: options.Storage.RemoveLatestSuccess}
	c.shared = &ListTemplate{Actions: options.ViewActions}
	c.large = &LargeTemplate{Actions: options.ViewActions,
		Pages:            Pages{Skill: &c.State.SkillPage, Roster: &c.State.RosterPage},
		IsTimed:          func(id string) bool { _, ok := common.BenchmarkFor(id); return ok },
		OpenTimedSkill:   c.actions.OpenTimedSkill,
		OpenSkillSwimmer: c.actions.OpenSkillSwimmer,
		SelectSkill: func(id string) error {
			c.State.SelectedID = id
			c.State.RosterPage = 0
			return c.actions.OpenSelectedMode("class")
		},
	}
	return c
}

// skillSelection is the data needed by the two-panel skill picker.
type skillSelection struct {
	selected navigation.Row
	swimmers []navigation.Row
	skill    *curriculum.Skill
}

func (c *Component) loadSkills(roster []domain.Enrollment, version string) ([]navigation.Row, error) {
	catalog, err := c.storage.CatalogFor(version)
	if err != nil {
		return nil, err
	}
	rows := []navigation.Row{}
	seenLevels := map[string]bool{}
	for _, person := range roster {
		if !person.Active || person.Level == "" || seenLevels[person.Level] {
			continue
		}
		seenLevels[person.Level] = true
		level, ok := catalog.Level(person.Level)
		if !ok || len(level.Skills) == 0 {
			rows = append(rows, missingCurriculumRow(person.Level))
			continue
		}
		for _, skill := range level.Skills {
			rows = append(rows, skillRow(skill, person.Level, version))
		}
	}
	return rows, nil
}

func (c *Component) loadSelection(rows []navigation.Row, roster []domain.Enrollment, version string) (skillSelection, error) {
	selected := selectedSkill(rows, c.State.SelectedID)
	c.State.SelectedID = selected.ID
	selection := skillSelection{selected: selected}
	if selected.ID == "" {
		return selection, nil
	}
	params := navigation.Destination(selected)
	swimmers, err := c.loadSwimmers(roster, params.Level, selected.ID, version)
	if err != nil {
		return selection, err
	}
	// Panel rows open swimmers through OpenSkillSwimmer rather than row destinations.
	for i := range swimmers {
		swimmers[i].Data = nil
	}
	selection.swimmers = swimmers
	if skill, err := c.storage.Skill(version, selected.ID); err == nil {
		selection.skill = &skill
	}
	return selection, nil
}

func selectedSkill(rows []navigation.Row, selectedID string) navigation.Row {
	var first navigation.Row
	for _, row := range rows {
		if row.Disabled {
			continue
		}
		if row.ID == selectedID {
			return row
		}
		if first.ID == "" {
			first = row
		}
	}
	return first
}

func (c *Component) resolveSkill(frame navigation.Frame, version string) (curriculum.Skill, string, error) {
	if frame.Version != "" {
		version = frame.Version
	}
	skill, err := c.storage.Skill(version, frame.Skill)
	return skill, version, err
}

func (c *Component) loadSwimmers(roster []domain.Enrollment, level, skillID, version string) ([]navigation.Row, error) {
	rows := []navigation.Row{}
	for _, person := range roster {
		if !person.Active || person.Level != level {
			continue
		}
		progress, err := c.storage.Progress(person.ID, skillID, version)
		if err != nil {
			return nil, err
		}
		row := swimmerRow(person, skillID, version, progress)
		c.bindProgressCorrection(&row, person.ID, skillID, version, progress.Successes)
		rows = append(rows, row)
	}
	return rows, nil
}

func (c *Component) bindProgressCorrection(row *navigation.Row, person, skill, version string, successes int) {
	if successes > 0 {
		row.OnHold = func() { _ = c.progress.RemoveLatestSuccess(person, skill, version) }
		row.HoldDuration = 2 * time.Second
	}
}

// Layout is chosen by the assessment root.
type Layout struct{ Panels bool }
