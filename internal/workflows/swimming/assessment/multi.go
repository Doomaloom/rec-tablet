package assessment

import (
	"fmt"

	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
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

type multiStudentSnapshot struct {
	PersonID, Name, Status string
	Passed, Total          int
}

func (c *Controller) EnterMultiMode() error {
	if c.Frame().Screen != navigation.SkillSwimmers {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		skill, err := c.Store.Skill(f.Version, f.Skill)
		if err != nil {
			return err
		}
		if !allowsMultiMode(skill) {
			return nil
		}
		roster, err := c.Store.ListRoster(f.Course)
		if err != nil {
			return err
		}
		attendance, err := c.Store.Attendance(f.Lesson)
		if err != nil {
			return err
		}
		course, err := c.Store.GetCourse(f.Course)
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
		c.multi, c.multiEdit = session, nil
		c.Push(navigation.MultiGroupCount, nil)
		return nil
	})
}

func (c *Controller) leaveMultiSetup() {
	for c.Frame().Screen == navigation.MultiSelection || c.Frame().Screen == navigation.MultiGroupCount {
		c.Pop()
	}
}

func (c *Controller) EnterSingleMode() error {
	if c.Frame().Screen != navigation.MultiSelection && c.Frame().Screen != navigation.MultiGroupCount {
		return nil
	}
	return c.Run(func() error {
		c.leaveMultiSetup()
		c.multi, c.multiEdit = nil, nil
		return nil
	})
}

func (c *Controller) ConfirmMultiSelection() error {
	if c.Frame().Screen != navigation.MultiSelection || !c.multiReady() {
		return nil
	}
	return c.Run(func() error {
		people := make([]string, 0, len(c.multi.Roster))
		for _, group := range c.multi.Groups {
			people = append(people, group.Members...)
		}
		f := c.Frame()
		attempts, err := c.Store.OpenMultiAttempts(people, f.Lesson, f.Skill)
		if err != nil {
			return err
		}
		byPerson := make(map[string]domain.MultiAttempt, len(attempts))
		for _, attempt := range attempts {
			byPerson[attempt.PersonID] = attempt
		}
		for i := range c.multi.Groups {
			group := &c.multi.Groups[i]
			group.Attempts = nil
			// Keep each group's display order consistent with the roster, not tap order.
			for _, student := range c.multi.Roster {
				if containsString(group.Members, student.ID) {
					group.Attempts = append(group.Attempts, byPerson[student.ID])
				}
			}
		}
		c.leaveMultiSetup()
		c.Push(navigation.MultiAssessment, nil)
		return nil
	})
}

func (c *Controller) buildMultiAssessmentState(f navigation.Frame) (navigation.State, error) {
	state := c.InitialState()
	if c.multi == nil || len(c.multi.Groups) == 0 {
		return state, fmt.Errorf("no pending groups")
	}
	state.PageCount = len(c.multi.Groups)
	state.Page = max(0, min(f.Page, state.PageCount-1))
	c.Frames[len(c.Frames)-1].Page = state.Page
	group, skill := c.multi.Groups[state.Page], c.multi.Skill
	label := fmt.Sprintf("Group %d of %d", group.Number, c.multi.GroupCount)
	state.Title = label
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: label}
	state.PrimaryInfo = fmt.Sprintf("%s · %d students", c.multi.CourseInfo, len(group.Attempts))
	state.HideFooter = true
	state.LeftRailActions = c.multiRailActions()
	students := make([]multiStudentSnapshot, 0, len(group.Attempts))
	for _, attempt := range group.Attempts {
		answers, err := c.Store.Draft(attempt.AttemptID)
		if err != nil {
			return state, err
		}
		student := multiStudentSnapshot{PersonID: attempt.PersonID, Name: attempt.Name, Status: "unmarked", Total: len(skill.Subskills)}
		for _, subskill := range skill.Subskills {
			if answers[subskill.ID] == "successful" {
				student.Passed++
			}
		}
		if len(answers) > 0 {
			student.Status = "not_yet"
		}
		if student.Total > 0 && student.Passed == student.Total {
			student.Status = "successful"
		}
		students = append(students, student)
		state.Rows = append(state.Rows, navigation.Row{ID: student.PersonID, Label: student.Name, Status: student.Status})
	}
	state.RowCount, state.PageSize = len(students), len(students)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.composeLargeMultiAssessmentBody(students)
	} else {
		state.ComposeBody = c.composeMultiAssessmentBody(students)
	}
	return state, nil
}

func (c *Controller) multiRailActions() []shell.RailAction {
	save, discard := *saveRailAction(c.SaveMultiGroup), *discardRailAction(c.DiscardMultiGroup)
	save.ID, save.Label = "save-group", "Save group"
	discard.ID, discard.Label = "discard-group", "Discard group"
	discardAll, saveAll := *discardRailAction(c.DiscardMultiAssessment), *saveRailAction(c.SaveMultiAssessment)
	discardAll.ID, discardAll.Label, discardAll.Icon = "discard-all", "Discard all pending groups", discardAllActionIcon
	saveAll.ID, saveAll.Label, saveAll.Icon = "save-all", "Save all pending groups", saveAllActionIcon
	return []shell.RailAction{save, discard, discardAll, saveAll}
}

