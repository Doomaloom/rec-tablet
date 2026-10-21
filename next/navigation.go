package next

import (
	"rectablet/next/contracts"
	"rectablet/next/navigation"
)

// Snapshot and Subscribe expose the workflow's presentation boundary.
func (c *Controller) Snapshot() contracts.State { return c.navigation.Snapshot() }
func (c *Controller) Subscribe(changed func())  { c.navigation.Subscribe(changed) }

func (c *Controller) Home() error              { return c.navigation.Home() }
func (c *Controller) Back() error              { return c.navigation.Back() }
func (c *Controller) Page(direction int) error { return c.navigation.Page(direction) }
func (c *Controller) Header() error            { return c.navigation.Header() }
func (c *Controller) Select(id string) error   { return c.navigation.Select(id) }
func (c *Controller) Mark(status string) error { return c.navigation.Mark(status) }
func (c *Controller) Submit() error            { return c.navigation.Submit() }
func (c *Controller) Dismiss()                 { c.navigation.Dismiss() }
func (c *Controller) OpenSetup() error         { return c.navigation.OpenSetup() }
func (c *Controller) Maintenance() error       { return c.navigation.Maintenance() }
func (c *Controller) Refresh() error           { return c.navigation.Refresh() }
func (c *Controller) Fail(err error) error     { return c.navigation.Fail(err) }

// Frame returns a copy of the active route, while Push stages a route change.
// Call Refresh after Push to build and display the new route.
func (c *Controller) Frame() navigation.Frame { return c.navigation.Frame() }
func (c *Controller) Push(route navigation.Route, change func(*navigation.Frame)) {
	c.navigation.Push(route, change)
}

func (c *Controller) Date() string { return c.navigation.Date }

// SetDate stages the lesson date; Refresh applies it to the active screen.
func (c *Controller) SetDate(date string) { c.navigation.Date = date }
