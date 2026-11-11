// Package login presents the device export-code gate before lesson access.
package login

import (
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

type Options struct {
	DisplayMode     ui.DisplayMode
	SaveDisplayMode func(ui.DisplayMode)
	Server          func() string
	SaveServer      func(string) error
	Login           func(string, func(error))
	Cancel          func()
}

type Component struct {
	options                Options
	code                   string
	busy, closed, settings bool
	notice                 string
	error                  bool
	changed                func()
}

var _ contracts.Presenter = (*Component)(nil)

func New(options Options) *Component {
	options.DisplayMode = options.DisplayMode.Normalize()
	return &Component{options: options}
}
func (c *Component) Subscribe(changed func()) { c.changed = changed }
func (c *Component) refresh() {
	if !c.closed && c.changed != nil {
		c.changed()
	}
}
func (c *Component) Fail(err error) {
	if c.closed {
		return
	}
	c.busy = false
	c.notice, c.error = err.Error(), true
	c.refresh()
}
func (c *Component) Digit(digit string) {
	if c.closed || c.busy || c.settings || len(c.code) == 6 || len(digit) != 1 || digit[0] < '0' || digit[0] > '9' {
		return
	}
	c.code += digit
	c.notice, c.error = "", false
	c.refresh()
}
func (c *Component) Clear() {
	if c.closed || c.busy {
		return
	}
	c.code, c.notice, c.error = "", "", false
	c.refresh()
}
func (c *Component) Delete() {
	if c.closed || c.busy || len(c.code) == 0 {
		return
	}
	c.code = c.code[:len(c.code)-1]
	c.notice, c.error = "", false
	c.refresh()
}
func (c *Component) Submit() error {
	if c.closed || c.busy || c.settings || len(c.code) != 6 || c.options.Login == nil {
		return nil
	}
	c.busy, c.notice, c.error = true, "", false
	c.refresh()
	c.options.Login(c.code, func(err error) {
		if c.closed {
			return
		}
		if err != nil {
			c.Fail(err)
			return
		}
		c.code = ""
	})
	return nil
}
func (c *Component) OpenSetup() error {
	if c.closed || c.busy {
		return nil
	}
	c.settings = true
	c.notice, c.error = "", false
	c.refresh()
	return nil
}
func (c *Component) Back() error {
	if c.closed || c.busy {
		return nil
	}
	c.settings = false
	c.notice, c.error = "", false
	c.refresh()
	return nil
}
func (c *Component) Home() error    { return c.Back() }
func (c *Component) Header() error  { return nil }
func (c *Component) Page(int) error { return nil }
func (c *Component) Dismiss() {
	if !c.closed {
		c.notice, c.error = "", false
		c.refresh()
	}
}
func (c *Component) Deactivate() {
	if c.closed {
		return
	}
	c.closed = true
	c.code = ""
	c.changed = nil
	if c.options.Cancel != nil {
		c.options.Cancel()
	}
}
