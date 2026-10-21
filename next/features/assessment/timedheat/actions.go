package timedheat

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

func (c *Component) ConfirmTimedSelection() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers || len(c.State.Selection) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		c.State.LapsInput = ""
		c.navigation.Push(navigation.TimedLaps, nil)
		return nil
	})
}

func (c *Component) DiscardTimedSelection() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.navigation.Run(func() error {
		c.State.Selection = nil
		c.State.LapsInput = ""
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) ConfirmTimedLaps() error {
	if c.navigation.Frame().Screen != navigation.TimedLaps || len(c.State.Selection) == 0 {
		return nil
	}
	laps := 0
	if _, e := fmt.Sscanf(c.State.LapsInput, "%d", &laps); e != nil || laps < 1 || laps > 999 {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		participants, e := c.storage.NewTimedHeat(c.State.Selection, f.Lesson, f.Skill, laps)
		if e != nil {
			return nil
		}
		heat := &Heat{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps, Stop: make(chan struct{})}
		for _, participant := range participants {
			heat.Participants = append(heat.Participants, ParticipantState{Participant: Participant{AttemptID: participant.AttemptID, PersonID: participant.PersonID, Name: participant.Name, LapsRequired: participant.LapsRequired}})
		}
		c.State.Heat = heat
		c.navigation.Push(navigation.TimedHeat, nil)
		return nil
	})
}

func (c *Component) ToggleTimedTimer() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.State.Heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		if c.State.Heat.Running {
			c.State.Heat.Elapsed = c.Elapsed()
			c.State.Heat.Running = false
			c.StopTicker()
			return nil
		}
		c.State.Heat.StartedAt = time.Now().Add(-c.State.Heat.Elapsed)
		c.State.Heat.Running = true
		c.StartTicker(c.State.Heat)
		return nil
	})
}

func (c *Component) ResetTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.State.Heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		wasRunning := c.State.Heat.Running
		if wasRunning {
			c.State.Heat.Elapsed = c.Elapsed()
		}
		c.StopTicker()
		ids := []string{}
		for i := range c.State.Heat.Participants {
			participant := &c.State.Heat.Participants[i]
			ids = append(ids, participant.Participant.AttemptID)
		}
		if e := c.storage.ResetTimedHeat(ids); e != nil {
			if wasRunning {
				c.State.Heat.StartedAt = time.Now().Add(-c.State.Heat.Elapsed)
				c.StartTicker(c.State.Heat)
			}
			return nil
		}
		c.State.Heat.Running = false
		c.State.Heat.Elapsed = 0
		for i := range c.State.Heat.Participants {
			participant := &c.State.Heat.Participants[i]
			participant.Participant.LapsCompleted = 0
			participant.Completed = false
			participant.CompletedAt = 0
		}
		return nil
	})
}

func (c *Component) RecordTimedLap(personID string) error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.State.Heat == nil || !c.State.Heat.Running {
		return nil
	}
	return c.navigation.Run(func() error {
		for i := range c.State.Heat.Participants {
			participant := &c.State.Heat.Participants[i]
			if participant.Participant.PersonID != personID || participant.Completed {
				continue
			}
			completed, e := c.storage.RecordTimedLap(participant.Participant.AttemptID)
			if e != nil {
				return nil
			}
			participant.Participant.LapsCompleted = completed
			if completed >= participant.Participant.LapsRequired {
				participant.Completed = true
				participant.CompletedAt = c.Elapsed()
			}
			break
		}
		return nil
	})
}

func (c *Component) SaveTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.State.Heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		benchmark, ok := common.BenchmarkFor(c.State.Heat.SkillID)
		if !ok {
			return nil
		}
		wasRunning := c.State.Heat.Running
		elapsed := c.Elapsed()
		c.State.Heat.Running = false
		c.State.Heat.Elapsed = elapsed
		c.StopTicker()
		results := make([]domain.TimedResult, 0, len(c.State.Heat.Participants))
		for _, participant := range c.State.Heat.Participants {
			participantElapsed := elapsed
			if participant.Completed {
				participantElapsed = participant.CompletedAt
			}
			results = append(results, domain.TimedResult{AttemptID: participant.Participant.AttemptID, LapsRequired: participant.Participant.LapsRequired, LapsCompleted: participant.Participant.LapsCompleted, ElapsedMS: participantElapsed.Milliseconds(), Success: participant.Completed && participantElapsed <= benchmark.Limit})
		}
		if e := c.storage.SaveTimedHeat(results); e != nil {
			if wasRunning {
				c.State.Heat.StartedAt = time.Now().Add(-elapsed)
				c.State.Heat.Running = true
				c.StartTicker(c.State.Heat)
			}
			return nil
		}
		c.State.Heat = nil
		c.State.Selection = nil
		c.State.LapsInput = ""
		c.navigation.Pop()
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) DiscardTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.State.Heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		ids := []string{}
		for _, participant := range c.State.Heat.Participants {
			ids = append(ids, participant.Participant.AttemptID)
		}
		if e := c.storage.DiscardTimedHeat(ids); e != nil {
			return nil
		}
		c.StopTicker()
		c.State.Heat = nil
		c.State.Selection = nil
		c.State.LapsInput = ""
		c.navigation.Pop()
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) StartHeat() error {
	laps, err := strconv.Atoi(c.State.LapsInput)
	if err != nil || laps < 1 || laps > 999 || len(c.State.Selection) == 0 {
		return nil
	}
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		participants, err := c.storage.NewTimedHeat(c.State.Selection, f.Lesson, f.Skill, laps)
		if err != nil {
			return err
		}
		heat := &Heat{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps}
		for _, person := range participants {
			heat.Participants = append(heat.Participants, ParticipantState{Participant: Participant{
				AttemptID: person.AttemptID, PersonID: person.PersonID, Name: person.Name, LapsRequired: person.LapsRequired,
			}})
		}
		c.State.Heat, c.State.HeatPage = heat, 0
		c.navigation.Push(navigation.TimedHeat, nil)
		return nil
	})
}

func (c *Component) SelectedSwimmers() []ui.Item {
	var swimmers []ui.Item
	for _, id := range c.State.Selection {
		student, err := c.storage.GetEnrollment(id)
		if err != nil {
			continue
		}
		swimmers = append(swimmers, ui.Item{ID: id, Label: student.Name, Selected: true, Fill: ui.Green})
	}
	return swimmers
}

func (c *Component) Select(id string) error {
	switch c.navigation.Frame().Screen {
	case navigation.SkillSwimmers:
		if common.Contains(c.State.Selection, id) {
			c.State.Selection = common.Remove(c.State.Selection, id)
		} else {
			c.State.Selection = append(c.State.Selection, id)
		}
		c.navigation.Saved("")
	case navigation.TimedLaps:
		if len(id) == len("timed-digit:")+1 && strings.HasPrefix(id, "timed-digit:") {
			c.State.LapsInput += strings.TrimPrefix(id, "timed-digit:")
			if len(c.State.LapsInput) > 3 {
				c.State.LapsInput = c.State.LapsInput[:3]
			}
		} else if id == "timed-clear" {
			c.State.LapsInput = ""
		}
	case navigation.TimedHeat:
		if strings.HasPrefix(id, "timed-lap:") {
			return c.RecordTimedLap(strings.TrimPrefix(id, "timed-lap:"))
		}
	}
	return nil
}
