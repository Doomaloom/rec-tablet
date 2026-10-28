package feedback

import (
	"rectablet/next/features/feedback/skilldetail"
	"rectablet/next/features/feedback/skillsummary"
	"rectablet/next/features/feedback/swimmerpicker"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

// Component loads feedback and coordinates navigation through owned child components.
type Component struct {
	navigation    *navigation.Controller
	store         Repository
	combined      bool
	complete      bool
	swimmerPicker swimmerpicker.SwimmerPicker
	skillSummary  skillsummary.SkillSummary
	skillDetail   skilldetail.SkillDetail
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	c.swimmerPicker = swimmerpicker.SwimmerPicker{OnSelect: func(id string) { _ = nav.Select(id) }}
	c.skillSummary = skillsummary.SkillSummary{
		OnSelect:   func(id string) { _ = nav.Select(id) },
		OnAttempts: c.openAttempts,
		Detail:     c.skillDetail.Text,
		OnFilter:   c.setFilter,
	}
	return c
}

func (c *Component) Build() (navigation.State, error) {
	frame := c.navigation.Frame()
	state := c.navigation.InitialState()
	c.configureLayout(state.DisplayMode)
	course, roster, err := navigation.LessonContext(c.store, frame, &state)
	if err != nil {
		return state, err
	}
	state.PrimaryInfo = course.Name
	rows := []navigation.Row{}
	switch frame.Screen {
	case navigation.FeedbackSwimmers:
		selected := frame.Person
		if c.combined {
			found := false
			for _, person := range roster {
				if person.ID == selected {
					found = true
					break
				}
			}
			if !found && len(roster) > 0 {
				selected = roster[0].ID
				c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Person = selected
			}
		}
		entries := make(map[string][]domain.Feedback, len(roster))
		summary := skillsummary.Snapshot{}
		for _, person := range roster {
			feedback, err := c.store.Feedback(person.ID)
			if c.combined && c.complete {
				feedback, err = c.store.CompletedFeedback(person.ID)
			}
			if err != nil {
				return state, err
			}
			entries[person.ID] = feedback
			if person.ID == selected {
				summary = skillsummary.Snapshot{Person: person, Entries: feedback}
			}
		}
		c.skillSummary.Selected = summary
		c.skillSummary.Complete = c.complete
		rows = c.swimmerPicker.Configure(&state, roster, entries, selected)
	case navigation.Feedback, navigation.FeedbackDetail:
		person, err := c.store.GetEnrollment(frame.Person)
		if err != nil {
			return state, err
		}
		entries, err := c.store.Feedback(frame.Person)
		if err != nil {
			return state, err
		}
		if frame.Screen == navigation.Feedback {
			rows = c.skillSummary.Configure(&state, person, entries)
			for i := range rows {
				rows[i].Data = navigation.Parameters{Version: entries[i].Version}
			}
		} else {
			rows = c.skillDetail.Configure(&state, person, entries, frame.Skill, frame.Version)
		}
	}
	if frame.Screen == navigation.FeedbackSwimmers && c.combined {
		state.AlwaysShowPagePosition = true
		state.HeaderStartFraction = feedbackSidebarFraction
		state.Rows = rows
		state.RowCount = len(rows)
		state.PageSize = 2
		state.PageCount = max(1, (len(c.skillSummary.Selected.Entries)+1)/2)
		state.Page = min(max(0, frame.Page), state.PageCount-1)
		c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = state.Page
	} else {
		c.navigation.Paginate(&state, rows)
	}
	c.composeFeatureBody(&state)
	return state, nil
}

func (c *Component) Select(id string) error {
	frame := c.navigation.Frame()
	switch frame.Screen {
	case navigation.FeedbackSwimmers:
		if c.combined {
			c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Person = id
			c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = 0
			return nil
		}
		c.navigation.Push(navigation.Feedback, func(next *navigation.Frame) { next.Person = id })
	case navigation.Feedback:
		for _, row := range c.navigation.State.Rows {
			if row.ID == id {
				c.navigation.Push(navigation.FeedbackDetail, func(next *navigation.Frame) { next.Skill = id; next.Version = navigation.Destination(row).Version })
				break
			}
		}
	}
	return nil
}

func (c *Component) openAttempts(person, skill, version string) {
	_ = c.navigation.Run(func() error {
		c.navigation.Push(navigation.Attempts, func(frame *navigation.Frame) {
			frame.Person, frame.Skill, frame.Version = person, skill, version
		})
		return nil
	})
}

func (c *Component) Header() error {
	if c.navigation.Frame().Screen == navigation.FeedbackDetail {
		c.navigation.Push(navigation.Attempts, nil)
	}
	return nil
}
func (c *Component) Mark(string) error { return nil }
func (c *Component) Submit() error     { return nil }

func (c *Component) setFilter(complete bool) {
	_ = c.navigation.Run(func() error {
		c.complete = complete
		c.navigation.History.Frames[len(c.navigation.History.Frames)-1].Page = 0
		return nil
	})
}