func (c *Controller) currentMultiStudent(person string) (domain.MultiAttempt, bool) {
	if c.Frame().Screen != navigation.MultiAssessment || c.multi == nil {
		return domain.MultiAttempt{}, false
	}
	page := c.Frame().Page
	if page < 0 || page >= len(c.multi.Groups) {
		return domain.MultiAttempt{}, false
	}
	for _, attempt := range c.multi.Groups[page].Attempts {
		if attempt.PersonID == person {
			return attempt, true
		}
	}
	return domain.MultiAttempt{}, false
}

func (c *Controller) PassMultiStudent(person string) error {
	attempt, ok := c.currentMultiStudent(person)
	if !ok {
		return nil
	}
	return c.Run(func() error {
		answers := make(map[string]string, len(c.multi.Skill.Subskills))
		for _, subskill := range c.multi.Skill.Subskills {
			answers[subskill.ID] = "successful"
		}
		if err := c.Store.ReplaceDraftAnswers(attempt.AttemptID, answers); err != nil {
			return err
		}
		c.Saved("")
		return nil
	})
}

func (c *Controller) EditMultiStudent(person string) error {
	attempt, ok := c.currentMultiStudent(person)
	if !ok {
		return nil
	}
	return c.Run(func() error {
		answers, err := c.Store.Draft(attempt.AttemptID)
		if err != nil {
			return err
		}
		c.multiEdit = &multiStudentEdit{Attempt: attempt, Answers: answers}
		c.Push(navigation.MultiStudent, func(f *navigation.Frame) { f.Person, f.Attempt = person, attempt.AttemptID })
		return nil
	})
}

func (c *Controller) buildMultiStudentState(f navigation.Frame) (navigation.State, error) {
	state := c.InitialState()
	if c.multi == nil || c.multiEdit == nil || c.multiEdit.Attempt.AttemptID != f.Attempt {
		return state, fmt.Errorf("no student selected")
	}
	skill, edit := c.multi.Skill, c.multiEdit
	state.Title = edit.Attempt.Name
	state.SplitHeader = &shell.SplitHeader{Left: skill.Number + ". " + curriculum.CompactTitle(skill), Right: edit.Attempt.Name}
	state.PrimaryInfo = c.multi.CourseInfo + " · " + skill.Number + ". " + skill.Name
	state.HideFooter = true
	state.HomeRailAction = saveRailAction(c.SaveMultiStudent)
	state.BackRailAction = discardRailAction(c.DiscardMultiStudent)
	// Checklist edits stay local until Save; leaving via Escape also abandons them.
	state.OnLeave = func() { c.multiEdit = nil }
	for _, subskill := range skill.Subskills {
		state.Rows = append(state.Rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Status: edit.Answers[subskill.ID]})
	}
	state.RowCount, state.PageSize = len(state.Rows), len(state.Rows)
	if state.DisplayMode.IsLarge() {
		state.ComposeBody = c.composeLargeChecklistBody(state.Rows, func(id string) { _ = c.Controller.Select(id) })
	} else {
		state.ComposeBody = c.composeClassChecklist(state.Rows)
	}
	return state, nil
}

func (c *Controller) toggleMultiSubskill(id string) error {
	if c.multiEdit == nil {
		return nil
	}
	for _, subskill := range c.multi.Skill.Subskills {
		if subskill.ID == id {
			if c.multiEdit.Answers[id] == "successful" {
				delete(c.multiEdit.Answers, id)
			} else {
				c.multiEdit.Answers[id] = "successful"
			}
			break
		}
	}
	return nil
}

func (c *Controller) SaveMultiStudent() error {
	if c.Frame().Screen != navigation.MultiStudent || c.multiEdit == nil {
		return nil
	}
	return c.Run(func() error {
		answers := make(map[string]string, len(c.multi.Skill.Subskills))
		for _, subskill := range c.multi.Skill.Subskills {
			answers[subskill.ID] = "not_yet"
			if c.multiEdit.Answers[subskill.ID] == "successful" {
				answers[subskill.ID] = "successful"
			}
		}
		if err := c.Store.ReplaceDraftAnswers(c.multiEdit.Attempt.AttemptID, answers); err != nil {
			return err
		}
		c.Pop()
		c.Saved("")
		return nil
	})
}

func (c *Controller) DiscardMultiStudent() error {
	if c.Frame().Screen != navigation.MultiStudent {
		return nil
	}
	return c.Run(func() error { c.Pop(); return nil })
}

func (c *Controller) SaveMultiGroup() error         { return c.finishMultiGroups(false, true) }
func (c *Controller) DiscardMultiGroup() error      { return c.finishMultiGroups(false, false) }
func (c *Controller) SaveMultiAssessment() error    { return c.finishMultiGroups(true, true) }
func (c *Controller) DiscardMultiAssessment() error { return c.finishMultiGroups(true, false) }

func (c *Controller) finishMultiGroups(all, save bool) error {
	if c.Frame().Screen != navigation.MultiAssessment || c.multi == nil || len(c.multi.Groups) == 0 {
		return nil
	}
	return c.Run(func() error {
		index := c.Frame().Page
		groups := c.multi.Groups
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
			err = c.Store.SubmitAttempts(ids)
		} else {
			err = c.Store.DiscardMultiAttempts(attempts)
		}
		if err != nil {
			return err
		}
		if all {
			c.multi.Groups = nil
		} else {
			c.multi.Groups = append(c.multi.Groups[:index], c.multi.Groups[index+1:]...)
		}
		if len(c.multi.Groups) == 0 {
			c.multi, c.multiEdit = nil, nil
			c.Pop()
		} else {
			c.Frames[len(c.Frames)-1].Page = min(index, len(c.multi.Groups)-1)
		}
		c.Saved("")
		return nil
	})
}
