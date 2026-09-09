// Package navigation defines swimming routes and coordinates feature controllers.
package navigation

import (
	"fmt"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/domain"
	"time"
)

type Route string

const (
	Classes          Route = "classes"
	Dates            Route = "dates"
	Class            Route = "class"
	Attendance       Route = "attendance"
	AttendanceMark   Route = "attendance_mark"
	Skills           Route = "skills"
	SkillSwimmers    Route = "skill_swimmers"
	Attempts         Route = "attempts"
	Assessment       Route = "assessment"
	FeedbackSwimmers Route = "feedback_swimmers"
	Feedback         Route = "feedback"
	FeedbackDetail   Route = "feedback_detail"
	Maintenance      Route = "maintenance"
)

type Parameters struct{ Course, Lesson, Person, Skill, Level, Version, Attempt string }
type Frame struct {
	Screen Route
	Page   int
	Parameters
}
type Row = shell.Row
type State = shell.State

func Destination(row Row) Parameters { p, _ := row.Data.(Parameters); return p }

type Feature interface {
	Build() (State, error)
	Select(string) error
	Header() error
	Mark(string) error
	Submit() error
}

type Controller struct {
	*shell.History[Frame]
	Date     string
	State    State
	OnChange func()
	Features map[Route]Feature
}

func New() *Controller {
	return &Controller{History: shell.NewHistory(Frame{Screen: Classes}), Date: time.Now().Format("2006-01-02"), Features: map[Route]Feature{}}
}
func (c *Controller) Snapshot() shell.State { return c.State }
func (c *Controller) Subscribe(fn func())   { c.OnChange = fn }
func (c *Controller) Frame() Frame          { return c.Frames[len(c.Frames)-1] }
func (c *Controller) Push(route Route, change func(*Frame)) {
	f := c.Frame()
	f.Screen = route
	f.Page = 0
	if change != nil {
		change(&f)
	}
	c.Append(f)
}
func (c *Controller) changed() {
	if c.OnChange != nil {
		c.OnChange()
	}
}
func (c *Controller) Fail(err error) error {
	if err != nil {
		c.State.Notice = err.Error()
		c.State.Error = true
		c.changed()
	}
	return err
}
func (c *Controller) Saved(message string) { c.State.Notice = message; c.State.Error = false }
func (c *Controller) Run(fn func() error) error {
	before := append([]Frame{}, c.Frames...)
	date := c.Date
	err := fn()
	if err == nil {
		err = c.Refresh()
	}
	if err != nil {
		c.Frames = before
		c.Date = date
		c.Fail(err)
	}
	return err
}
func (c *Controller) Dismiss()    { c.State.Notice = ""; c.State.Error = false; c.changed() }
func (c *Controller) Home() error { return c.Run(func() error { c.Reset(); return nil }) }
func (c *Controller) Back() error { return c.Run(func() error { c.Pop(); return nil }) }
func (c *Controller) Page(direction int) error {
	return c.Run(func() error {
		f := &c.Frames[len(c.Frames)-1]
		f.Page = max(0, min(f.Page+direction, c.State.PageCount-1))
		return nil
	})
}
func (c *Controller) Maintenance() error {
	return c.Run(func() error { c.Reset(); c.Push(Maintenance, nil); return nil })
}
func (c *Controller) OpenSetup() error { return c.Maintenance() }
func (c *Controller) active() (Feature, error) {
	f, ok := c.Features[c.Frame().Screen]
	if !ok {
		return nil, fmt.Errorf("unknown swimming route %s", c.Frame().Screen)
	}
	return f, nil
}
func (c *Controller) dispatch(fn func(Feature) error) error {
	return c.Run(func() error {
		f, e := c.active()
		if e != nil {
			return e
		}
		return fn(f)
	})
}
func (c *Controller) Header() error { return c.dispatch(func(f Feature) error { return f.Header() }) }
func (c *Controller) Select(id string) error {
	for _, r := range c.State.Rows {
		if r.ID == id && !r.Disabled {
			return c.dispatch(func(f Feature) error { return f.Select(id) })
		}
	}
	return nil
}
func (c *Controller) Mark(status string) error {
	return c.dispatch(func(f Feature) error { return f.Mark(status) })
}
func (c *Controller) Submit() error { return c.dispatch(func(f Feature) error { return f.Submit() }) }
func (c *Controller) Refresh() error {
	feature, err := c.active()
	if err != nil {
		return c.Fail(err)
	}
	state, err := feature.Build()
	if err != nil {
		return c.Fail(err)
	}
	state.Notice = c.State.Notice
	state.Error = c.State.Error
	c.State = state
	c.changed()
	return nil
}
func (c *Controller) InitialState() State {
	return State{Screen: string(c.Frame().Screen), Title: "My Classes", TitleDetails: true, DetailsTitle: "Lesson details", CanBack: len(c.Frames) > 1, PageCount: 1, PageSize: 3, Rows: []Row{}}
}
func (c *Controller) Paginate(state *State, rows []Row) {
	size := max(1, state.PageSize)
	state.RowCount = len(rows)
	state.PageCount = max(1, (len(rows)+size-1)/size)
	state.Page = min(max(0, c.Frame().Page), state.PageCount-1)
	c.Frames[len(c.Frames)-1].Page = state.Page
	start := state.Page * size
	state.Rows = rows[start:min(start+size, len(rows))]
}
func ProgressText(p domain.Progress) string {
	if p.Complete {
		return "Complete · 3/3"
	}
	return fmt.Sprintf("%d/3 successful attempts", p.Successes)
}

type LessonReader interface {
	GetCourse(string) (domain.Course, error)
	ListRoster(string) ([]domain.Enrollment, error)
}

func LessonContext(s LessonReader, f Frame, date string, state *State) (domain.Course, []domain.Enrollment, error) {
	course, e := s.GetCourse(f.Course)
	if e != nil {
		return course, nil, e
	}
	roster, e := s.ListRoster(f.Course)
	if e != nil {
		return course, nil, e
	}
	state.Subtitle = course.Name + " · " + course.Time + " · " + date
	state.Context = state.Subtitle
	return course, roster, nil
}
