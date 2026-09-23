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
	Bounds    Bounds
}
type Item struct {
	ID, Label, Status            string
	Disabled, Selected           bool
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
		b.MaxLabelLines = -1
		b.Disabled = row.Disabled
		b.Selected = row.Selected
		b.OnHold = row.OnHold
		b.HoldDuration = row.HoldDuration
		b.ProgressSegments = row.ProgressSegments
		b.ProgressValue = row.ProgressValue
		b.ProgressTotal = row.ProgressTotal
		if row.Fill != nil {
			b.Fill = row.Fill
		}
		rowPosition := func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, availableHeight := surface.BodyBounds(w, h)
			height := (availableHeight - 2*Gap) / 3
			return x, y + float32(i)*(height+Gap), width, height
		}
		if surface.AddNamed != nil {
			surface.AddNamed("list-row-"+row.ID, b, rowPosition)
		} else {
			surface.Add(b, rowPosition)
		}
	}
}
func Empty(surface *Surface, message string) {
	empty := NewText(message, 28, Navy)
	empty.Bold = true
	empty.Center = true
	surface.Add(empty, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x + ButtonPadding, y, width - 2*ButtonPadding, height
	})
}
