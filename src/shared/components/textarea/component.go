// Package textarea provides a COB-styled native Fyne entry.
package textarea

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/design"
)

type Options struct {
	Value, Placeholder string
	Disabled           bool
	OnChanged          func(string)
}

// Textarea embeds the native control for editing, focus, validation, and binding APIs.
// Add Object() to a layout to retain the component's scoped styling.
type Textarea struct {
	*widget.Entry
	object  fyne.CanvasObject
	control *keyEntry
}

func New(o Options) *Textarea {
	c := &Textarea{}
	c.object = c.build(o)
	c.OnChanged = o.OnChanged
	if o.Disabled {
		c.Disable()
	}
	return c
}
func (c *Textarea) Object() fyne.CanvasObject { return c.object }

type keyEntry struct{ widget.Entry }

func (c *keyEntry) TypedKey(key *fyne.KeyEvent) {
	if design.HandleEscape(c, key) {
		return
	}
	c.Entry.TypedKey(key)
}
