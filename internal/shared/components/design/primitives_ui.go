package design

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

func Label(text string, size float32, ink color.Color, bold bool) fyne.CanvasObject {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.TextStyle.Bold = bold
	return container.NewThemeOverride(label, Theme{Foreground: ink, TextSize: size})
}
func Pad(obj fyne.CanvasObject, p float32) *fyne.Container {
	return container.New(layout.NewCustomPaddedLayout(p, p, p, p), obj)
}
func Frame(content fyne.CanvasObject, fill, border color.Color, radius, padding, stroke float32, shadow bool) fyne.CanvasObject {
	bg := canvas.NewRectangle(fill)
	bg.CornerRadius = radius
	bg.StrokeColor = border
	bg.StrokeWidth = stroke
	if shadow {
		bg.Shadow = canvas.Shadow{Color: Alpha(Secondary, 35), BlurRadius: 8, Offset: fyne.NewPos(0, 3)}
	}
	return container.NewStack(bg, Pad(content, padding))
}

// Column uses explicit spacing so enclosing themes cannot change component geometry.
func Column(gap float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&column{gap: gap}, objects...)
}

type column struct{ gap float32 }

func (c *column) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := fyne.Size{}
	n := 0
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		m := o.MinSize()
		s.Width = max(s.Width, m.Width)
		s.Height += m.Height
		n++
	}
	s.Height += float32(max(0, n-1)) * c.gap
	return s
}
func (c *column) Layout(objects []fyne.CanvasObject, s fyne.Size) {
	y := float32(0)
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		o.Resize(fyne.NewSize(s.Width, o.MinSize().Height))
		o.Move(fyne.NewPos(0, y))
		o.Resize(fyne.NewSize(s.Width, o.MinSize().Height))
		y += o.MinSize().Height + c.gap
	}
}
