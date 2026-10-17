package selectcontrol

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/internal/shared/components/design"
)

func (c *Select) build(o Options) fyne.CanvasObject {
	c.control = &keySelect{Select: widget.Select{Options: append([]string(nil), o.Options...), OnChanged: nil}}
	c.Select = &c.control.Select
	c.control.ExtendBaseWidget(c.control)
	c.PlaceHolder = o.Placeholder
	for _, v := range o.Options {
		if v == o.Value {
			c.Selected = v
			break
		}
	}

	minimum := canvas.NewRectangle(color.Transparent)
	minimum.SetMinSize(fyne.NewSize(44, 44))
	return container.NewThemeOverride(container.NewStack(minimum, design.Frame(c.control, design.Background, design.Secondary, 16, 2, 2, false)), design.Theme{Input: design.Background})
}
