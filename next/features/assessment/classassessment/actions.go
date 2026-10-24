package classassessment

import (
	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

func (c *Component) Enter() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers && c.navigation.Frame().Screen != navigation.MultiSelection && c.navigation.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.enter(c.navigation.Frame())
}

// OpenSkill enters class marking directly, keeping the skill picker as the back destination.
func (c *Component) OpenSkill(skill navigation.Row) error {
	f := c.navigation.Frame()
	if (f.Screen != navigation.Skills && f.Screen != navigation.LessonPlanDetail) || skill.Disabled {
		return nil
	}
	params := navigation.Destination(skill)
	f.Skill, f.Level, f.Version = skill.ID, params.Level, params.Version
	return c.enter(f)
}

func (c *Component) enter(f navigation.Frame) error {
	return c.navigation.Run(func() error {
		roster, err := c.storage.ListRoster(f.Course)
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
		attempts, err := c.storage.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		if c.actions.BeforeEnter != nil {
			c.actions.BeforeEnter()
		}
		c.State.Attempts = attempts
		c.State.Roster = append([]domain.MultiAttempt(nil), attempts...)
		c.State.Status = make(map[string]string, len(attempts))
		c.navigation.Push(navigation.ClassAssessment, func(frame *navigation.Frame) {
			frame.Skill, frame.Level, frame.Version = f.Skill, f.Level, f.Version
		})
		return nil
	})
}

func (c *Component) classRailActions() []contracts.RailAction {
	confirm, cancel := *common.SaveRailAction(c.SaveClass), *common.DiscardRailAction(c.DiscardClass)
	confirm.ID, confirm.Label = "confirm-class", "Confirm class"
	cancel.ID, cancel.Label = "cancel-class", "Cancel class"
	return []contracts.RailAction{confirm, cancel}
}

func (c *Component) railActions() []contracts.RailAction {
	save, discard := *common.SaveRailAction(c.SaveClassStudent), *common.DiscardRailAction(c.DiscardClassStudent)
	save.ID, save.Label = "save-student", "Save current student"
	discard.ID, discard.Label = "discard-student", "Discard current attempt"
	discardAll, saveAll := *common.DiscardRailAction(c.DiscardClass), *common.SaveRailAction(c.SaveClass)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending attempts", common.DiscardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending attempts", common.SaveAllActionIcon
	return []contracts.RailAction{save, discard, discardAll, saveAll}
}

func (c *Component) ToggleSubskill(id string) error {
	index := c.navigation.Frame().Page
	if index < 0 || index >= len(c.State.Attempts) {
		return nil
	}
	status := "successful"
	for _, row := range c.navigation.Rows() {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
		}
	}
	if err := c.storage.Mark(c.State.Attempts[index].AttemptID, id, status); err != nil {
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
	if c.navigation.Frame().Screen != navigation.ClassAssessment || len(c.State.Attempts) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		index := c.navigation.Frame().Page
		pending := c.State.Attempts
		if !all {
			pending = pending[index : index+1]
		}
		ids := make([]string, len(pending))
		for i, student := range pending {
			ids[i] = student.AttemptID
		}
		var err error
		if save {
			err = c.storage.SubmitAttempts(ids)
		} else {
			err = c.storage.DiscardAttempts(ids)
		}
		if err != nil {
			return err
		}
		for _, student := range pending {
			if save {
				c.State.Status[student.PersonID] = "Saved"
			} else {
				c.State.Status[student.PersonID] = "Discarded"
			}
		}
		if all {
			c.State.Attempts = nil
		} else {
			c.State.Attempts = append(c.State.Attempts[:index], c.State.Attempts[index+1:]...)
		}
		if len(c.State.Attempts) == 0 {
			c.navigation.Pop()
		} else {
			c.navigation.SetPage(min(index, len(c.State.Attempts)-1))
		}
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) SelectPerson(id string) {
	_ = c.navigation.Run(func() error {
		for i, student := range c.State.Attempts {
			if student.PersonID == id {
				c.navigation.SetPage(i)
				break
			}
		}
		return nil
	})
}
