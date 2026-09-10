package assessment

import (
	"fmt"
	"rectablet/internal/workflows/swimming/navigation"
	"strings"
	"time"
)

type Controller struct {
	*navigation.Controller
	Store Repository
}

func New(c *navigation.Controller, s Repository) *Controller { return &Controller{c, s} }

func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	rows := []navigation.Row{}
	s := c.Store
	course, roster, e := navigation.LessonContext(s, f, c.Date, &state)
	if e != nil {
		return state, e
	}
	version := course.CurriculumVersion

	switch f.Screen {
	case navigation.Skills:
		state.Title = "Choose a Skill"
		catalog, e := s.CatalogFor(version)
		if e != nil {
			return state, e
		}
		seenLevels := map[string]bool{}
		for _, p := range roster {
			if !p.Active || seenLevels[p.Level] {
				continue
			}
			seenLevels[p.Level] = true
			level, ok := catalog.Level(p.Level)
			if !ok || len(level.Skills) == 0 {
				rows = append(rows, navigation.Row{ID: p.Level, Label: p.Level, Detail: "Missing curriculum · update swimmer level", Disabled: true})
				continue
			}
			for _, skill := range level.Skills {
				rows = append(rows, navigation.Row{ID: skill.ID, Label: skill.Number + ". " + skill.Name, Detail: level.Name, Data: navigation.Parameters{Level: p.Level, Version: version}})
			}
		}
		state.Empty = "No enrolled swimmers with an assigned level."

	case navigation.SkillSwimmers:
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = "Choose a Swimmer"
		state.Subtitle = skill.Name
		state.Context += "\n" + skill.Name
		for _, p := range roster {
			if !p.Active || p.Level != f.Level {
				continue
			}
			progress, e := s.Progress(p.ID, skill.ID, version)
			if e != nil {
				return state, e
			}
			row := navigation.Row{
				ID:               p.ID,
				Label:            p.Name,
				ProgressSegments: min(3, max(0, progress.Successes)),
				Data:             navigation.Parameters{Skill: skill.ID, Level: p.Level, Version: version},
			}
			if progress.Successes > 0 {
				person, skillID, curriculumVersion := p.ID, skill.ID, version
				row.OnHold = func() { _ = c.RemoveLatestSuccess(person, skillID, curriculumVersion) }
				row.HoldDuration = 2 * time.Second
			}
			rows = append(rows, row)
		}
		state.Empty = "No enrolled swimmers are assigned to this skill's level."

	case navigation.Attempts:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		if f.Version != "" {
			version = f.Version
		}
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = skill.Name
		state.Context += "\n" + state.Title + "\n" + skill.Name
		progress, e := s.Progress(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		state.Progress = navigation.ProgressText(progress)
		attempts, e := s.ListAttempts(f.Person, skill.ID, version)
		if e != nil {
			return state, e
		}
		for _, a := range attempts {
			label := "Not yet successful"
			if a.Success {
				label = "Successful"
			}
			if a.SubmittedAt == nil {
				label = "Unfinished attempt"
			} else if a.Editing {
				label += " · Edit saved"
			}
			created := a.CreatedAt
			clock := created
			if len(created) >= 19 {
				clock = created[11:19]
			}
			rows = append(rows, navigation.Row{ID: a.ID, Label: label, Detail: a.Date + " · " + clock + " UTC · Tap to review / correct"})
		}
		state.Empty = "No attempts recorded for this skill."

	case navigation.Assessment:
		a, e := s.GetAttempt(f.Attempt)
		if e != nil {
			return state, e
		}
		p, e := s.GetEnrollment(a.EnrollmentID)
		if e != nil {
			return state, e
		}
		skill, e := s.Skill(a.CurriculumVersion, a.SkillID)
		if e != nil {
			return state, e
		}
		answers, e := s.Draft(f.Attempt)
		if e != nil {
			return state, e
		}
		lesson, e := s.GetLesson(a.LessonID)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = skill.Name
		state.Context = course.Name + " · " + lesson.Date + "\n" + state.Title + "\n" + skill.Name
		for _, subskill := range skill.Subskills {
			status := answers[subskill.ID]
			if status == "" {
				status = "unmarked"
			}
			rows = append(rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Detail: strings.Join(subskill.Group, " › "), Status: status})
		}
		if f.Review {
			state.FooterAction = "Save correction"
			state.CanSubmit = len(answers) == len(rows)
			state.PageSize = 1
			state.Progress = fmt.Sprintf("%d/%d assessed · %s", len(answers), len(rows), lesson.Date)
		} else {
			state.PageSize = 3
			state.Progress = fmt.Sprintf("%d/%d complete · %s", successfulAnswers(answers), len(rows), lesson.Date)
		}
	}

	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}

