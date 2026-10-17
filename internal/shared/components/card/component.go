// Package card provides reusable COB panel surfaces.
package card

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Variant string

const (
	Panel       Variant = "panel"
	Subtle      Variant = "subtle"
	Interactive Variant = "interactive"
	Plain       Variant = "plain"
)

type Card struct {
	widget.BaseWidget
	content fyne.CanvasObject
	variant Variant
	hovered bool
}

func New(content fyne.CanvasObject, variant Variant) *Card {
	c := &Card{content: content, variant: variant}
	c.ExtendBaseWidget(c)
	return c
}
func (c *Card) SetContent(content fyne.CanvasObject) { c.content = content; c.Refresh() }
func (c *Card) MouseIn(*desktop.MouseEvent)          { c.hovered = true; c.Refresh() }
func (c *Card) MouseMoved(*desktop.MouseEvent)       {}
func (c *Card) MouseOut()                            { c.hovered = false; c.Refresh() }
