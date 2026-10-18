package emptystate

import (
	"fyne.io/fyne/v2"
	"rectablet/src/shared/components/design"
)

func build(o Options) fyne.CanvasObject {
	items := []fyne.CanvasObject{}
	if o.Title != "" {
		items = append(items, design.Label(o.Title, 14, design.Secondary, true))
	}
	if o.Description != "" {
		items = append(items, design.Label(o.Description, 14, design.Alpha(design.Secondary, 179), false))
	}
	if o.Content != nil {
		items = append(items, o.Content)
	}
	return design.Frame(design.Column(8, items...), design.Background, design.Alpha(design.Secondary, 51), 16, 16, 1, false)
}
