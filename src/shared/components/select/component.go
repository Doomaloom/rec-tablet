// Package selectcontrol provides a COB-styled native Fyne select.
package selectcontrol

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/design"
)

type Options struct {
	Options            []string
	Value, Placeholder string
	Disabled           bool
	OnChanged          func(string)
}

// Select embeds the native control for editing, focus, validation, and binding APIs.
// Add Object() to a layout to retain the component's scoped styling.
type Select struct {
	*widget.Select
	object  fyne.CanvasObject
	control *keySelect
}

func New(o Options) *Select {
	c := &Select{}
	c.object = c.build(o)
	c.OnChanged = o.OnChanged
	if o.Disabled {
		c.Disable()
	}
	return c
}
func (c *Select) Object() fyne.CanvasObject { return c.object }

type keySelect struct{ widget.Select }

func (c *keySelect) TypedKey(key *fyne.KeyEvent) {
	if design.HandleEscape(c, key) {
		return
	}
	c.Select.TypedKey(key)
}
