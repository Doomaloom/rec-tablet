package textarea

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/src/shared/components/design"
)

func (c *Textarea) build(o Options) fyne.CanvasObject {
	c.control = &keyEntry{Entry: widget.Entry{PlaceHolder: o.Placeholder, Text: o.Value, OnChanged: nil}}
	c.Entry = &c.control.Entry
	c.MultiLine = true
	c.Wrapping = fyne.TextWrapWord
	c.control.ExtendBaseWidget(c.control)
	c.SetMinRowsVisible(4)

	minimum := canvas.NewRectangle(color.Transparent)
	minimum.SetMinSize(fyne.NewSize(44, 120))
	return container.NewThemeOverride(container.NewStack(minimum, c.control), design.Theme{Input: design.Background})
}
