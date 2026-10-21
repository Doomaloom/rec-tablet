package singleassessment

import (
	"rectablet/next/navigation"
)

func (c *Component) ToggleSubskill(id string) error {
	f := c.navigation.Frame()
	status := "successful"
	for _, row := range c.navigation.Rows() {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
			break
		}
	}
	if e := c.storage.Mark(f.Attempt, id, status); e != nil {
		// Keep the checklist usable if a local write fails; the persisted draft
		// remains unchanged and the next refresh shows its previous state.
		return nil
	}
	c.navigation.Saved("")
	return nil
}

// SaveAttempt stores an incomplete attempt without counting it toward progress.
func (c *Component) SaveAttempt() error {
	if c.navigation.Frame().Screen != navigation.Assessment || c.navigation.Frame().Review {
		return nil
	}
	return c.navigation.Run(func() error {
		if _, e := c.storage.Submit(c.navigation.Frame().Attempt); e != nil {
			return nil
		}
		c.navigation.Pop()
		return nil
	})
}

// DiscardAttempt removes the unfinished draft and returns to the student list.
func (c *Component) DiscardAttempt() error {
	if c.navigation.Frame().Screen != navigation.Assessment || c.navigation.Frame().Review {
		return nil
	}
	return c.navigation.Run(func() error {
		if e := c.storage.DiscardAttempt(c.navigation.Frame().Attempt); e != nil {
			return nil
		}
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) SwitchPerson(person string) {
	if person == c.navigation.Frame().Person {
		return
	}
	_ = c.navigation.Run(func() error {
		f := c.navigation.Frame()
		progress, err := c.storage.Progress(person, f.Skill, f.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.storage.NewAttempt(person, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		f.Person, f.Attempt, f.Page = person, attempt, 0
		c.navigation.UpdateFrame(func(frame *navigation.Frame) { *frame = f })
		return nil
	})
}

func (c *Component) ToggleCorrection(id string) error {
	if c.navigation.Frame().Screen != navigation.Assessment || !c.navigation.Frame().Review {
		return nil
	}
	status := "successful"
	for _, row := range c.navigation.Rows() {
		if row.ID == id && row.Status == "successful" {
			status = "not_yet"
			break
		}
	}
	return c.navigation.Run(func() error { return c.storage.Mark(c.navigation.Frame().Attempt, id, status) })
}

func (c *Component) Mark(status string) error {
	if c.navigation.Frame().Screen != navigation.Assessment || len(c.navigation.Rows()) == 0 {
		return nil
	}
	if e := c.storage.Mark(c.navigation.Frame().Attempt, c.navigation.Rows()[0].ID, status); e != nil {
		return nil
	}
	c.navigation.Saved("")
	return nil
}

func (c *Component) Submit() error {
	if c.navigation.Frame().Screen != navigation.Assessment {
		return nil
	}
	_, e := c.storage.Submit(c.navigation.Frame().Attempt)
	if e != nil {
		return nil
	}
	c.navigation.Saved("")
	c.navigation.Pop()
	return nil
}

// Select handles individual swimmer selection, checklist edits, and attempt review.
func (c *Component) Select(id string) error {
	f := c.navigation.Frame()
	var row navigation.Row
	for _, candidate := range c.navigation.Rows() {
		if candidate.ID == id {
			row = candidate
			break
		}
	}
	switch f.Screen {
	case navigation.SkillSwimmers:
		params := navigation.Destination(row)
		version := params.Version
		if version == "" {
			version = f.Version
		}
		skillID := f.Skill
		if params.Skill != "" {
			skillID = params.Skill
		}
		progress, err := c.storage.Progress(id, skillID, version)
		if err != nil {
			return nil
		}
		if progress.Complete {
			c.navigation.Saved("")
			return nil
		}
		attempt, err := c.storage.NewAttempt(id, f.Lesson, skillID)
		if err != nil {
			return nil
		}
		c.navigation.Push(navigation.Assessment, func(next *navigation.Frame) {
			next.Attempt, next.Person, next.Skill, next.Level, next.Version = attempt, id, skillID, f.Level, version
		})
	case navigation.Attempts:
		if err := c.storage.EditAttempt(id); err != nil {
			return nil
		}
		c.navigation.Push(navigation.Assessment, func(next *navigation.Frame) { next.Attempt = id; next.Review = true })
	case navigation.Assessment:
		return c.ToggleSubskill(id)
	}
	return nil
}

func (c *Component) OpenSwimmer(skill navigation.Row, person string) error {
	params := navigation.Destination(skill)
	return c.navigation.Run(func() error {
		progress, err := c.storage.Progress(person, skill.ID, params.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.storage.NewAttempt(person, c.navigation.Frame().Lesson, skill.ID)
		if err != nil {
			return err
		}
		c.navigation.Push(navigation.Assessment, func(frame *navigation.Frame) {
			frame.Person, frame.Attempt, frame.Skill, frame.Level, frame.Version = person, attempt, skill.ID, params.Level, params.Version
		})
		return nil
	})
}
