package selectablecard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/actionbutton"
)

func (c *Card) build(content fyne.CanvasObject, o Options, onTap func()) {
	c.button = actionbutton.NewContent(content, actionbutton.Options{Variant: actionbutton.Selectable, Disabled: o.Disabled}, onTap)
	c.button.SetSelected(o.Selected)
}
func (c *Card) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(c.button) }
