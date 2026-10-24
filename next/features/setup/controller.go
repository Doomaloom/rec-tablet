package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"rectablet/next/features/setup/displaypicker"
	"rectablet/next/features/setup/filetransfer"
	"rectablet/next/features/setup/sessionsharing"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

// Component owns setup persistence and navigation through export, server, and display components.
type Component struct {
	navigation      *navigation.Controller
	store           Repository
	fileTransfer    filetransfer.FileTransfer
	sessionSharing  sessionsharing.SessionSharing
	displayPicker   displaypicker.DisplayPicker
	ShareServer     string
	SaveShareServer func(string) error
	SaveDisplayMode func(ui.DisplayMode)
}

var _ navigation.Feature = (*Component)(nil)

func New(nav *navigation.Controller, store Repository) *Component {
	c := &Component{navigation: nav, store: store}
	reportError := func(err error) { c.navigation.Fail(err) }
	c.fileTransfer = filetransfer.FileTransfer{OnExport: c.Export, OnError: reportError}
	c.sessionSharing = sessionsharing.SessionSharing{
		Server:       func() string { return c.ShareServer },
		OnSaveServer: c.saveShareServer,
		OnError:      reportError,
	}
	c.displayPicker = displaypicker.DisplayPicker{OnSelect: c.SetDisplayMode}
	return c
}

func (c *Component) Build() (navigation.State, error) {
	f := c.navigation.Frame()
	state := c.navigation.InitialState()
	rows := []navigation.Row{}
	switch f.Screen {
	case navigation.Maintenance:
		state.Title = "Device Setup"
		rows = []navigation.Row{
			c.fileTransfer.ExportRow(),
			{ID: "server", Label: "Configure Share Server"},
			c.displayPicker.Row(c.navigation.DisplayMode),
		}

	}
	c.navigation.Paginate(&state, rows)
	c.composeFeatureBody(&state)
	return state, nil
}

func (c *Component) SetDisplayMode(mode ui.DisplayMode) error {
	if c.navigation.Closed {
		return nil
	}
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

func (c *Component) Select(string) error { return nil }

func (c *Component) saveShareServer(server string) error {
	if c.navigation.Closed {
		return nil
	}
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

// Import is deliberately blocked: only a verified login may select class data.
func (c *Component) Import(string) error {
	return c.navigation.Fail(fmt.Errorf("log out and use a device export code to load classes"))
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
