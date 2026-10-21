package skillpicker

import (
	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/features/assessment/internal/pagedgrid"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
)

type LargeTemplate struct {
	Actions          common.Actions
	Pages            Pages
	IsTimed          func(string) bool
	OpenTimedSkill   func(string) error
	OpenSkillSwimmer func(navigation.Row, string) error
	SelectSkill      func(string) error
}

func (c *LargeTemplate) ComposeLargeSkillsBody(skills []navigation.Row, selected navigation.Row, swimmers []navigation.Row) ui.BodyComposer {
	return func(surface *ui.Surface) {
		halfSkills := (len(skills) + 1) / 2
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
		if len(skillRows) <= 6 {
			c.largePagedGrid(ui.SelectableSurface(surface), "large-skills", skillRows, 1, len(skills), c.Pages.Skill, func(id string) {
				if c.IsTimed(id) {
					_ = c.OpenTimedSkill(id)
					return
				}
				_ = c.SelectSkill(id)
			})
		} else {
			for column := range 2 {
				panel := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
					x, y, width, height := surface.BodyBounds(w, h)
					cellWidth := (width - ui.Gap) / 2
					return x + float32(column)*(cellWidth+ui.Gap), y, cellWidth, height
				})
				start := column * halfSkills
				c.largePagedGrid(ui.SelectableSurface(panel), "large-skills", skillRows[start:min(start+halfSkills, len(skillRows))], 1, halfSkills, c.Pages.Skill, func(id string) {
					if c.IsTimed(id) {
						_ = c.OpenTimedSkill(id)
						return
					}
					_ = c.SelectSkill(id)
				})
			}
		}
	}
}
func (c *LargeTemplate) ComposeLargeSwimmerListBody(state contracts.State) ui.BodyComposer {
	return func(surface *ui.Surface) {
		c.largePagedGrid(surface, "large-return-swimmers", state.Rows, 2, 6, c.Pages.Roster, func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *LargeTemplate) largePagedGrid(surface *ui.Surface, key string, rows []ui.Item, columns, rowCount int, page *int, selectRow func(string)) {
	pagedgrid.Controller{Run: c.Actions.Run}.Render(surface, key, rows, columns, rowCount, page, selectRow)
}

type Pages struct{ Skill, Roster *int }
type ListTemplate struct {
	Actions common.Actions
}

func (c *ListTemplate) ComposeAssessmentListBody(state contracts.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.Grid(surface, snapshot.Rows, 1, max(1, snapshot.PageSize), func(id string) { _ = c.Actions.Select(id) })
	}
}
func (c *ListTemplate) ComposeSkillListBody(state contracts.State) ui.BodyComposer {
	compose := c.ComposeAssessmentListBody(state)
	return func(surface *ui.Surface) { compose(ui.SelectableSurface(surface)) }
}
func (c *Component) compose(state *contracts.State, f navigation.Frame) {
	switch f.Screen {
	case navigation.Skills:
		state.ComposeBody = c.shared.ComposeSkillListBody(*state)
	case navigation.SkillSwimmers:
		state.ComposeBody = c.ComposeSwimmers(*state)
	}
}
func (c *Component) ComposeSwimmers(state contracts.State) ui.BodyComposer {
	if c.Layout.Panels {
		return c.large.ComposeLargeSwimmerListBody(state)
	}
	return c.shared.ComposeAssessmentListBody(state)
}

// Build prepares the screen state, then paginates before composing list templates.
func (c *Component) Build(input common.Input) (navigation.State, error) {
	state := input.State
	rows := []navigation.Row{}
	var err error
	switch input.Frame.Screen {
	case navigation.Skills:
		rows, err = c.buildSkills(&state, input)
	case navigation.SkillSwimmers:
		rows, err = c.buildSwimmers(&state, input)
	}
	if err != nil {
		return state, err
	}
	c.paginate(&state, rows)
	if state.ComposeBody == nil {
		c.compose(&state, input.Frame)
	}
	return state, nil
}

func (c *Component) buildSkills(state *navigation.State, input common.Input) ([]navigation.Row, error) {
	state.Title = "Choose a Skill"
	state.PrimaryInfo = ""
	version := input.Course.CurriculumVersion
	rows, err := c.loadSkills(input.Roster, version)
	if err != nil {
		return nil, err
	}
	state.Empty = "No enrolled swimmers with an assigned level."
	if c.Layout.Panels {
		state.PageSize = max(1, len(rows))
		selection, err := c.loadSelection(rows, input.Roster, version)
		if err != nil {
			return nil, err
		}
		state.HeaderSwitch = nil
		state.ComposeBody = c.large.ComposeLargeSkillsBody(rows, selection.selected, selection.swimmers)
	}
	return rows, nil
}

func (c *Component) buildSwimmers(state *navigation.State, input common.Input) ([]navigation.Row, error) {
	state.PrimaryInfo = ""
	skill, version, err := c.resolveSkill(input.Frame, input.Course.CurriculumVersion)
	if err != nil {
		return nil, err
	}
	state.Title = "Choose a Swimmer"
	state.HeaderSwitch = c.actions.ModeSwitch(skill, false)
	rows, err := c.loadSwimmers(input.Roster, input.Frame.Level, skill.ID, version)
	if err != nil {
		return nil, err
	}
	state.Empty = "No enrolled swimmers are assigned to this skill's level."
	if c.Layout.Panels {
		state.PageSize = max(1, len(rows))
	}
	return rows, nil
}

func missingCurriculumRow(level string) navigation.Row {
	return navigation.Row{ID: level, Label: level + " · Missing curriculum · update swimmer level", Disabled: true}
}

func skillRow(skill curriculum.Skill, level, version string) navigation.Row {
	label := skill.Number + ". " + curriculum.CompactTitle(skill)
	if benchmark, ok := common.BenchmarkFor(skill.ID); ok {
		label += " · Timed heat · " + benchmark.Description
	}
	return navigation.Row{ID: skill.ID, Label: label, Data: navigation.Parameters{Level: level, Version: version}}
}

func swimmerRow(person domain.Enrollment, skill, version string, progress domain.Progress) navigation.Row {
	return navigation.Row{
		ID:               person.ID,
		Label:            person.Name,
		ProgressSegments: min(3, max(0, progress.Successes)),
		Data:             navigation.Parameters{Skill: skill, Level: person.Level, Version: version},
	}
}
