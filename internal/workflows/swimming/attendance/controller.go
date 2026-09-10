package attendance

import (
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
	case navigation.Attendance:
		state.Title = "Attendance"
		marks, e := s.Attendance(f.Lesson)
		if e != nil {
			return state, e
		}
		for _, p := range roster {
			if !p.Active {
				continue
			}
			status := marks[p.ID]
			if status == "" {
				status = "unmarked"
			}
			rows = append(rows, navigation.Row{ID: p.ID, Label: p.Name, Detail: strings.ToUpper(status[:1]) + status[1:], Status: status})
		}
		state.Empty = "No enrolled swimmers."

	case navigation.AttendanceMark:
		p, e := s.GetEnrollment(f.Person)
		if e != nil {
			return state, e
		}
		state.Title = p.Name
		state.Subtitle = "Attendance · " + state.Subtitle
		rows = []navigation.Row{{ID: "present", Label: "Present"}, {ID: "absent", Label: "Absent"}, {ID: "unmarked", Label: "Unmarked"}}

	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Controller) Select(id string) error {
	f := c.Frame()
	switch f.Screen {
	case navigation.Attendance:
		c.Push(navigation.AttendanceMark, func(n *navigation.Frame) { n.Person = id })

	case navigation.AttendanceMark:
		// Attendance is an immediate touch action. A failed write leaves the
		// selection screen unchanged and is intentionally quiet.
		if e := c.Store.SetAttendance(f.Person, f.Lesson, id); e != nil {
			return nil
		}
		c.Saved("")
		c.Pop()

	}
	return nil
}
func (c *Controller) Header() error     { return nil }
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
