package assessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
)

func (c *Component) EnterClassMode() error  { return c.class.Enter() }
func (c *Component) EnterMultiMode() error  { return c.multi.Enter() }
func (c *Component) EnterSingleMode() error { return c.multi.CancelSetup() }

func (c *Component) openLargeSkillSwimmer(skill navigation.Row, person string) error {
	if _, timed := common.BenchmarkFor(skill.ID); timed {
		if err := c.openLargeTimedSkill(skill.ID); err != nil {
			return err
		}
		return c.navigation.Select(person)
	}
	return c.single.OpenSwimmer(skill, person)
}

func (c *Component) openLargeTimedSkill(id string) error {
	for _, row := range c.navigation.State.Rows {
		if row.ID != id {
			continue
		}
		params := navigation.Destination(row)
		return c.navigation.Run(func() error {
			c.skillPicker.State.SelectedID = id
			c.timed.State.Selection = nil
			c.timed.State.LapsInput = ""
			c.navigation.Push(navigation.SkillSwimmers, func(f *navigation.Frame) { f.Skill, f.Level, f.Version = id, params.Level, params.Version })
			return nil
		})
	}
	return nil
}

func (c *Component) openLargeSelectedMode(mode string) error {
	if c.navigation.Frame().Screen != navigation.Skills {
		return nil
	}
	if mode != "class" {
		return nil
	}
	var chosen navigation.Row
	for _, row := range c.navigation.State.Rows {
		if row.ID == c.skillPicker.State.SelectedID {
			chosen = row
			break
		}
	}
	if chosen.ID == "" {
		return nil
	}
	if _, timed := common.BenchmarkFor(chosen.ID); timed {
		return c.openLargeTimedSkill(chosen.ID)
	}
	return c.class.OpenSkill(chosen)
}

// OpenPlannedSkill shares the existing assessment sessions with the lesson planner.
func (c *Component) OpenPlannedSkill(skill navigation.Row) error {
	if c.navigation.Frame().Screen != navigation.LessonPlanDetail || skill.Disabled {
		return nil
	}
	if _, timed := common.BenchmarkFor(skill.ID); timed {
		params := navigation.Destination(skill)
		return c.navigation.Run(func() error {
			c.timed.State.Selection = nil
			c.timed.State.LapsInput = ""
			c.navigation.Push(navigation.SkillSwimmers, func(f *navigation.Frame) { f.Skill, f.Level, f.Version = skill.ID, params.Level, params.Version })
			return nil
		})
	}
	return c.class.OpenSkill(skill)
}
