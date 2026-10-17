package segmentedtabs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"rectablet/internal/shared/components/actionbutton"
)

func (t *Tabs) build() {
	for _, item := range t.items {
		item := item
		b := actionbutton.New(item.Label, actionbutton.Options{Variant: actionbutton.Selectable, Size: t.options.Size, Disabled: item.Disabled}, func() { t.choose(item.Key) })
		if t.options.Shape == Pill {
			b.SetRadius(999)
		}
		t.buttons = append(t.buttons, b)
	}
}
func (t *Tabs) CreateRenderer() fyne.WidgetRenderer {
	objects := []fyne.CanvasObject{}
	for i, b := range t.buttons {
		b.SetSelected(t.items[i].Key == t.options.ActiveKey)
		objects = append(objects, b)
	}
	return &renderer{t: t, host: container.New(layout.NewGridWrapLayout(fyne.NewSize(140, 48)), objects...)}
}

type renderer struct {
	t    *Tabs
	host *fyne.Container
}

func (r *renderer) MinSize() fyne.Size           { return r.host.MinSize() }
func (r *renderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.host} }
func (r *renderer) Layout(s fyne.Size)           { r.host.Resize(s) }
func (r *renderer) Destroy()                     {}
func (r *renderer) Refresh() {
	for i, b := range r.t.buttons {
		b.SetSelected(r.t.items[i].Key == r.t.options.ActiveKey)
	}
	r.host.Refresh()
}
