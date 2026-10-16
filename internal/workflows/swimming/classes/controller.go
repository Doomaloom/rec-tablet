package classes

import (
	"fmt"

	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/navigation"

	"time"
)

type Controller struct {
	*navigation.Controller
	Store         Repository
	calendarMonth time.Time
}

func New(c *navigation.Controller, s Repository) *Controller {
	return &Controller{Controller: c, Store: s}
}
func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	if c.DisplayMode.IsLarge() {
		switch f.Screen {
		case navigation.Classes:
			state.PageSize = 6
		case navigation.Class:
			state.PageSize = 3
		}
	}
	rows := []navigation.Row{}
	s := c.Store
	switch f.Screen {
	case navigation.Classes:
		state.HeaderAction = "Date"
		courses, e := s.ListCourses(c.Date)
		if e != nil {
			return state, e
		}
		for _, r := range courses {
			rows = append(rows, navigation.Row{ID: r.ID, Label: r.Name + " · " + r.Time, Data: navigation.Parameters{Lesson: r.LessonID}})
		}
		dates, e := s.Dates()
		if e != nil {
			return state, e
		}
		for i, d := range dates {
			if d == c.Date {
				state.SplitHeader = &shell.SplitHeader{Left: state.Title, Right: fmt.Sprintf("Class %d/%d", i+1, len(dates))}
				break
			}
		}
		state.Empty = "No classes scheduled. Choose Date to view other lessons."
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
		if c.DisplayMode.IsLarge() {
			if c.calendarMonth.IsZero() {
				selected, parseErr := time.Parse("2006-01-02", c.Date)
				if parseErr != nil {
					selected = time.Now()
				}
				c.calendarMonth = time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
			}
		}
		state.Empty = "Import a class package to add lesson dates."

	case navigation.Class:
		course, _, e := navigation.LessonContext(s, f, &state)
		if e != nil {
			return state, e
		}
		level := course.Level
		if level == "" {
			level = course.Name
		}
		state.SplitHeader = &shell.SplitHeader{Left: level, Right: course.Time}
		state.PrimaryInfo = ""
		rows = []navigation.Row{{ID: "attendance", Label: "Attendance"}, {ID: "skills", Label: "Skills"}, {ID: "feedback_swimmers", Label: "Feedback"}}

	}
	if c.DisplayMode.IsLarge() && f.Screen == navigation.Dates {
		// Calendar month navigation needs the complete set of scheduled dates.
		state.PageSize = max(1, len(rows))
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
		c.calendarMonth = time.Time{}
		c.Reset()

	case navigation.Class:
		c.Push(navigation.Route(id), nil)

	}
	return nil
}

func (c *Controller) moveCalendarMonth(delta int) error {
	return c.Controller.Run(func() error {
		if c.calendarMonth.IsZero() {
			selected, err := time.Parse("2006-01-02", c.Date)
			if err != nil {
				selected = time.Now()
			}
			c.calendarMonth = time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
		}
		c.calendarMonth = c.calendarMonth.AddDate(0, delta, 0)
		return nil
	})
}
func (c *Controller) Header() error {
	f := c.Frame()
	switch f.Screen {
	case navigation.Classes:
		c.calendarMonth = time.Time{}
		c.Push(navigation.Dates, nil)

	}
	return nil
}
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
