package assessment

import (
	"strings"

	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) Select(id string) error {
	f := c.navigation.Frame()
	var row navigation.Row
	for _, r := range c.navigation.State.Rows {
		if r.ID == id {
			row = r
			break
		}
	}
	switch f.Screen {
	case navigation.Skills:
		params := navigation.Destination(row)
		c.timed.selection = nil
		c.navigation.Push(navigation.SkillSwimmers, func(n *navigation.Frame) {
			n.Skill = id
			n.Level = params.Level
			n.Version = params.Version
		})

	case navigation.SkillSwimmers:
		if _, timed := benchmarkFor(f.Skill); timed {
			if containsString(c.timed.selection, id) {
				c.timed.selection = removeString(c.timed.selection, id)
			} else {
				c.timed.selection = append(c.timed.selection, id)
			}
			c.navigation.Saved("")
			return nil
		}
		// Skill taps are best-effort on the device. Failed reads or writes leave
		// the current list intact without surfacing an error state.
		params := navigation.Destination(row)
		version := params.Version
		if version == "" {
			version = f.Version
		}
		skillID := f.Skill
		if params.Skill != "" {
			skillID = params.Skill
		}
		progress, e := c.store.Progress(id, skillID, version)
		if e != nil {
			return nil
		}
		if progress.Complete {
			c.navigation.Saved("")
			return nil
		}
		attempt, e := c.store.NewAttempt(id, f.Lesson, skillID)
		if e != nil {
			return nil
		}
		c.navigation.Push(navigation.Assessment, func(n *navigation.Frame) {
			n.Attempt = attempt
			n.Person = id
			n.Skill = skillID
			n.Level = f.Level
			n.Version = version
		})

	case navigation.MultiSelection:
		return c.multi.toggleMultiMember(id)
	case navigation.MultiStudent:
		return c.multi.toggleMultiSubskill(id)

	case navigation.Attempts:
		if e := c.store.EditAttempt(id); e != nil {
			return nil
		}
		c.navigation.Push(navigation.Assessment, func(n *navigation.Frame) { n.Attempt = id; n.Review = true })

	case navigation.Assessment:
		return c.toggleSubskill(id)

	case navigation.ClassAssessment:
		return c.toggleClassSubskill(id)

	case navigation.TimedLaps:
		if len(id) == len("timed-digit:")+1 && strings.HasPrefix(id, "timed-digit:") {
			c.timed.lapsInput += strings.TrimPrefix(id, "timed-digit:")
			if len(c.timed.lapsInput) > 3 {
				c.timed.lapsInput = c.timed.lapsInput[:3]
			}
			return nil
		}
		if id == "timed-clear" {
			c.timed.lapsInput = ""
		}

	case navigation.TimedHeat:
		if strings.HasPrefix(id, "timed-lap:") {
			return c.RecordTimedLap(strings.TrimPrefix(id, "timed-lap:"))
		}
	}
	return nil
}

func (c *Component) RemoveLatestSuccess(person, skill, version string) error {
	return c.navigation.Run(func() error {
		if e := c.store.RemoveLatestSuccess(person, skill, version); e != nil {
			return nil
		}
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) Header() error { return nil }

func (c *Component) Mark(status string) error {
	if c.navigation.Frame().Screen != navigation.Assessment || len(c.navigation.State.Rows) == 0 {
		return nil
	}
	if e := c.store.Mark(c.navigation.Frame().Attempt, c.navigation.State.Rows[0].ID, status); e != nil {
		return nil
	}
	c.navigation.Saved("")
	return nil
}

func (c *Component) Submit() error {
	if c.navigation.Frame().Screen == navigation.MultiGroupCount {
		return c.chooseMultiGroupCount()
	}
	if c.navigation.Frame().Screen == navigation.MultiSelection {
		return c.ConfirmMultiSelection()
	}
	if c.navigation.Frame().Screen != navigation.Assessment {
		return nil
	}
	_, e := c.store.Submit(c.navigation.Frame().Attempt)
	if e != nil {
		return nil
	}
	c.navigation.Saved("")
	c.navigation.Pop()
	return nil
}
