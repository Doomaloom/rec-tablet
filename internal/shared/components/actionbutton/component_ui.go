package actionbutton

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/internal/shared/components/design"
)

func (b *Button) CreateRenderer() fyne.WidgetRenderer {
	r := &renderer{b: b, bg: canvas.NewRectangle(color.Transparent), focus: canvas.NewRectangle(color.Transparent), host: container.NewStack()}
	r.objects = []fyne.CanvasObject{r.focus, r.bg, r.host}
	r.Refresh()
	return r
}

type renderer struct {
	b         *Button
	bg, focus *canvas.Rectangle
	host      *fyne.Container
	objects   []fyne.CanvasObject
}

func (r *renderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *renderer) Destroy()                     {}
func (r *renderer) padding() float32 {
	if r.b.variant == Selectable {
		return 16
	}
	switch r.b.size {
	case Small:
		return 12
	case Large:
		return 20
	}
	return 16
}
func (r *renderer) MinSize() fyne.Size {
	m := r.host.MinSize()
	return fyne.NewSize(m.Width+2*r.padding(), max(design.Target, m.Height+16))
}
func (r *renderer) Layout(s fyne.Size) {
	r.focus.Move(fyne.NewPos(1, 1))
	r.focus.Resize(fyne.NewSize(max(0, s.Width-2), max(0, s.Height-2)))
	r.bg.Move(fyne.NewPos(4, 4))
	r.bg.Resize(fyne.NewSize(max(0, s.Width-8), max(0, s.Height-8)))
	p := r.padding()
	r.host.Move(fyne.NewPos(p, 8))
	r.host.Resize(fyne.NewSize(max(0, s.Width-2*p), max(0, s.Height-16)))
}
func (r *renderer) Refresh() {
	b := r.b
	var fill, ink, border color.Color = design.Secondary, design.White, color.Transparent
	stroke := float32(0)
	switch b.variant {
	case Primary:
		fill = design.Primary
		if b.hovered {
			fill = design.Secondary
		}
	case Danger:
		fill = design.Danger
		if b.hovered {
			fill = design.DangerHover
		}
	case Outline:
		fill = color.Transparent
		ink = design.Secondary
		border = design.Alpha(design.Secondary, 102)
		stroke = 1
		if b.hovered {
			fill = design.Background
		}
	case Ghost:
		fill = color.Transparent
		ink = design.Secondary
		if b.hovered {
			fill = design.Background
		}
	case Selectable:
		stroke = 1
		fill = design.Background
		ink = design.Secondary
		border = design.Alpha(design.Secondary, 51)
		if b.hovered {
			border = design.Secondary
		}
		if b.selected {
			fill = design.Secondary
			ink = design.White
			border = design.Secondary
		}
	default:
		if b.hovered {
			fill = design.White
			ink = design.Secondary
		}
	}
	if b.Disabled() {
		fill = design.Opacity(fill, 153)
		ink = design.Opacity(ink, 153)
	}
	r.bg.FillColor = fill
	r.bg.StrokeColor = border
	r.bg.StrokeWidth = stroke
	r.bg.CornerRadius = design.ControlRadius
	if b.radius > 0 {
		r.bg.CornerRadius = b.radius
	}
	r.focus.CornerRadius = r.bg.CornerRadius + 3
	r.focus.StrokeWidth = 0
	if b.focused {
		r.focus.StrokeColor = design.Primary
		r.focus.StrokeWidth = 2
	}
	content := b.content
	if content == nil {
		label := widget.NewLabel(b.text)
		label.Alignment = fyne.TextAlignCenter
		label.TextStyle.Bold = true
		label.Wrapping = fyne.TextWrapWord
		content = label
	}
	textSize := float32(14)
	if b.size == Large {
		textSize = 16
	}
	r.host.Objects = []fyne.CanvasObject{container.NewThemeOverride(content, design.Theme{Foreground: ink, TextSize: textSize})}
	r.host.Refresh()
	r.Layout(b.Size())
	r.bg.Refresh()
	r.focus.Refresh()
}
