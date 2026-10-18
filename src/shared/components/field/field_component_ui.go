package field

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/src/shared/components/design"
)

func (f *Field) content() fyne.CanvasObject {
	o := f.options
	items := []fyne.CanvasObject{}
	if o.Label != "" {
		items = append(items, design.Label(o.Label, 14, design.Secondary, true))
	}
	items = append(items, f.control)
	if o.Error != "" {
		items = append(items, design.Label(o.Error, 12, design.Danger, true))
	} else if o.Helper != "" {
		items = append(items, design.Label(o.Helper, 12, design.Alpha(design.Secondary, 179), false))
	}
	return design.Column(8, items...)
}
func (f *Field) CreateRenderer() fyne.WidgetRenderer {
	return &renderer{f: f, host: container.NewStack(f.content())}
}

type renderer struct {
	f    *Field
	host *fyne.Container
}

func (r *renderer) MinSize() fyne.Size           { return r.host.MinSize() }
func (r *renderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.host} }
func (r *renderer) Layout(s fyne.Size)           { r.host.Resize(s) }
func (r *renderer) Destroy()                     {}
func (r *renderer) Refresh() {
	r.host.Objects = []fyne.CanvasObject{r.f.content()}
	r.host.Refresh()
	r.Layout(r.f.Size())
}
