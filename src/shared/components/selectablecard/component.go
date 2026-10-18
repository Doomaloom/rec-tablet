// Package selectablecard provides selectable COB cards.
package selectablecard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/actionbutton"
)

type Options struct{ Selected, Disabled bool }
type Card struct {
	widget.BaseWidget
	button *actionbutton.Button
}

func New(content fyne.CanvasObject, options Options, onTap func()) *Card {
	c := &Card{}
	c.ExtendBaseWidget(c)
	c.build(content, options, onTap)
	return c
}
func (c *Card) SetSelected(selected bool)            { c.button.SetSelected(selected) }
func (c *Card) Selected() bool                       { return c.button.Selected() }
func (c *Card) SetContent(content fyne.CanvasObject) { c.button.SetContent(content); c.Refresh() }
func (c *Card) Disable()                             { c.button.Disable() }
func (c *Card) Enable()                              { c.button.Enable() }
func (c *Card) Disabled() bool                       { return c.button.Disabled() }
func (c *Card) Tapped(e *fyne.PointEvent)            { c.button.Tapped(e) }
