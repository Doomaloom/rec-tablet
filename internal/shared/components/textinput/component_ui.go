package textinput

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/internal/shared/components/design"
)

func (c *Input) build(o Options) fyne.CanvasObject {
	c.control = &keyEntry{Entry: widget.Entry{PlaceHolder: o.Placeholder, Text: o.Value, OnChanged: nil}}
	c.Entry = &c.control.Entry
	c.Wrapping = fyne.TextWrapOff
	c.control.ExtendBaseWidget(c.control)
	fill := design.Background
	if c.tone == Accent {
		fill = design.White
	}
	minimum := canvas.NewRectangle(color.Transparent)
	minimum.SetMinSize(fyne.NewSize(44, 44))
	return container.NewThemeOverride(container.NewStack(minimum, c.control), design.Theme{Input: fill})
}
