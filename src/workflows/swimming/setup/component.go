package setup

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/navigation"
)

// Component owns setup persistence and navigation through four child components.
type Component struct {
	navigation      *navigation.Controller
	store           Repository
	fileTransfer    FileTransfer
	sessionSharing  SessionSharing
	displayPicker   DisplayPicker
	profilePicker   ProfilePicker
	JoinShare       func(code string, done func(error))
	Profiles        []ProfileOption
	SwitchProfile   func(string) error
	ShareServer     string
	SaveShareServer func(string) error
	CancelShare     func()
	SaveDisplayMode func(ui.DisplayMode)
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	reportError := func(err error) { c.navigation.Fail(err) }
	c.fileTransfer = FileTransfer{onImport: c.Import, onExport: c.Export, onError: reportError}
	c.sessionSharing = SessionSharing{
		server:       func() string { return c.ShareServer },
		onJoin:       c.joinSession,
		onSaveServer: c.saveShareServer,
		onError:      reportError,
	}
	c.displayPicker = DisplayPicker{onSelect: c.SetDisplayMode}
	c.profilePicker = ProfilePicker{onSelect: func(key string) error {
		if c.SwitchProfile != nil {
			return c.SwitchProfile(key)
		}
		return nil
	}}
	return c
}

func (c *Component) Build() (navigation.State, error) {
	f := c.navigation.Frame()
	state := c.navigation.InitialState()
	rows := []navigation.Row{}
	switch f.Screen {
	case navigation.Maintenance:
		state.Title = "Device Setup"
		state.OnLeave = c.CancelShare
		rows = []navigation.Row{
			c.fileTransfer.ImportRow(), c.sessionSharing.Row(), c.fileTransfer.ExportRow(),
			c.displayPicker.Row(c.navigation.DisplayMode),
		}
		rows = append(rows, c.profilePicker.Rows(c.Profiles)...)

	}
	c.navigation.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}

func (c *Component) SetDisplayMode(mode ui.DisplayMode) error {
	mode = mode.Normalize()
	previous := c.navigation.DisplayMode
	c.navigation.DisplayMode = mode
	if err := c.navigation.Refresh(); err != nil {
		c.navigation.DisplayMode = previous
		return err
	}
	if c.SaveDisplayMode != nil {
		c.SaveDisplayMode(mode)
	}
	return nil
}

func (c *Component) Select(id string) error { return c.profilePicker.Select(id) }

// joinSession coordinates status and asynchronous download errors in navigation.
func (c *Component) joinSession(code string) error {
	if c.JoinShare == nil {
		return fmt.Errorf("session sharing is not configured")
	}
	c.navigation.State.Notice = "Connecting to the sharing server…"
	c.navigation.State.Error = false
	_ = c.navigation.Refresh()
	c.JoinShare(code, func(err error) {
		if err != nil {
			c.navigation.Fail(err)
		}
	})
	return nil
}

func (c *Component) saveShareServer(server string) error {
	if c.SaveShareServer == nil {
		return fmt.Errorf("server settings are unavailable")
	}
	if err := c.SaveShareServer(server); err != nil {
		return err
	}
	c.ShareServer = strings.TrimRight(server, "/")
	_ = c.navigation.Refresh()
	return nil
}

func (c *Component) Header() error     { return nil }
func (c *Component) Mark(string) error { return nil }
func (c *Component) Submit() error     { return nil }
func (c *Component) Import(path string) error {
	return c.navigation.Run(func() error {
		_, e := c.store.ImportFile(path)
		if e != nil {
			return e
		}
		dates, e := c.store.Dates()
		if e != nil {
			return e
		}
		found := false
		for _, d := range dates {
			if d == c.navigation.Date {
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
			c.navigation.Date = nearest
		}
		c.navigation.Saved("")
		return nil
	})
}
func (c *Component) Export(path string) error {
	return c.navigation.Run(func() error {
		if strings.ToLower(filepath.Ext(path)) != ".json" {
			path = strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
		}
		if e := c.store.Export(path); e != nil {
			return e
		}
		c.navigation.Saved("")
		return nil
	})
}
