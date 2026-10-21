package multiassessment

import (
	"fmt"

	"rectablet/next/contracts"
	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

func (c *Component) Enter() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		skill, err := c.storage.Skill(f.Version, f.Skill)
		if err != nil {
			return err
		}
		if !common.AllowsMultiMode(skill) {
			return nil
		}
		roster, err := c.storage.ListRoster(f.Course)
		if err != nil {
			return err
		}
		attendance, err := c.storage.Attendance(f.Lesson)
		if err != nil {
			return err
		}
		course, err := c.storage.GetCourse(f.Course)
		if err != nil {
			return err
		}
		session := &Session{Skill: skill, CourseInfo: course.Name + " · " + course.Time}
		for _, student := range roster {
			if student.Active && student.Level == f.Level && attendance[student.ID] == "present" {
				session.Roster = append(session.Roster, student)
			}
		}
		if len(session.Roster) > 0 {
			session.ProposedGroupCount = 1
		}
		c.State.Session, c.State.Edit = session, nil
		c.navigation.Push(navigation.MultiGroupCount, nil)
		return nil
	})
}

func (c *Component) LeaveSetup() {
	for c.navigation.Frame().Screen == navigation.MultiSelection || c.navigation.Frame().Screen == navigation.MultiGroupCount {
		c.navigation.Pop()
	}
}

func (c *Component) CancelSetup() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection && c.navigation.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.navigation.Run(func() error {
		c.LeaveSetup()
		c.State.Session, c.State.Edit = nil, nil
		return nil
	})
}

func (c *Component) ConfirmSelection() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection || !c.Ready() {
		return nil
	}
	return c.navigation.Run(func() error {
		people := make([]string, 0, len(c.State.Session.Roster))
		for _, group := range c.State.Session.Groups {
			people = append(people, group.Members...)
		}
		f := c.navigation.Frame()
		attempts, err := c.storage.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		byPerson := make(map[string]domain.MultiAttempt, len(attempts))
		for _, attempt := range attempts {
			byPerson[attempt.PersonID] = attempt
		}
		for i := range c.State.Session.Groups {
			group := &c.State.Session.Groups[i]
			group.Attempts = nil
			// Keep each group's display order consistent with the roster, not tap order.
			for _, student := range c.State.Session.Roster {
				if common.Contains(group.Members, student.ID) {
					group.Attempts = append(group.Attempts, byPerson[student.ID])
				}
			}
		}
		c.LeaveSetup()
		c.navigation.Push(navigation.MultiAssessment, nil)
		return nil
	})
}

func (c *Component) railActions() []contracts.RailAction {
	save, discard := *common.SaveRailAction(c.SaveMultiGroup), *common.DiscardRailAction(c.DiscardMultiGroup)
	save.ID, save.Label = "save-group", "Save group"
	discard.ID, discard.Label = "discard-group", "Discard group"
	discardAll, saveAll := *common.DiscardRailAction(c.DiscardMultiAssessment), *common.SaveRailAction(c.SaveMultiAssessment)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending groups", common.DiscardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending groups", common.SaveAllActionIcon
	return []contracts.RailAction{save, discard, discardAll, saveAll}
}

func (c *Component) currentStudent(person string) (domain.MultiAttempt, bool) {
	if c.navigation.Frame().Screen != navigation.MultiAssessment || c.State.Session == nil {
		return domain.MultiAttempt{}, false
	}
	page := c.navigation.Frame().Page
	if page < 0 || page >= len(c.State.Session.Groups) {
		return domain.MultiAttempt{}, false
	}
	for _, attempt := range c.State.Session.Groups[page].Attempts {
		if attempt.PersonID == person {
			return attempt, true
		}
	}
	return domain.MultiAttempt{}, false
}

