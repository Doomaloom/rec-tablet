package feedback

import (
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
	"rectablet/src/workflows/swimming/navigation"
)

// Component loads feedback and coordinates navigation through owned child components.
type Component struct {
	navigation    *navigation.Controller
	store         Repository
	swimmerPicker SwimmerPicker
	skillSummary  SkillSummary
	skillDetail   SkillDetail
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	c.swimmerPicker = SwimmerPicker{onSelect: func(id string) { _ = nav.Select(id) }}
	c.skillSummary = SkillSummary{
		onSelect:   func(id string) { _ = nav.Select(id) },
		onAttempts: c.openAttempts,
		detail:     &c.skillDetail,
	}
	return c
}

func (c *Component) Build() (navigation.State, error) {
	frame := c.navigation.Frame()
	state := c.navigation.InitialState()
	course, roster, err := navigation.LessonContext(c.store, frame, &state)
	if err != nil {
		return state, err
	}
	state.PrimaryInfo = course.Name
	rows := []navigation.Row{}
	switch frame.Screen {
	case navigation.FeedbackSwimmers:
		selected := frame.Person
		if state.DisplayMode.IsLarge() {
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
		summary := summaryData{}
		for _, person := range roster {
			feedback, err := c.store.Feedback(person.ID)
			if err != nil {
				return state, err
			}
			entries[person.ID] = feedback
			if person.ID == selected {
				summary = summaryData{person: person, entries: feedback}
			}
		}
		c.skillSummary.selected = summary
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
	if frame.Screen == navigation.FeedbackSwimmers && state.DisplayMode.IsLarge() {
		state.Rows = rows
		state.RowCount = len(rows)
		state.PageSize = 1
		state.PageCount = max(1, len(c.skillSummary.selected.entries))
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
		if c.navigation.DisplayMode == ui.DisplayLarge {
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
