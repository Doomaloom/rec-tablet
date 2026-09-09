package setup

import (
	"fmt"
	"path/filepath"
	"rectablet/internal/workflows/swimming/navigation"
	"strings"
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
	case navigation.Maintenance:
		state.Title = "Device Setup"
		state.Subtitle = "Temporary file transfer · Works offline"
		rows = []navigation.Row{{ID: "import", Label: "Import Class Data", Detail: "Choose a class package (.json)"}, {ID: "export", Label: "Export Data", Detail: "Save all attendance, attempts, and drafts"}}
		last, e := s.Meta("last_import")
		if e != nil {
			return state, e
		}
		if last == "" {
			last = "Never"
		}
		state.Context = "Database: " + s.DatabasePath() + "\nLast import: " + last

	}
	c.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}
func (c *Controller) Select(string) error { return nil }
func (c *Controller) Header() error       { return nil }
func (c *Controller) Mark(string) error   { return nil }
func (c *Controller) Submit() error       { return nil }
func (c *Controller) Import(path string) error {
	return c.Run(func() error {
		n, e := c.Store.ImportFile(path)
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
		c.Saved(fmt.Sprintf("Imported %d course(s) · Saved on device", n))
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
		c.Saved("Export saved · " + filepath.Base(path))
		return nil
	})
}
