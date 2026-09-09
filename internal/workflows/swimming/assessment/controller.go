package assessment

import (
	"fmt"

	"rectablet/internal/workflows/swimming/navigation"
	"strings"
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
	_ = roster
	version := course.CurriculumVersion
	_ = version
	switch f.Screen {
	case navigation.Skills:
		state.Title = "Choose a Skill"
		catalog, e := s.CatalogFor(version)
		if e != nil {
			return state, e
		}
		seen := map[string]bool{}
		for _, p := range roster {
			levelID := p.Level
			if !p.Active || seen[levelID] {
				continue
			}
			seen[levelID] = true
			level, ok := catalog.Level(levelID)
			if !ok || len(level.Skills) == 0 {
				rows = append(rows, navigation.Row{Label: levelID, Detail: "Missing curriculum · update swimmer level", Disabled: true})
				continue
			}
			for _, skill := range level.Skills {
				rows = append(rows, navigation.Row{ID: skill.ID, Label: skill.Number + ". " + skill.Name, Detail: level.Name, Data: navigation.Parameters{Level: levelID}})
			}
		}
		state.Empty = "No enrolled swimmers with an assigned level."

	case navigation.SkillSwimmers:
		state.Title = "Choose a Swimmer"
		skill, e := s.Skill(version, f.Skill)
		if e != nil {
			return state, e
		}
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
			rows = append(rows, navigation.Row{ID: p.ID, Label: p.Name, Detail: navigation.ProgressText(progress)})
		}

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
		if p.Active && p.Level == strings.Split(skill.ID, ":")[0] && version == course.CurriculumVersion {
			state.HeaderAction = "New attempt"
		}
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
		state.Empty = "No attempts yet. Choose New attempt to assess every subskill."

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
		l, e := s.GetLesson(a.LessonID)
		if e != nil {
			return state, e
		}
		d := l.Date
		state.Title = p.Name
		state.Subtitle = skill.Name
		state.Context = course.Name + " · " + d + "\n" + state.Title + "\n" + skill.Name
		for _, subskill := range skill.Subskills {
			status := answers[subskill.ID]
			if status == "" {
				status = "unmarked"
			}
			rows = append(rows, navigation.Row{ID: subskill.ID, Label: subskill.Text, Detail: strings.Join(subskill.Group, " › "), Status: status})
		}
		state.FooterAction = "Submit attempt"
		if a.SubmittedAt != nil {
			state.FooterAction = "Save correction"
		}
		state.CanSubmit = len(answers) == len(rows)
		state.Progress = fmt.Sprintf("%d/%d assessed · %s", len(answers), len(rows), d)

	}
	if f.Screen == navigation.Assessment {
		state.PageSize = 1
	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
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
		c.Push(navigation.SkillSwimmers, func(n *navigation.Frame) { n.Skill = id; n.Level = navigation.Destination(row).Level })

	case navigation.SkillSwimmers:
		c.Push(navigation.Attempts, func(n *navigation.Frame) { n.Person = id })

	case navigation.Attempts:
		if e := c.Store.EditAttempt(id); e != nil {
			return e
		}
		c.Push(navigation.Assessment, func(n *navigation.Frame) { n.Attempt = id })

	}
	return nil
}
func (c *Controller) Header() error {
	f := c.Frame()
	switch f.Screen {
	case navigation.Attempts:
		if c.State.HeaderAction != "" {
			id, e := c.Store.NewAttempt(f.Person, f.Lesson, f.Skill)
			if e != nil {
				return e
			}
			c.Push(navigation.Assessment, func(n *navigation.Frame) { n.Attempt = id })
		}

	}
	return nil
}
func (c *Controller) Mark(status string) error {

	if c.Frame().Screen != navigation.Assessment || len(c.State.Rows) == 0 {
		return nil
	}
	if e := c.Store.Mark(c.Frame().Attempt, c.State.Rows[0].ID, status); e != nil {
		return e
	}
	c.Saved("Draft saved · Submit when all subskills are assessed")
	return nil
}
func (c *Controller) Submit() error {

	if c.Frame().Screen != navigation.Assessment {
		return nil
	}
	success, e := c.Store.Submit(c.Frame().Attempt)
	if e != nil {
		return e
	}
	message := "Attempt saved · Not yet successful"
	if success {
		message = "Attempt saved · Successful"
	}
	c.Saved(message)
	c.Pop()
	return nil
}
