package assessment

import (
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// EnterClassMode opens one pending attempt for every active student at this level.
// Existing drafts are reused, just as they are when entering Multi mode.
func (c *Component) EnterClassMode() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.MultiSelection && c.navigation.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		roster, err := c.store.ListRoster(f.Course)
		if err != nil {
			return err
		}
		var people []string
		for _, student := range roster {
			if student.Active && student.Level == f.Level {
				people = append(people, student.ID)
			}
		}
		if len(people) == 0 {
			return nil
		}
		attempts, err := c.store.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		c.leaveMultiSetup()
		c.multi.session, c.multi.edit = nil, nil
		c.class.attempts = attempts
		c.class.roster = append([]domain.MultiAttempt(nil), attempts...)
		c.class.status = make(map[string]string, len(attempts))
		c.navigation.Push(navigation.ClassAssessment, nil)
		return nil
	})
}

func (c *Component) classRailActions() []shell.RailAction {
	save, discard := *saveRailAction(c.SaveClassStudent), *discardRailAction(c.DiscardClassStudent)
	save.ID, save.Label = "save-student", "Save current student"
	discard.ID, discard.Label = "discard-student", "Discard current attempt"
	discardAll, saveAll := *discardRailAction(c.DiscardClass), *saveRailAction(c.SaveClass)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending attempts", shared.DiscardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending attempts", shared.SaveAllActionIcon
	return []shell.RailAction{save, discard, discardAll, saveAll}
}

func (c *Component) toggleClassSubskill(id string) error {
	index := c.navigation.Frame().Page
	if index < 0 || index >= len(c.class.attempts) {
		return nil
	}
	status := "successful"
	for _, row := range c.navigation.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
		}
	}
	if err := c.store.Mark(c.class.attempts[index].AttemptID, id, status); err != nil {
		return err
	}
	c.navigation.Saved("")
	return nil
}

func (c *Component) SaveClassStudent() error { return c.finishClassAttempts(false, true) }

func (c *Component) DiscardClassStudent() error { return c.finishClassAttempts(false, false) }

func (c *Component) SaveClass() error { return c.finishClassAttempts(true, true) }

func (c *Component) DiscardClass() error { return c.finishClassAttempts(true, false) }

// Only remove students from the queue after the complete storage operation succeeds.
func (c *Component) finishClassAttempts(all, save bool) error {
	if c.navigation.Frame().Screen != navigation.ClassAssessment || len(c.class.attempts) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		index := c.navigation.Frame().Page
		pending := c.class.attempts
		if !all {
			pending = pending[index : index+1]
		}
		ids := make([]string, len(pending))
		for i, student := range pending {
			ids[i] = student.AttemptID
		}
		var err error
		if save {
			err = c.store.SubmitAttempts(ids)
		} else {
			err = c.store.DiscardAttempts(ids)
		}
		if err != nil {
			return err
		}
		for _, student := range pending {
			if save {
				c.class.status[student.PersonID] = "Saved"
			} else {
				c.class.status[student.PersonID] = "Discarded"
			}
		}
		if all {
			c.class.attempts = nil
		} else {
			c.class.attempts = append(c.class.attempts[:index], c.class.attempts[index+1:]...)
		}
		if len(c.class.attempts) == 0 {
			c.navigation.Pop()
		} else {
			c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = min(index, len(c.class.attempts)-1)
		}
		c.navigation.Saved("")
		return nil
	})
}
