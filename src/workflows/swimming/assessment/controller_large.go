package assessment

import (
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/navigation"
	"strconv"
)

func (c *Controller) openLargeSkillSwimmer(skill navigation.Row, person string) error {
	params := navigation.Destination(skill)
	if _, timed := benchmarkFor(skill.ID); timed {
		if err := c.openLargeTimedSkill(skill.ID); err != nil {
			return err
		}
		return c.Controller.Select(person)
	}
	return c.Run(func() error {
		progress, err := c.Store.Progress(person, skill.ID, params.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.Store.NewAttempt(person, c.Frame().Lesson, skill.ID)
		if err != nil {
			return err
		}
		c.Push(navigation.Assessment, func(f *navigation.Frame) {
			f.Person, f.Attempt, f.Skill, f.Level, f.Version = person, attempt, skill.ID, params.Level, params.Version
		})
		return nil
	})
}

func (c *Controller) openLargeTimedSkill(id string) error {
	for _, row := range c.State.Rows {
		if row.ID != id {
			continue
		}
		params := navigation.Destination(row)
		return c.Run(func() error {
			c.largeSkillID = id
			c.timedSelection = nil
			c.timedLapsInput = ""
			c.Push(navigation.SkillSwimmers, func(f *navigation.Frame) { f.Skill, f.Level, f.Version = id, params.Level, params.Version })
			return nil
		})
	}
	return nil
}

func (c *Controller) openLargeSelectedMode(mode string) error {
	if c.Frame().Screen != navigation.Skills {
		return nil
	}
	if mode == "single" {
		return nil
	}
	var chosen navigation.Row
	for _, row := range c.State.Rows {
		if row.ID == c.largeSkillID {
			chosen = row
			break
		}
	}
	if chosen.ID == "" {
		return nil
	}
	params := navigation.Destination(chosen)
	if err := c.Run(func() error {
		c.Push(navigation.SkillSwimmers, func(f *navigation.Frame) {
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

func (c *Controller) switchLargeAssessmentPerson(person string) {
	if person == c.Frame().Person {
		return
	}
	_ = c.Run(func() error {
		f := &c.Frames[len(c.Frames)-1]
		progress, err := c.Store.Progress(person, f.Skill, f.Version)
		if err != nil {
			return err
		}
		if progress.Complete {
			return nil
		}
		attempt, err := c.Store.NewAttempt(person, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		f.Person, f.Attempt, f.Page = person, attempt, 0
		return nil
	})
}

func (c *Controller) toggleLargeCorrectionSubskill(id string) error {
	if c.Frame().Screen != navigation.Assessment || !c.Frame().Review {
		return nil
	}
	status := "successful"
	for _, row := range c.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "not_yet"
			break
		}
	}
	return c.Run(func() error { return c.Store.Mark(c.Frame().Attempt, id, status) })
}

func (c *Controller) configureLargeTimedSetup(state *navigation.State, rows []navigation.Row) {
	state.Title = "Set Up Timed Heat"
	state.PageSize = max(1, len(rows))
	state.HomeRailAction = nil
	state.BackRailAction = discardRailAction(c.DiscardTimedSelection)
	state.Progress = c.timedHeatInput()
	state.ProgressInBody = true
	state.ComposeBody = c.largeView().ComposeLargeTimedSetup(*state, rows)
}

func (c *Controller) startLargeTimedHeat() error {
	laps, err := strconv.Atoi(c.timedLapsInput)
	if err != nil || laps < 1 || laps > 999 || len(c.timedSelection) == 0 {
		return nil
	}
	if c.Frame().Screen != navigation.SkillSwimmers && c.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		participants, err := c.Store.NewTimedHeat(c.timedSelection, f.Lesson, f.Skill, laps)
		if err != nil {
			return err
		}
		heat := &timedHeatState{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps}
		for _, person := range participants {
			heat.Participants = append(heat.Participants, timedParticipantState{participant: timedParticipant{
				AttemptID: person.AttemptID, PersonID: person.PersonID, Name: person.Name, LapsRequired: person.LapsRequired,
			}})
		}
		c.timedHeat, c.largeHeatPage = heat, 0
		c.Push(navigation.TimedHeat, nil)
		return nil
	})
}

func configureLargeAssessmentHeader(state *navigation.State, skill curriculum.Skill) {
	if ui.WrappedTextHeight(skill.Name, 450, 16, true) <= shell.SplitHeaderHeight {
		return
	}
	state.SplitHeader.Right = curriculum.CompactTitle(skill)
	state.PrimaryInfo = skill.Number + ". " + skill.Name
	state.PrimaryInfoFontSize = 16
}
