package modalshell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image/color"
	"rectablet/internal/shared/components/actionbutton"
	"rectablet/internal/shared/components/design"
)

func (m *Modal) build() {
	m.close = actionbutton.New("×", actionbutton.Options{Variant: actionbutton.Outline}, m.Hide)
	m.close.SetRadius(999)
	heading := container.NewBorder(nil, nil, nil, m.close, design.Label(m.options.Title, 24, design.Secondary, true))
	items := []fyne.CanvasObject{heading}
	if m.options.Description != "" {
		items = append(items, design.Label(m.options.Description, 14, design.Secondary, false))
	}
	if m.options.Notice != nil {
		items = append(items, m.options.Notice)
	}
	items = append(items, m.content)
	m.body = design.Column(16, items...)
	m.panel = design.Frame(design.Scope(container.NewVScroll(m.body)), design.White, design.Alpha(design.Secondary, 51), 24, 24, 2, true)
}
func (m *Modal) CreateRenderer() fyne.WidgetRenderer {
	return &renderer{m: m, bg: canvas.NewRectangle(color.NRGBA{0, 0, 0, 102})}
}

type renderer struct {
	m  *Modal
	bg *canvas.Rectangle
}

func (r *renderer) MinSize() fyne.Size           { return fyne.NewSize(1, 1) }
func (r *renderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.bg, r.m.panel} }
func (r *renderer) Layout(s fyne.Size) {
	r.bg.Resize(s)
	w := max(0, min(r.m.options.MaxWidth, s.Width-32))
	r.m.body.Resize(fyne.NewSize(max(0, w-48), r.m.body.MinSize().Height))
	h := max(0, min(min(560, s.Height-48), max(200, r.m.body.MinSize().Height+48)))
	r.m.panelPos = fyne.NewPos((s.Width-w)/2, (s.Height-h)/2)
	r.m.panelSize = fyne.NewSize(w, h)
	r.m.panel.Move(r.m.panelPos)
	r.m.panel.Resize(r.m.panelSize)
}
func (r *renderer) Destroy() {}
func (r *renderer) Refresh() { r.Layout(r.m.Size()); r.bg.Refresh(); r.m.panel.Refresh() }
