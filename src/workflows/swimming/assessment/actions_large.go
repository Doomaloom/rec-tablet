package assessment

import (
	"strconv"

	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) openLargeSkillSwimmer(skill navigation.Row, person string) error {
	params := navigation.Destination(skill)
	if _, timed := benchmarkFor(skill.ID); timed {
		if err := c.openLargeTimedSkill(skill.ID); err != nil {
			return err
		}
		return c.navigation.Select(person)
	}
	return c.navigation.Run(func() error {
		progress, err := c.store.Progress(person, skill.ID, params.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.store.NewAttempt(person, c.navigation.Frame().Lesson, skill.ID)
		if err != nil {
			return err
		}
		c.navigation.Push(navigation.Assessment, func(f *navigation.Frame) {
			f.Person, f.Attempt, f.Skill, f.Level, f.Version = person, attempt, skill.ID, params.Level, params.Version
		})
		return nil
	})
}

func (c *Component) openLargeTimedSkill(id string) error {
	for _, row := range c.navigation.State.Rows {
		if row.ID != id {
			continue
		}
		params := navigation.Destination(row)
		return c.navigation.Run(func() error {
			c.skillPicker.selectedID = id
			c.timed.selection = nil
			c.timed.lapsInput = ""
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
		if row.ID == c.skillPicker.selectedID {
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

func (c *Component) switchLargeAssessmentPerson(person string) {
	if person == c.navigation.Frame().Person {
		return
	}
	_ = c.navigation.Run(func() error {
		f := &c.navigation.History.Frames[len(c.navigation.History.Frames)-1]
		progress, err := c.store.Progress(person, f.Skill, f.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.store.NewAttempt(person, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		f.Person, f.Attempt, f.Page = person, attempt, 0
		return nil
	})
}

func (c *Component) toggleLargeCorrectionSubskill(id string) error {
	if c.navigation.Frame().Screen != navigation.Assessment || !c.navigation.Frame().Review {
		return nil
	}
	status := "successful"
	for _, row := range c.navigation.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "not_yet"
			break
		}
	}
	return c.navigation.Run(func() error { return c.store.Mark(c.navigation.Frame().Attempt, id, status) })
}

func (c *Component) startLargeTimedHeat() error {
	laps, err := strconv.Atoi(c.timed.lapsInput)
	if err != nil || laps < 1 || laps > 999 || len(c.timed.selection) == 0 {
		return nil
	}
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		participants, err := c.store.NewTimedHeat(c.timed.selection, f.Lesson, f.Skill, laps)
		if err != nil {
			return err
		}
		heat := &timedHeatState{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps}
		for _, person := range participants {
			heat.Participants = append(heat.Participants, timedParticipantState{participant: timedParticipant{
				AttemptID: person.AttemptID, PersonID: person.PersonID, Name: person.Name, LapsRequired: person.LapsRequired,
			}})
		}
		c.timed.heat, c.timed.heatPage = heat, 0
		c.navigation.Push(navigation.TimedHeat, nil)
		return nil
	})
}
