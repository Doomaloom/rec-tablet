package ui

import (
	"fyne.io/fyne/v2"
	"image/color"
	"rectablet/internal/shared/components/emptystate"
	"time"
)

// Surface exposes layout primitives in the selected profile's logical design space.
type BodyComposer func(surface *Surface)
type ComponentAdder func(object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32))
type NamedComponentAdder func(name string, object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32))

type Surface struct {
	DisplayMode DisplayMode
	Window      fyne.Window
	Add         ComponentAdder
	AddNamed    NamedComponentAdder
	Button      func(string, string, func()) *TouchButton
	Criterion   *TextBlock
	Bounds      Bounds
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
	count := 3
	if surface.DisplayMode.IsLarge() {
		count = 6
	}
	Grid(surface, rows, 1, count, selectRow)
}

// Grid renders a page in row-major order. Callers own slicing and pagination.
// rowsPerPage is the number of visual rows, not the total number of cells.
func Grid(surface *Surface, rows []Item, columns, rowsPerPage int, selectRow func(string)) {
	columns, rowsPerPage = max(1, columns), max(1, rowsPerPage)
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
			height := (availableHeight - float32(rowsPerPage-1)*Gap) / float32(rowsPerPage)
			cellWidth := (width - float32(columns-1)*Gap) / float32(columns)
			return x + float32(i%columns)*(cellWidth+Gap), y + float32(i/columns)*(height+Gap), cellWidth, height
		}
		if surface.AddNamed != nil {
			surface.AddNamed("list-row-"+row.ID, b, rowPosition)
		} else {
			surface.Add(b, rowPosition)
		}
	}
}
func Empty(surface *Surface, message string) {
	panel := emptystate.New(emptystate.Options{Title: message})
	surface.Add(panel, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x + ButtonPadding, y, width - 2*ButtonPadding, height
	})
}

// SelectableSurface styles only this surface's choices as selectable cards.
// The original button factory still owns activation guards and long presses.
func SelectableSurface(surface *Surface) *Surface {
	scoped := *surface
	scoped.Button = func(id, label string, run func()) *TouchButton {
		button := surface.Button(id, label, run)
		button.Selectable = true
		return button
	}
	return &scoped
}
