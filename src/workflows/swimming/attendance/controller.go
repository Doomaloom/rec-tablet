package attendance

import (
	"rectablet/src/workflows/swimming/navigation"
	"strings"
	"time"
)

type Controller struct {
	*navigation.Controller
	Store Repository
}

func New(c *navigation.Controller, s Repository) *Controller {
	return &Controller{Controller: c, Store: s}
}
func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	if c.DisplayMode.IsLarge() {
		state.PageSize = 12
	}
	rows := []navigation.Row{}
	s := c.Store
	_, roster, e := navigation.LessonContext(s, f, &state)
	if e != nil {
		return state, e
	}
	switch f.Screen {
	case navigation.Attendance:
		state.Title = "Attendance"
		// Attendance finalization is best effort on the device. A failed
		// background write leaves the roster usable without an error popup.
		_ = s.MarkPastAttendanceAbsent(f.Lesson, time.Now().Format("2006-01-02"))
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
			label := p.Name + " · " + strings.ToUpper(status[:1]) + status[1:]
			rows = append(rows, navigation.Row{ID: p.ID, Label: label, Status: status})
		}
		state.Empty = "No enrolled swimmers."
	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Controller) Select(id string) error {
	f := c.Frame()
	switch f.Screen {
	case navigation.Attendance:
		status := "present"
		found := false
		for _, row := range c.State.Rows {
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
		if e := c.Store.SetAttendance(id, f.Lesson, status); e != nil {
			return nil
		}
		c.Saved("")

	}
	return nil
}
func (c *Controller) Header() error     { return nil }
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
