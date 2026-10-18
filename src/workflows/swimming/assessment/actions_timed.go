package assessment

import (
	"fmt"
	"time"

	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

func (c *Component) ConfirmTimedSelection() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers || len(c.timed.selection) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		c.timed.lapsInput = ""
		c.navigation.Push(navigation.TimedLaps, nil)
		return nil
	})
}

func (c *Component) DiscardTimedSelection() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.TimedLaps {
		return nil
	}
	return c.navigation.Run(func() error {
		c.timed.selection = nil
		c.timed.lapsInput = ""
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) ConfirmTimedLaps() error {
	if c.navigation.Frame().Screen != navigation.TimedLaps || len(c.timed.selection) == 0 {
		return nil
	}
	laps := 0
	if _, e := fmt.Sscanf(c.timed.lapsInput, "%d", &laps); e != nil || laps < 1 || laps > 999 {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		participants, e := c.store.NewTimedHeat(c.timed.selection, f.Lesson, f.Skill, laps)
		if e != nil {
			return nil
		}
		heat := &timedHeatState{SkillID: f.Skill, Version: f.Version, Lesson: f.Lesson, LapsRequired: laps, stop: make(chan struct{})}
		for _, participant := range participants {
			heat.Participants = append(heat.Participants, timedParticipantState{participant: timedParticipant{AttemptID: participant.AttemptID, PersonID: participant.PersonID, Name: participant.Name, LapsRequired: participant.LapsRequired}})
		}
		c.timed.heat = heat
		c.navigation.Push(navigation.TimedHeat, nil)
		return nil
	})
}

func (c *Component) ToggleTimedTimer() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.timed.heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		if c.timed.heat.Running {
			c.timed.heat.Elapsed = c.timed.timedHeatElapsed()
			c.timed.heat.Running = false
			c.timed.stopTimedTicker()
			return nil
		}
		c.timed.heat.StartedAt = time.Now().Add(-c.timed.heat.Elapsed)
		c.timed.heat.Running = true
		c.timed.startTimedTicker(c.timed.heat)
		return nil
	})
}

func (c *Component) ResetTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.timed.heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		wasRunning := c.timed.heat.Running
		if wasRunning {
			c.timed.heat.Elapsed = c.timed.timedHeatElapsed()
		}
		c.timed.stopTimedTicker()
		ids := []string{}
		for i := range c.timed.heat.Participants {
			participant := &c.timed.heat.Participants[i]
			ids = append(ids, participant.participant.AttemptID)
		}
		if e := c.store.ResetTimedHeat(ids); e != nil {
			if wasRunning {
				c.timed.heat.StartedAt = time.Now().Add(-c.timed.heat.Elapsed)
				c.timed.startTimedTicker(c.timed.heat)
			}
			return nil
		}
		c.timed.heat.Running = false
		c.timed.heat.Elapsed = 0
		for i := range c.timed.heat.Participants {
			participant := &c.timed.heat.Participants[i]
			participant.participant.LapsCompleted = 0
			participant.completed = false
			participant.completedAt = 0
		}
		return nil
	})
}

func (c *Component) RecordTimedLap(personID string) error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.timed.heat == nil || !c.timed.heat.Running {
		return nil
	}
	return c.navigation.Run(func() error {
		for i := range c.timed.heat.Participants {
			participant := &c.timed.heat.Participants[i]
			if participant.participant.PersonID != personID || participant.completed {
				continue
			}
			completed, e := c.store.RecordTimedLap(participant.participant.AttemptID)
			if e != nil {
				return nil
			}
			participant.participant.LapsCompleted = completed
			if completed >= participant.participant.LapsRequired {
				participant.completed = true
				participant.completedAt = c.timed.timedHeatElapsed()
			}
			break
		}
		return nil
	})
}

func (c *Component) SaveTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.timed.heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		benchmark, ok := benchmarkFor(c.timed.heat.SkillID)
		if !ok {
			return nil
		}
		wasRunning := c.timed.heat.Running
		elapsed := c.timed.timedHeatElapsed()
		c.timed.heat.Running = false
		c.timed.heat.Elapsed = elapsed
		c.timed.stopTimedTicker()
		results := make([]domain.TimedResult, 0, len(c.timed.heat.Participants))
		for _, participant := range c.timed.heat.Participants {
			participantElapsed := elapsed
			if participant.completed {
				participantElapsed = participant.completedAt
			}
			results = append(results, domain.TimedResult{AttemptID: participant.participant.AttemptID, LapsRequired: participant.participant.LapsRequired, LapsCompleted: participant.participant.LapsCompleted, ElapsedMS: participantElapsed.Milliseconds(), Success: participant.completed && participantElapsed <= benchmark.Limit})
		}
		if e := c.store.SaveTimedHeat(results); e != nil {
			if wasRunning {
				c.timed.heat.StartedAt = time.Now().Add(-elapsed)
				c.timed.heat.Running = true
				c.timed.startTimedTicker(c.timed.heat)
			}
			return nil
		}
		c.timed.heat = nil
		c.timed.selection = nil
		c.timed.lapsInput = ""
		c.navigation.Pop()
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) DiscardTimedHeat() error {
	if c.navigation.Frame().Screen != navigation.TimedHeat || c.timed.heat == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		ids := []string{}
		for _, participant := range c.timed.heat.Participants {
			ids = append(ids, participant.participant.AttemptID)
		}
		if e := c.store.DiscardTimedHeat(ids); e != nil {
			return nil
		}
		c.timed.stopTimedTicker()
		c.timed.heat = nil
		c.timed.selection = nil
		c.timed.lapsInput = ""
		c.navigation.Pop()
		c.navigation.Pop()
		return nil
	})
}
