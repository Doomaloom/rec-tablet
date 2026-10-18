package notice

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/src/shared/components/design"
)

func (n *Notice) content() fyne.CanvasObject {
	ink := design.Secondary
	fill := design.Background
	switch n.tone {
	case Success:
		ink = design.Primary
	case Warning:
		ink = design.Header
	case Danger:
		ink = design.Danger
	}
	content := n.body
	if content == nil {
		content = design.Label(n.text, 14, ink, true)
	}
	if n.tone == Danger {
		return design.Frame(content, design.Alpha(design.Danger, 26), design.Alpha(ink, 77), 16, 12, 1, false)
	}
	return design.Frame(content, fill, design.Alpha(ink, 51), 16, 12, 1, false)
}
func (n *Notice) CreateRenderer() fyne.WidgetRenderer {
	return &renderer{n: n, host: container.NewStack(n.content())}
}

type renderer struct {
	n    *Notice
	host *fyne.Container
}

func (r *renderer) MinSize() fyne.Size           { return r.host.MinSize() }
func (r *renderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.host} }
func (r *renderer) Layout(s fyne.Size)           { r.host.Resize(s) }
func (r *renderer) Destroy()                     {}
func (r *renderer) Refresh() {
	r.host.Objects = []fyne.CanvasObject{r.n.content()}
	r.host.Refresh()
	r.Layout(r.n.Size())
}
