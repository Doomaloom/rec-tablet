package feedback

import (
	"fmt"

	"rectablet/internal/shared/ui"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"
)

type Controller struct {
	*navigation.Controller
	Store Repository
	large largeFeedback
}

type largeFeedback struct {
	person  domain.Enrollment
	entries []domain.Feedback
}

func New(c *navigation.Controller, s Repository) *Controller {
	return &Controller{Controller: c, Store: s}
}
func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	rows := []navigation.Row{}
	s := c.Store
	course, roster, e := navigation.LessonContext(s, f, &state)
	if e != nil {
		return state, e
	}
	state.PrimaryInfo = course.Name
	switch f.Screen {
	case navigation.FeedbackSwimmers:
		state.Title = "Feedback"
		selected := f.Person
		if state.DisplayMode.IsLarge() {
			found := false
			for _, p := range roster {
				if p.ID == selected {
					found = true
					break
				}
			}
			if !found && len(roster) > 0 {
				selected = roster[0].ID
				c.Frames[len(c.Frames)-1].Person = selected
			}
			state.PageSize = 6
		}
		for _, p := range roster {
			feedback, e := s.Feedback(p.ID)
			if e != nil {
				return state, e
			}
			unit := "skills"
			if len(feedback) == 1 {
				unit = "skill"
			}
			label := fmt.Sprintf("%s · %d %s in progress", p.Name, len(feedback), unit)
			if !p.Active {
				label += " · Former enrollment"
			}
			if state.DisplayMode.IsLarge() {
				label = p.Name
				if !p.Active {
					label += " · Former enrollment"
				}
			}
			rows = append(rows, navigation.Row{ID: p.ID, Label: label, Selected: state.DisplayMode.IsLarge() && p.ID == selected})
			if state.DisplayMode.IsLarge() && p.ID == selected {
				c.large = largeFeedback{person: p, entries: feedback}
			}
		}
		if state.DisplayMode.IsLarge() && len(roster) == 0 {
			c.large = largeFeedback{}
		}
		state.Empty = "No swimmers loaded."

	case navigation.Feedback:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		feedback, e := s.Feedback(f.Person)
		if e != nil {
			return state, e
		}
		for _, entry := range feedback {
			label := entry.Skill.Name + " · " + navigation.ProgressText(entry.Progress)
			if entry.Timed {
				label += " · Last heat " + formatTimedElapsed(entry.TimedElapsedMS)
			}
			rows = append(rows, navigation.Row{ID: entry.Skill.ID, Label: label, Data: navigation.Parameters{Version: entry.Version}})
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
			state.Progress = navigation.ProgressText(entry.Progress)
			navigation.AppendPrimaryInfo(&state, entry.Skill.Name)
			state.Empty = fmt.Sprintf("Latest attempt successful. Needs %d more successful observation(s).", 3-entry.Successes)
			if entry.Timed {
				rows = append(rows, navigation.Row{Label: fmt.Sprintf("Timed heat · %d/%d laps · %s", entry.TimedLaps, entry.TimedRequired, formatTimedElapsed(entry.TimedElapsedMS)), Disabled: true})
			}
			for _, subskill := range entry.Remaining {
				rows = append(rows, navigation.Row{Label: subskill.Text, Disabled: true})
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
		if c.DisplayMode == ui.DisplayLarge {
			c.Frames[len(c.Frames)-1].Person = id
			return nil
		}
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
