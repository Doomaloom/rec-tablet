package card

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"image/color"
	"rectablet/src/shared/components/design"
)

func (c *Card) CreateRenderer() fyne.WidgetRenderer {
	return &renderer{c: c, host: container.NewStack(c.surface())}
}
func (c *Card) surface() fyne.CanvasObject {
	var fill, border color.Color = design.White, design.Alpha(design.Secondary, 51)
	radius, padding, stroke := design.CardRadius, float32(24), float32(2)
	shadow := true
	switch c.variant {
	case Subtle:
		fill = design.Background
		radius = 16
		padding = 16
		stroke = 1
		shadow = false
	case Plain:
		fill = color.Transparent
		border = color.Transparent
		padding = 0
		stroke = 0
		shadow = false
	}
	if c.variant == Interactive && c.hovered {
		border = design.Secondary
	}
	return design.Frame(design.Scope(c.content), fill, border, radius, padding, stroke, shadow)
}

type renderer struct {
	c    *Card
	host *fyne.Container
}

func (r *renderer) MinSize() fyne.Size           { return r.host.MinSize() }
func (r *renderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.host} }
func (r *renderer) Layout(s fyne.Size)           { r.host.Resize(s) }
func (r *renderer) Destroy()                     {}
func (r *renderer) Refresh() {
	r.host.Objects = []fyne.CanvasObject{r.c.surface()}
	r.host.Refresh()
	r.Layout(r.c.Size())
}
