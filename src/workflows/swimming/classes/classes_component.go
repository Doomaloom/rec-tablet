package classes

import (
	"rectablet/src/workflows/swimming/classes/classlist"
	"rectablet/src/workflows/swimming/classes/datepicker"
	"rectablet/src/workflows/swimming/classes/lessonmenu"
	"rectablet/src/workflows/swimming/navigation"
)

// Component coordinates class browsing through explicitly owned child components.
type Component struct {
	navigation *navigation.Controller
	store      Repository
	classList  *classlist.Component
	datePicker *datepicker.Component
	lessonMenu *lessonmenu.Component
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	selectItem := func(id string) { _ = nav.Select(id) }
	c.classList = classlist.New(selectItem)
	c.lessonMenu = lessonmenu.New(selectItem)
	c.datePicker = datepicker.New(datepicker.Options{
		OnSelect:    selectItem,
		OnMoveMonth: func(delta int) { _ = c.moveCalendarMonth(delta) },
	})
	return c
}
func (c *Component) Build() (navigation.State, error) {
	f := c.navigation.Frame()
	state := c.navigation.InitialState()
	rows := []navigation.Row{}
	s := c.store
	switch f.Screen {
	case navigation.Classes:
		courses, e := s.ListCourses(c.navigation.Date)
		if e != nil {
			return state, e
		}
		dates, e := s.Dates()
		if e != nil {
			return state, e
		}
		rows = c.classList.Configure(&state, courses, dates, c.navigation.Date)
		for i := range rows {
			rows[i].Data = navigation.Parameters{Lesson: courses[i].LessonID}
		}

	case navigation.Dates:
		dates, e := s.Dates()
		if e != nil {
			return state, e
		}
		rows = c.datePicker.Configure(&state, dates, c.navigation.Date)

	case navigation.Class:
		course, _, e := navigation.LessonContext(s, f, &state)
		if e != nil {
			return state, e
		}
		rows = c.lessonMenu.Configure(&state, course)

	}
	c.navigation.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Component) Select(id string) error {
	f := c.navigation.Frame()
	var row navigation.Row
	for _, r := range c.navigation.State.Rows {
		if r.ID == id {
			row = r
			break
		}
	}
	switch f.Screen {
	case navigation.Classes:
		c.navigation.Push(navigation.Class, func(n *navigation.Frame) { n.Course = id; n.Lesson = navigation.Destination(row).Lesson })

	case navigation.Dates:
		c.navigation.Date = id
		c.datePicker.Reset()
		c.navigation.Reset()

	case navigation.Class:
		c.navigation.Push(navigation.Route(id), nil)

	}
	return nil
}

func (c *Component) moveCalendarMonth(delta int) error {
	return c.navigation.Run(func() error {
		c.datePicker.MoveMonth(delta, c.navigation.Date)
		return nil
	})
}
func (c *Component) Header() error {
	f := c.navigation.Frame()
	switch f.Screen {
	case navigation.Classes:
		c.datePicker.Reset()
		c.navigation.Push(navigation.Dates, nil)

	}
	return nil
}
func (c *Component) Mark(string) error { return nil }
func (c *Component) Submit() error     { return nil }
