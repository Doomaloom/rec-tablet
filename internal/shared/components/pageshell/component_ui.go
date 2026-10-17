package pageshell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/internal/shared/components/design"
)

func build(o Options, content []fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&pageLayout{width: o.MaxWidth}, design.Column(24, content...))
}

type pageLayout struct{ width float32 }

func (p *pageLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	m := objects[0].MinSize()
	return fyne.NewSize(min(p.width, m.Width)+32, m.Height+48)
}
func (p *pageLayout) Layout(objects []fyne.CanvasObject, s fyne.Size) {
	w := max(0, min(p.width, s.Width-32))
	objects[0].Move(fyne.NewPos((s.Width-w)/2, 24))
	objects[0].Resize(fyne.NewSize(w, max(0, s.Height-48)))
}
