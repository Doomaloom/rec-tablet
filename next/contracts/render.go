package contracts

import (
	"fmt"

	"fyne.io/fyne/v2"

	"rectablet/src/shared/ui"
)

type LayoutSlot func(width, height float32) (x, y, widthOut, heightOut float32)
type LayoutComponent struct {
	Name     string
	Object   fyne.CanvasObject
	Position LayoutSlot
}
type ScreenComposition struct {
	Components []LayoutComponent
	Buttons    map[string]*ui.TouchButton
	Criterion  *ui.TextBlock
}

// RenderContext collects one render. Controllers and feature state stay outside it.
type RenderContext struct {
	Controller Presenter
	State      State
	Window     fyne.Window
	Ready      func() bool
	Components []LayoutComponent
	Buttons    map[string]*ui.TouchButton
	Criterion  *ui.TextBlock
}

func (c *RenderContext) Add(name string, object fyne.CanvasObject, position LayoutSlot) {
	c.Components = append(c.Components, LayoutComponent{Name: name, Object: object, Position: position})
}
func (c *RenderContext) AddBody(object fyne.CanvasObject, position LayoutSlot) {
	c.Add(fmt.Sprintf("feature-body-element-%d", len(c.Components)), object, position)
}
func (c *RenderContext) Button(id, label string, callback func()) *ui.TouchButton {
	b := ui.NewButton(label, callback)
	b.CanActivate = c.Ready
	c.Buttons[id] = b
	return b
}
