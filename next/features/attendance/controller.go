package attendance

import (
	"time"

	"rectablet/next/features/attendance/roster"
	"rectablet/next/navigation"
)

// Component owns attendance persistence and has a roster child for presentation.
type Component struct {
	navigation                    *navigation.Controller
	store                         Repository
	roster                        *roster.Component
	finalizedLesson, finalizedDay string
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	return &Component{
		navigation: nav,
		store:      store,
		roster:     roster.New(func(id string) { _ = nav.Select(id) }),
	}
}
func (c *Component) Build() (navigation.State, error) {
	f := c.navigation.Frame()
	state := c.navigation.InitialState()
	c.configureLayout(state.DisplayMode)
	rows := []navigation.Row{}
	s := c.store
	_, roster, e := navigation.LessonContext(s, f, &state)
	if e != nil {
		return state, e
	}
	switch f.Screen {
	case navigation.Attendance:
		// Attendance finalization is best effort on the device. A failed
		// background write leaves the roster usable without an error popup.
		today := time.Now().Format("2006-01-02")
		if c.finalizedLesson != f.Lesson || c.finalizedDay != today {
			if err := s.MarkPastAttendanceAbsent(f.Lesson, today); err == nil {
				c.finalizedLesson, c.finalizedDay = f.Lesson, today
			}
		}
		marks, e := s.Attendance(f.Lesson)
		if e != nil {
			return state, e
		}
		rows = c.roster.Configure(&state, roster, marks)
	}
	c.navigation.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Component) Select(id string) error {
	f := c.navigation.Frame()
	switch f.Screen {
	case navigation.Attendance:
		status := "present"
		found := false
		for _, row := range c.navigation.State.Rows {
			if row.ID != id {
				continue
			}
			found = true
			if row.Status == "present" {
				status = "absent"
			}
			break
		}
		if !found {
			return nil
		}
		// Attendance is an immediate toggle. A failed write leaves the current
		// row unchanged and is intentionally quiet.
		if e := c.store.SetAttendance(id, f.Lesson, status); e != nil {
			return nil
		}
		c.navigation.Saved("")

	}
	return nil
}
func (c *Component) Header() error     { return nil }
func (c *Component) Mark(string) error { return nil }
func (c *Component) Submit() error     { return nil }
