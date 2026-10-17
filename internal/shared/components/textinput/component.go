// Package textinput provides a COB-styled native Fyne entry.
package textinput

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/components/design"
)

type Tone string

const (
	Default Tone = "default"
	Accent  Tone = "accent"
)

type Options struct {
	Tone               Tone
	Value, Placeholder string
	Disabled           bool
	OnChanged          func(string)
}

// Input embeds the native control for editing, focus, validation, and binding APIs.
// Add Object() to a layout to retain the component's scoped styling.
type Input struct {
	*widget.Entry
	object  fyne.CanvasObject
	control *keyEntry
	tone    Tone
}

func New(o Options) *Input {
	c := &Input{tone: o.Tone}
	c.object = c.build(o)
	c.OnChanged = o.OnChanged
	if o.Disabled {
		c.Disable()
	}
	return c
}
func (c *Input) Object() fyne.CanvasObject { return c.object }

type keyEntry struct{ widget.Entry }

func (c *keyEntry) TypedKey(key *fyne.KeyEvent) {
	if design.HandleEscape(c, key) {
		return
	}
	c.Entry.TypedKey(key)
}
