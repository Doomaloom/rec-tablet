package feedback

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
	case navigation.FeedbackSwimmers:
		state.Title = "Feedback"
		for _, p := range roster {
			feedback, e := s.Feedback(p.ID)
			if e != nil {
				return state, e
			}
			detail := fmt.Sprintf("%d skills in progress", len(feedback))
			if !p.Active {
				detail += " · Former enrollment"
			}
			rows = append(rows, navigation.Row{ID: p.ID, Label: p.Name, Detail: detail})
		}
		state.Empty = "No swimmers loaded."

	case navigation.Feedback:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = "Feedback · All assessed skills in this course"
		feedback, e := s.Feedback(f.Person)
		if e != nil {
			return state, e
		}
		for _, entry := range feedback {
			detail := navigation.ProgressText(entry.Progress)
			if entry.Timed {
				detail += " · Last heat " + formatTimedElapsed(entry.TimedElapsedMS)
			}
			rows = append(rows, navigation.Row{ID: entry.Skill.ID, Label: entry.Skill.Name, Detail: detail, Data: navigation.Parameters{Version: entry.Version}})
		}
		state.Empty = "No assessed skills remain incomplete. Unassessed skills are not included here."

	case navigation.FeedbackDetail:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.HeaderAction = "Attempts"
		state.Empty = "This skill is now complete."
		feedback, e := s.Feedback(f.Person)
		if e != nil {
			return state, e
		}
		for _, entry := range feedback {
			if entry.Skill.ID != f.Skill || entry.Version != f.Version {
				continue
			}
			state.Subtitle = entry.Skill.Name
			state.Context += "\n" + state.Title + "\n" + entry.Skill.Name
			state.Progress = navigation.ProgressText(entry.Progress)
			state.Empty = fmt.Sprintf("Latest attempt successful. Needs %d more successful observation(s).", 3-entry.Successes)
			if entry.Timed {
				rows = append(rows, navigation.Row{Label: "Timed heat", Detail: fmt.Sprintf("%d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatTimedElapsed(entry.TimedElapsedMS)), Disabled: true})
			}
			for _, subskill := range entry.Remaining {
				rows = append(rows, navigation.Row{Label: subskill.Text, Detail: strings.Join(subskill.Group, " › "), Disabled: true})
			}
		}

	}
	if f.Screen == navigation.FeedbackDetail {
		state.PageSize = 1
	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}

func formatTimedElapsed(milliseconds int64) string {
	if milliseconds < 0 {
		milliseconds = 0
	}
	minutes := milliseconds / 60000
	seconds := (milliseconds / 1000) % 60
	tenths := (milliseconds % 1000) / 100
	return fmt.Sprintf("%02d:%02d.%d", minutes, seconds, tenths)
}
func (c *Controller) Select(id string) error {
	var row navigation.Row
	for _, r := range c.State.Rows {
		if r.ID == id {
			row = r
			break
		}
	}
	f := c.Frame()
	switch f.Screen {
	case navigation.FeedbackSwimmers:
		c.Push(navigation.Feedback, func(n *navigation.Frame) { n.Person = id })

	case navigation.Feedback:
		c.Push(navigation.FeedbackDetail, func(n *navigation.Frame) { n.Skill = id; n.Version = navigation.Destination(row).Version })
	}
	return nil
}
func (c *Controller) Header() error {
	f := c.Frame()
	switch f.Screen {
	case navigation.FeedbackDetail:
		c.Push(navigation.Attempts, nil)
	}
	return nil
}
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
