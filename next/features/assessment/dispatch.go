package assessment

import (
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
)

func (c *Component) Select(id string) error {
	switch c.navigation.Frame().Screen {
	case navigation.Skills:
		var row navigation.Row
		for _, candidate := range c.navigation.State.Rows {
			if candidate.ID == id {
				row = candidate
				break
			}
		}
		params := navigation.Destination(row)
		c.timed.State.Selection = nil
		c.navigation.Push(navigation.SkillSwimmers, func(frame *navigation.Frame) {
			frame.Skill, frame.Level, frame.Version = id, params.Level, params.Version
		})
	case navigation.SkillSwimmers:
		if _, timed := common.BenchmarkFor(c.navigation.Frame().Skill); timed {
			return c.timed.Select(id)
		}
		return c.single.Select(id)
	case navigation.Attempts, navigation.Assessment:
		return c.single.Select(id)
	case navigation.MultiSelection:
		return c.multi.ToggleMember(id)
	case navigation.MultiStudent:
		return c.multi.ToggleSubskill(id)
	case navigation.ClassAssessment:
		return c.class.ToggleSubskill(id)
	case navigation.TimedLaps, navigation.TimedHeat:
		return c.timed.Select(id)
	}
	return nil
}

func (c *Component) Header() error            { return nil }
func (c *Component) Mark(status string) error { return c.single.Mark(status) }
func (c *Component) Submit() error {
	switch c.navigation.Frame().Screen {
	case navigation.MultiGroupCount:
		return c.multi.ChooseGroupCount()
	case navigation.MultiSelection:
		return c.multi.ConfirmSelection()
	default:
		return c.single.Submit()
	}
}