func successfulAnswers(answers map[string]string) int {
	count := 0
	for _, status := range answers {
		if status == "successful" {
			count++
		}
	}
	return count
}

func (c *Controller) Select(id string) error {
	f := c.Frame()
	var row navigation.Row
	for _, r := range c.State.Rows {
		if r.ID == id {
			row = r
			break
		}
	}
	switch f.Screen {
	case navigation.Skills:
		params := navigation.Destination(row)
		c.Push(navigation.SkillSwimmers, func(n *navigation.Frame) {
			n.Skill = id
			n.Level = params.Level
			n.Version = params.Version
		})

	case navigation.SkillSwimmers:
		// Skill taps are best-effort on the device. Failed reads or writes leave
		// the current list intact without surfacing an error state.
		params := navigation.Destination(row)
		version := params.Version
		if version == "" {
			version = f.Version
		}
		skillID := f.Skill
		if params.Skill != "" {
			skillID = params.Skill
		}
		progress, e := c.Store.Progress(id, skillID, version)
		if e != nil {
			return nil
		}
		if progress.Complete {
			c.Saved("")
			return nil
		}
		attempt, e := c.Store.NewAttempt(id, f.Lesson, skillID)
		if e != nil {
			return nil
		}
		c.Push(navigation.Assessment, func(n *navigation.Frame) {
			n.Attempt = attempt
			n.Person = id
			n.Skill = skillID
			n.Level = f.Level
			n.Version = version
		})

	case navigation.Attempts:
		if e := c.Store.EditAttempt(id); e != nil {
			return nil
		}
		c.Push(navigation.Assessment, func(n *navigation.Frame) { n.Attempt = id; n.Review = true })

	case navigation.Assessment:
		return c.toggleSubskill(id)
	}
	return nil
}

func (c *Controller) toggleSubskill(id string) error {
	f := c.Frame()
	status := "successful"
	for _, row := range c.State.Rows {
		if row.ID == id && row.Status == "successful" {
			status = "unmarked"
			break
		}
	}
	if e := c.Store.Mark(f.Attempt, id, status); e != nil {
		// Keep the checklist usable if a local write fails; the persisted draft
		// remains unchanged and the next refresh shows its previous state.
		return nil
	}
	c.Saved("")
	if !f.Review {
		complete, e := c.assessmentComplete(f.Attempt)
		if e != nil {
			return nil
		}
		if complete {
			success, e := c.Store.Submit(f.Attempt)
			if e != nil {
				return nil
			}
			if !success {
				return nil
			}
			c.Pop()
			return nil
		}
	}
	return nil
}

func (c *Controller) assessmentComplete(id string) (bool, error) {
	attempt, e := c.Store.GetAttempt(id)
	if e != nil {
		return false, e
	}
	skill, e := c.Store.Skill(attempt.CurriculumVersion, attempt.SkillID)
	if e != nil {
		return false, e
	}
	answers, e := c.Store.Draft(id)
	if e != nil {
		return false, e
	}
	if len(answers) != len(skill.Subskills) {
		return false, nil
	}
	for _, subskill := range skill.Subskills {
		if answers[subskill.ID] != "successful" {
			return false, nil
		}
	}
	return true, nil
}

func (c *Controller) RemoveLatestSuccess(person, skill, version string) error {
	return c.Run(func() error {
		if e := c.Store.RemoveLatestSuccess(person, skill, version); e != nil {
			return nil
		}
		c.Saved("")
		return nil
	})
}

func (c *Controller) Header() error { return nil }

func (c *Controller) Mark(status string) error {
	if c.Frame().Screen != navigation.Assessment || len(c.State.Rows) == 0 {
		return nil
	}
	if e := c.Store.Mark(c.Frame().Attempt, c.State.Rows[0].ID, status); e != nil {
		return nil
	}
	c.Saved("")
	return nil
}

func (c *Controller) Submit() error {
	if c.Frame().Screen != navigation.Assessment {
		return nil
	}
	_, e := c.Store.Submit(c.Frame().Attempt)
	if e != nil {
		return nil
	}
	c.Saved("")
	c.Pop()
	return nil
}
