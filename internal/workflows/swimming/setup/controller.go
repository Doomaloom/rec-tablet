package setup

import (
	"path/filepath"
	"rectablet/internal/workflows/swimming/navigation"
	"strings"
	"time"
)

type Controller struct {
	*navigation.Controller
	Store           Repository
	JoinShare       func(code string, done func(error))
	Profiles        []ProfileOption
	SwitchProfile   func(string) error
	ShareServer     string
	SaveShareServer func(string) error
	CancelShare     func()
}

func New(c *navigation.Controller, s Repository) *Controller {
	return &Controller{Controller: c, Store: s}
}
func (c *Controller) Build() (navigation.State, error) {
	f := c.Frame()
	state := c.InitialState()
	rows := []navigation.Row{}
	switch f.Screen {
	case navigation.Maintenance:
		state.Title = "Device Setup"
		state.OnLeave = c.CancelShare
		rows = []navigation.Row{
			{ID: "import", Label: "Import Class Data · Class package (.json)"},
			{ID: "share", Label: "Import from Session Code"},
			{ID: "export", Label: "Export Data · Attendance, attempts, and drafts"},
		}
		for _, profile := range c.Profiles {
			rows = append(rows, navigation.Row{ID: "profile:" + profile.Key, Label: "Switch Session · " + profile.Label})
		}

	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Controller) Select(id string) error {
	if strings.HasPrefix(id, "profile:") && c.SwitchProfile != nil {
		return c.SwitchProfile(strings.TrimPrefix(id, "profile:"))
	}
	if id == "server" {
		return nil
	}
	return nil
}
func (c *Controller) Header() error     { return nil }
func (c *Controller) Mark(string) error { return nil }
func (c *Controller) Submit() error     { return nil }
func (c *Controller) Import(path string) error {
	return c.Run(func() error {
		_, e := c.Store.ImportFile(path)
		if e != nil {
			return e
		}
		dates, e := c.Store.Dates()
		if e != nil {
			return e
		}
		found := false
		for _, d := range dates {
			if d == c.Date {
				found = true
			}
		}
		if !found && len(dates) > 0 {
			nearest := dates[0]
			distance := time.Duration(1<<63 - 1)
			for _, d := range dates {
				t, _ := time.Parse("2006-01-02", d)
				delta := time.Since(t).Abs()
				if delta < distance {
					distance = delta
					nearest = d
				}
			}
			c.Date = nearest
		}
		c.Saved("")
		return nil
	})
}
func (c *Controller) Export(path string) error {
	return c.Run(func() error {
		if strings.ToLower(filepath.Ext(path)) != ".json" {
			path = strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
		}
		if e := c.Store.Export(path); e != nil {
			return e
		}
		c.Saved("")
		return nil
	})
}
