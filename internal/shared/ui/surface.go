package ui

import (
	"fyne.io/fyne/v2"
	"image/color"
	"time"
)

// Surface exposes layout primitives to feature views in the 800x480 design space.
type BodyComposer func(surface *Surface)
type ComponentAdder func(object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32))
type NamedComponentAdder func(name string, object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32))

type Surface struct {
	Window    fyne.Window
	Add       ComponentAdder
	AddNamed  NamedComponentAdder
	Button    func(string, string, func()) *TouchButton
	Criterion *TextBlock
}
type Item struct {
	ID, Label, Detail, Status    string
	Disabled                     bool
	Data                         any
	Fill                         color.Color
	ProgressSegments             int
	ProgressValue, ProgressTotal int
	OnHold                       func()
	HoldDuration                 time.Duration
}

func List(surface *Surface, rows []Item, selectRow func(string)) {
	for i, row := range rows {
		i, row := i, row
		b := surface.Button("row:"+row.ID, row.Label, func() { selectRow(row.ID) })
		b.Detail = row.Detail
		b.Disabled = row.Disabled
		b.OnHold = row.OnHold
		b.HoldDuration = row.HoldDuration
		b.ProgressSegments = row.ProgressSegments
		b.ProgressValue = row.ProgressValue
		b.ProgressTotal = row.ProgressTotal
		if row.Fill != nil {
			b.Fill = row.Fill
		}
		rowPosition := func(w, h float32) (float32, float32, float32, float32) {
			height := (h - 200) / 3
			return 100, 118 + float32(i)*(height+10), w - 200, height
		}
		if surface.AddNamed != nil {
			surface.AddNamed("list-row-"+row.ID, b, rowPosition)
		} else {
			surface.Add(b, rowPosition)
		}
	}
}
func Empty(surface *Surface, message string) {
	empty := NewText(message, 24, Muted)
	empty.Center = true
	surface.Add(empty, func(w, h float32) (float32, float32, float32, float32) { return 124, 134, w - 248, h - 230 })
}