func (c *Component) PassMultiStudent(person string) error {
	attempt, ok := c.currentStudent(person)
	if !ok {
		return nil
	}
	return c.navigation.Run(func() error {
		answers := make(map[string]string, len(c.State.Session.Skill.Subskills))
		for _, subskill := range c.State.Session.Skill.Subskills {
			answers[subskill.ID] = "successful"
		}
		if err := c.storage.ReplaceDraftAnswers(attempt.AttemptID, answers); err != nil {
			return err
		}
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) EditMultiStudent(person string) error {
	attempt, ok := c.currentStudent(person)
	if !ok {
		return nil
	}
	return c.navigation.Run(func() error {
		answers, err := c.storage.Draft(attempt.AttemptID)
		if err != nil {
			return err
		}
		c.State.Edit = &StudentEdit{Attempt: attempt, Answers: answers}
		c.navigation.Push(navigation.MultiStudent, func(f *navigation.Frame) { f.Person, f.Attempt = person, attempt.AttemptID })
		return nil
	})
}

func (c *Component) SaveMultiStudent() error {
	if c.navigation.Frame().Screen != navigation.MultiStudent || c.State.Edit == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		answers := make(map[string]string, len(c.State.Session.Skill.Subskills))
		for _, subskill := range c.State.Session.Skill.Subskills {
			answers[subskill.ID] = "not_yet"
			if c.State.Edit.Answers[subskill.ID] == "successful" {
				answers[subskill.ID] = "successful"
			}
		}
		if err := c.storage.ReplaceDraftAnswers(c.State.Edit.Attempt.AttemptID, answers); err != nil {
			return err
		}
		c.navigation.Pop()
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) DiscardMultiStudent() error {
	if c.navigation.Frame().Screen != navigation.MultiStudent {
		return nil
	}
	return c.navigation.Run(func() error { c.navigation.Pop(); return nil })
}

func (c *Component) SaveMultiGroup() error { return c.finishMultiGroups(false, true) }

func (c *Component) DiscardMultiGroup() error { return c.finishMultiGroups(false, false) }

func (c *Component) SaveMultiAssessment() error { return c.finishMultiGroups(true, true) }

func (c *Component) DiscardMultiAssessment() error { return c.finishMultiGroups(true, false) }

func (c *Component) finishMultiGroups(all, save bool) error {
	if c.navigation.Frame().Screen != navigation.MultiAssessment || c.State.Session == nil || len(c.State.Session.Groups) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		index := c.navigation.Frame().Page
		groups := c.State.Session.Groups
		if !all {
			groups = groups[index : index+1]
		}
		var attempts []domain.MultiAttempt
		var ids []string
		for _, group := range groups {
			attempts = append(attempts, group.Attempts...)
			for _, attempt := range group.Attempts {
				ids = append(ids, attempt.AttemptID)
			}
		}
		var err error
		if save {
			err = c.storage.SubmitAttempts(ids)
		} else {
			err = c.storage.DiscardMultiAttempts(attempts)
		}
		if err != nil {
			return err
		}
		if all {
			c.State.Session.Groups = nil
		} else {
			c.State.Session.Groups = append(c.State.Session.Groups[:index], c.State.Session.Groups[index+1:]...)
		}
		if len(c.State.Session.Groups) == 0 {
			c.State.Session, c.State.Edit = nil, nil
			c.navigation.Pop()
		} else {
			c.navigation.SetPage(min(index, len(c.State.Session.Groups)-1))
		}
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) AdjustGroupCount(direction int) error {
	if c.navigation.Frame().Screen != navigation.MultiGroupCount || c.State.Session == nil || len(c.State.Session.Roster) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		c.State.Session.ProposedGroupCount = max(1, min(c.State.Session.ProposedGroupCount+direction, len(c.State.Session.Roster)))
		return nil
	})
}

func (c *Component) ChooseGroupCount() error {
	if c.navigation.Frame().Screen != navigation.MultiGroupCount || c.State.Session == nil {
		return nil
	}
	count := c.State.Session.ProposedGroupCount
	sizes := BalancedSizes(len(c.State.Session.Roster), count)
	if len(sizes) == 0 {
		return nil
	}
	if count != c.State.Session.GroupCount {
		c.State.Session.Groups = make([]Group, count)
		for i, size := range sizes {
			c.State.Session.Groups[i] = Group{Number: i + 1, Capacity: size}
		}
		c.State.Session.GroupCount = count
	}
	c.State.Session.AssignmentGroup = 0
	c.navigation.Push(navigation.MultiSelection, nil)
	return nil
}

func (c *Component) ChangeGroupCount() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection {
		return nil
	}
	return c.navigation.Run(func() error {
		c.State.Session.ProposedGroupCount = c.State.Session.GroupCount
		c.navigation.Pop()
		return nil
	})
}

func (c *Component) PageGroup(direction int) error {
	if c.navigation.Frame().Screen != navigation.MultiSelection || c.State.Session == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		c.State.Session.AssignmentGroup = max(0, min(c.State.Session.AssignmentGroup+direction, len(c.State.Session.Groups)-1))
		c.navigation.SetPage(0)
		return nil
	})
}

func (c *Component) SelectGroup(id string) {
	var index int
	if _, err := fmt.Sscan(id, &index); err != nil {
		return
	}
	_ = c.navigation.Run(func() error { c.State.Session.AssignmentGroup = index; c.State.MemberPage = 0; return nil })
}
