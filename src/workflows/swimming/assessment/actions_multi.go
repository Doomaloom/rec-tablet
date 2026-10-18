package assessment

import (
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/assessment/view/shared"
	"rectablet/src/workflows/swimming/curriculum"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

type multiGroup struct {
	Number, Capacity int
	Members          []string
	Attempts         []domain.MultiAttempt
}

type multiSession struct {
	Roster             []domain.Enrollment
	Skill              curriculum.Skill
	CourseInfo         string
	Groups             []multiGroup
	AssignmentGroup    int
	GroupCount         int
	ProposedGroupCount int
}

type multiStudentEdit struct {
	Attempt domain.MultiAttempt
	Answers map[string]string
}

type multiStudentSnapshot = shared.MultiStudentSnapshot

func (c *Component) EnterMultiMode() error {
	if c.navigation.Frame().Screen != navigation.SkillSwimmers {
		return nil
	}
	return c.navigation.Run(func() error {
		f := c.navigation.Frame()
		skill, err := c.store.Skill(f.Version, f.Skill)
		if err != nil {
			return err
		}
		if !allowsMultiMode(skill) {
			return nil
		}
		roster, err := c.store.ListRoster(f.Course)
		if err != nil {
			return err
		}
		attendance, err := c.store.Attendance(f.Lesson)
		if err != nil {
			return err
		}
		course, err := c.store.GetCourse(f.Course)
		if err != nil {
			return err
		}
		session := &multiSession{Skill: skill, CourseInfo: course.Name + " · " + course.Time}
		for _, student := range roster {
			if student.Active && student.Level == f.Level && attendance[student.ID] == "present" {
				session.Roster = append(session.Roster, student)
			}
		}
		if len(session.Roster) > 0 {
			session.ProposedGroupCount = 1
		}
		c.multi.session, c.multi.edit = session, nil
		c.navigation.Push(navigation.MultiGroupCount, nil)
		return nil
	})
}

func (c *Component) leaveMultiSetup() {
	for c.navigation.Frame().Screen == navigation.MultiSelection || c.navigation.Frame().Screen == navigation.MultiGroupCount {
		c.navigation.Pop()
	}
}

func (c *Component) EnterSingleMode() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection && c.navigation.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.navigation.Run(func() error {
		c.leaveMultiSetup()
		c.multi.session, c.multi.edit = nil, nil
		return nil
	})
}

func (c *Component) ConfirmMultiSelection() error {
	if c.navigation.Frame().Screen != navigation.MultiSelection || !c.multi.multiReady() {
		return nil
	}
	return c.navigation.Run(func() error {
		people := make([]string, 0, len(c.multi.session.Roster))
		for _, group := range c.multi.session.Groups {
			people = append(people, group.Members...)
		}
		f := c.navigation.Frame()
		attempts, err := c.store.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		byPerson := make(map[string]domain.MultiAttempt, len(attempts))
		for _, attempt := range attempts {
			byPerson[attempt.PersonID] = attempt
		}
		for i := range c.multi.session.Groups {
			group := &c.multi.session.Groups[i]
			group.Attempts = nil
			// Keep each group's display order consistent with the roster, not tap order.
			for _, student := range c.multi.session.Roster {
				if containsString(group.Members, student.ID) {
					group.Attempts = append(group.Attempts, byPerson[student.ID])
				}
			}
		}
		c.leaveMultiSetup()
		c.navigation.Push(navigation.MultiAssessment, nil)
		return nil
	})
}

func (c *Component) multiRailActions() []shell.RailAction {
	save, discard := *saveRailAction(c.SaveMultiGroup), *discardRailAction(c.DiscardMultiGroup)
	save.ID, save.Label = "save-group", "Save group"
	discard.ID, discard.Label = "discard-group", "Discard group"
	discardAll, saveAll := *discardRailAction(c.DiscardMultiAssessment), *saveRailAction(c.SaveMultiAssessment)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending groups", shared.DiscardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending groups", shared.SaveAllActionIcon
	return []shell.RailAction{save, discard, discardAll, saveAll}
}

func (c *Component) currentMultiStudent(person string) (domain.MultiAttempt, bool) {
	if c.navigation.Frame().Screen != navigation.MultiAssessment || c.multi.session == nil {
		return domain.MultiAttempt{}, false
	}
	page := c.navigation.Frame().Page
	if page < 0 || page >= len(c.multi.session.Groups) {
		return domain.MultiAttempt{}, false
	}
	for _, attempt := range c.multi.session.Groups[page].Attempts {
		if attempt.PersonID == person {
			return attempt, true
		}
	}
	return domain.MultiAttempt{}, false
}

func (c *Component) PassMultiStudent(person string) error {
	attempt, ok := c.currentMultiStudent(person)
	if !ok {
		return nil
	}
	return c.navigation.Run(func() error {
		answers := make(map[string]string, len(c.multi.session.Skill.Subskills))
		for _, subskill := range c.multi.session.Skill.Subskills {
			answers[subskill.ID] = "successful"
		}
		if err := c.store.ReplaceDraftAnswers(attempt.AttemptID, answers); err != nil {
			return err
		}
		c.navigation.Saved("")
		return nil
	})
}

func (c *Component) EditMultiStudent(person string) error {
	attempt, ok := c.currentMultiStudent(person)
	if !ok {
		return nil
	}
	return c.navigation.Run(func() error {
		answers, err := c.store.Draft(attempt.AttemptID)
		if err != nil {
			return err
		}
		c.multi.edit = &multiStudentEdit{Attempt: attempt, Answers: answers}
		c.navigation.Push(navigation.MultiStudent, func(f *navigation.Frame) { f.Person, f.Attempt = person, attempt.AttemptID })
		return nil
	})
}

func (c *Component) SaveMultiStudent() error {
	if c.navigation.Frame().Screen != navigation.MultiStudent || c.multi.edit == nil {
		return nil
	}
	return c.navigation.Run(func() error {
		answers := make(map[string]string, len(c.multi.session.Skill.Subskills))
		for _, subskill := range c.multi.session.Skill.Subskills {
			answers[subskill.ID] = "not_yet"
			if c.multi.edit.Answers[subskill.ID] == "successful" {
				answers[subskill.ID] = "successful"
			}
		}
		if err := c.store.ReplaceDraftAnswers(c.multi.edit.Attempt.AttemptID, answers); err != nil {
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
	if c.navigation.Frame().Screen != navigation.MultiAssessment || c.multi.session == nil || len(c.multi.session.Groups) == 0 {
		return nil
	}
	return c.navigation.Run(func() error {
		index := c.navigation.Frame().Page
		groups := c.multi.session.Groups
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
			err = c.store.SubmitAttempts(ids)
		} else {
			err = c.store.DiscardMultiAttempts(attempts)
		}
		if err != nil {
			return err
		}
		if all {
			c.multi.session.Groups = nil
		} else {
			c.multi.session.Groups = append(c.multi.session.Groups[:index], c.multi.session.Groups[index+1:]...)
		}
		if len(c.multi.session.Groups) == 0 {
			c.multi.session, c.multi.edit = nil, nil
			c.navigation.Pop()
		} else {
			c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = min(index, len(c.multi.session.Groups)-1)
		}
		c.navigation.Saved("")
		return nil
	})
}
