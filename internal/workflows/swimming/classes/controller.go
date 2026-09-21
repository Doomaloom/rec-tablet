package classes

import (
	"rectablet/internal/workflows/swimming/navigation"

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
	switch f.Screen {
	case navigation.Classes:
		state.HeaderAction = "Date"
		instructor, e := s.Meta("instructor")
		if e != nil {
			return state, e
		}
		if instructor == "" {
			instructor = "Ready for lessons"
		}
		state.Subtitle = instructor + " · " + c.Date
		courses, e := s.ListCourses(c.Date)
		if e != nil {
			return state, e
		}
		for _, r := range courses {
			rows = append(rows, navigation.Row{ID: r.ID, Label: r.Name, Detail: r.Time + " · " + r.Location, Data: navigation.Parameters{Lesson: r.LessonID}})
		}
		dates, e := s.Dates()
		if e != nil {
			return state, e
		}
		state.Empty = "No classes on this date. Choose Date to view loaded lessons."
		if len(dates) == 0 {
			state.Empty = "No classes loaded. Hold My Classes for 3 seconds to import a class package."
		}

	case navigation.Dates:
		state.Title = "Lesson Date"
		dates, e := s.Dates()
		if e != nil {
			return state, e
		}
		for _, d := range dates {
			t, _ := time.Parse("2006-01-02", d)
			rows = append(rows, navigation.Row{ID: d, Label: t.Format("Monday, Jan 02, 2006"), Selected: d == c.Date})
		}
		state.Empty = "Import a class package to add lesson dates."

	case navigation.Class:
		_, _, e := navigation.LessonContext(s, f, c.Date, &state)
		if e != nil {
			return state, e
		}
		state.Title = "This Lesson"
		rows = []navigation.Row{{ID: "attendance", Label: "Attendance"}, {ID: "skills", Label: "Skills"}, {ID: "feedback_swimmers", Label: "Feedback"}}

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
	case navigation.Classes:
		c.Push(navigation.Class, func(n *navigation.Frame) { n.Course = id; n.Lesson = navigation.Destination(row).Lesson })

	case navigation.Dates:
		c.Date = id
		c.Reset()

	case navigation.Class:
		c.Push(navigation.Route(id), nil)

	}
	return nil
}
func (c *Controller) Header() error {
	f := c.Frame()
	switch f.Screen {
	case navigation.Classes:
		c.Push(navigation.Dates, nil)

	}
	return nil
}
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
