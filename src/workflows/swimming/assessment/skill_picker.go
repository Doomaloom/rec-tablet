package assessment

import (
	"time"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/large"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// SkillPicker owns its mode state and presentation. Data and actions arrive as callbacks.
type SkillPicker struct {
	selectedID            string
	skillPage, rosterPage int
	data                  skillPickerData
	actions               skillPickerActions
	paginate              func(*navigation.State, []navigation.Row)
	large                 *large.View
	shared                *shared.View
}

type skillPickerData struct {
	CatalogFor func(string) (curriculum.Catalog, error)
	Progress   func(string, string, string) (domain.Progress, error)
	Skill      func(string, string) (curriculum.Skill, error)
}

type skillPickerActions struct {
	RemoveLatestSuccess   func(string, string, string) error
	openLargeSelectedMode func(string) error
	skillModeSwitch       func(curriculum.Skill, bool) []shell.HeaderSwitchOption
}

func (c *SkillPicker) build(input modeInput) (navigation.State, error) {
	state := input.state
	f := input.frame
	roster := input.roster
	version := input.course.CurriculumVersion
	rows := []navigation.Row{}
	s := c.data
	switch f.Screen {
	case navigation.Skills:
		state.Title = "Choose a Skill"
		state.PrimaryInfo = ""
		catalog, e := s.CatalogFor(version)
		if e != nil {
			return state, e
		}
		seenLevels := map[string]bool{}
		for _, p := range roster {
			if !p.Active || p.Level == "" || seenLevels[p.Level] {
				continue
			}
			seenLevels[p.Level] = true
			level, ok := catalog.Level(p.Level)
			if !ok || len(level.Skills) == 0 {
				rows = append(rows, navigation.Row{ID: p.Level, Label: p.Level + " · Missing curriculum · update swimmer level", Disabled: true})
				continue
			}
			for _, skill := range level.Skills {
				label := skill.Number + ". " + curriculum.CompactTitle(skill)
				if benchmark, ok := benchmarkFor(skill.ID); ok {
					label += " · Timed heat · " + benchmark.Description
				}
				rows = append(rows, navigation.Row{ID: skill.ID, Label: label, Data: navigation.Parameters{Level: p.Level, Version: version}})
			}
		}
		state.Empty = "No enrolled swimmers with an assigned level."
		if state.DisplayMode.IsLarge() {
			state.PageSize = max(1, len(rows))
			var selected navigation.Row
			for _, row := range rows {
				if row.ID == c.selectedID && !row.Disabled {
					selected = row
					break
				}
			}
			if selected.ID == "" {
				for _, row := range rows {
					if !row.Disabled {
						selected = row
						break
					}
				}
			}
			c.selectedID = selected.ID
			var swimmers []navigation.Row
			if selected.ID != "" {
				params := navigation.Destination(selected)
				for _, person := range roster {
					if !person.Active || person.Level != params.Level {
						continue
					}
					progress, err := s.Progress(person.ID, selected.ID, version)
					if err != nil {
						return state, err
					}
					row := navigation.Row{ID: person.ID, Label: person.Name, ProgressSegments: min(3, max(0, progress.Successes))}
					if progress.Successes > 0 {
						student, skillID := person.ID, selected.ID
						row.OnHold = func() { _ = c.actions.RemoveLatestSuccess(student, skillID, version) }
						row.HoldDuration = 2 * time.Second
					}
					swimmers = append(swimmers, row)
				}
				if skill, err := s.Skill(version, selected.ID); err == nil {
					if _, timed := benchmarkFor(skill.ID); !timed {
						state.HeaderSwitch = []shell.HeaderSwitchOption{{Label: "Single", Selected: true, Run: func() error { return c.actions.openLargeSelectedMode("single") }}}
						if allowsMultiMode(skill) {
							state.HeaderSwitch = append(state.HeaderSwitch, shell.HeaderSwitchOption{Label: "Multi", Run: func() error { return c.actions.openLargeSelectedMode("multi") }})
						}
						state.HeaderSwitch = append(state.HeaderSwitch, shell.HeaderSwitchOption{Label: "Class", Run: func() error { return c.actions.openLargeSelectedMode("class") }})
					}
				}
			}
			state.ComposeBody = c.large.ComposeLargeSkillsBody(rows, selected, swimmers)
		}
	case navigation.SkillSwimmers:
		state.PrimaryInfo = ""
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = "Choose a Swimmer"

		state.HeaderSwitch = c.actions.skillModeSwitch(skill, false)
		for _, p := range roster {
			if !p.Active || p.Level != f.Level {
				continue
			}
			progress, e := s.Progress(p.ID, skill.ID, version)
			if e != nil {
				return state, e
			}
			row := navigation.Row{
				ID:               p.ID,
				Label:            p.Name,
				ProgressSegments: min(3, max(0, progress.Successes)),
				Data:             navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version},
			}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.actions.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		state.Empty = "No enrolled swimmers are assigned to this skill's level."
		if state.DisplayMode.IsLarge() {
			state.PageSize = max(1, len(rows))
		}

	}
	c.paginate(&state, rows)
	if state.ComposeBody == nil {
		c.compose(&state, f)
	}
	return state, nil
}

func (c *SkillPicker) compose(state *shell.State, f navigation.Frame) {
	switch f.Screen {
	case navigation.Skills:
		state.ComposeBody = c.shared.ComposeSkillListBody(*state)
	case navigation.SkillSwimmers:
		state.ComposeBody = c.composeSwimmerListBody(*state)
	}
}

func (c *SkillPicker) composeSwimmerListBody(state shell.State) ui.BodyComposer {
	if state.DisplayMode.IsLarge() {
		return c.large.ComposeLargeSwimmerListBody(state)
	}
	return c.shared.ComposeAssessmentListBody(state)
}
