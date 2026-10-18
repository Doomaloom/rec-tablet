package assessment

import (
	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/navigation"
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
	if mode == "single" {
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
	params := navigation.Destination(chosen)
	if err := c.navigation.Run(func() error {
		c.navigation.Push(navigation.SkillSwimmers, func(f *navigation.Frame) {
			f.Skill, f.Level, f.Version = chosen.ID, params.Level, params.Version
		})
		return nil
	}); err != nil {
		return err
	}
	switch mode {
	case "multi":
		return c.EnterMultiMode()
	case "class":
		return c.EnterClassMode()
	}
	return nil
}
