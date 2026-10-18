// Package shell hosts the shared native touchscreen composition.
package shell

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"rectablet/src/shared/ui"
)

func chevronIcon(direction string, disabled bool) fyne.Resource {
	ink := "#007acc"
	if disabled {
		ink = "#669fbd"
	}
	path := "M12 24L32 44L52 24"
	switch direction {
	case "up":
		path = "M12 40L32 20L52 40"
	case "left":
		path = "M40 12L20 32L40 52"
	}
	return fyne.NewStaticResource(fmt.Sprintf("chevron-%s-%t.svg", direction, disabled), []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="%s" fill="none" stroke="%s" stroke-width="7"/></svg>`, path, ink)))
}

type View struct {
	Controller  Presenter
	Window      fyne.Window
	Root        *fyne.Container
	Buttons     map[string]*ui.TouchButton
	Criterion   *ui.TextBlock
	lastScreen  string
	readyAt     time.Time
	closed      bool
	displayMode ui.DisplayMode
	onLeave     func()
}

type placement struct {
	object fyne.CanvasObject
	rect   LayoutSlot
}
type surfaceLayout struct {
	items []placement
	mode  ui.DisplayMode
}

func (layout *surfaceLayout) MinSize([]fyne.CanvasObject) fyne.Size { return layout.mode.Size() }
func (layout *surfaceLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	base := layout.mode.Size()
	scale := max(float32(0.01), min(size.Width/base.Width, size.Height/base.Height))
	width, height := size.Width/scale, size.Height/scale
	for _, component := range layout.items {
		x, y, componentWidth, componentHeight := component.rect(width, height)
		if scalable, ok := component.object.(interface{ SetScale(float32) }); ok {
			scalable.SetScale(scale)
		}
		component.object.Move(fyne.NewPos(x*scale, y*scale))
		component.object.Resize(fyne.NewSize(max(1, componentWidth*scale), max(1, componentHeight*scale)))
	}
}

func New(controller Presenter, window fyne.Window) *View {
	mode := controller.Snapshot().DisplayMode.Normalize()
	view := &View{Controller: controller, Window: window, Buttons: map[string]*ui.TouchButton{}, displayMode: mode}
	window.SetPadded(false)
	view.Root = container.New(&surfaceLayout{mode: mode})
	window.SetContent(view.Root)
	window.Resize(mode.Size())
	controller.Subscribe(view.Render)
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, func(fyne.Shortcut) { _ = controller.OpenSetup() })
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}, func(fyne.Shortcut) { _ = controller.Back() })
	window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyF11:
			window.SetFullScreen(!window.FullScreen())
		case fyne.KeyEscape:
			_ = controller.Back()
		}
	})
	view.Render()
	return view
}

func (view *View) Close() {
	if view.closed {
		return
	}
	view.closed = true
	view.Controller.Subscribe(nil)
	if view.onLeave != nil {
		view.onLeave()
		view.onLeave = nil
	}
	view.releaseButtons()
}
func (view *View) releaseButtons() {
	for _, button := range view.Buttons {
		button.Release()
		button.CanActivate = func() bool { return false }
	}
}
func (view *View) Ready() bool { return time.Now().After(view.readyAt) }

func (view *View) Render() {
	if view.closed {
		return
	}
	state := view.Controller.Snapshot()
	if state.Screen != view.lastScreen {
		if view.onLeave != nil {
			view.onLeave()
		}
		view.lastScreen = state.Screen
		view.readyAt = time.Now().Add(120 * time.Millisecond)
	}
	view.onLeave = state.OnLeave
	view.releaseButtons()
	view.Criterion = nil
	composition := ComposeScreen(view, state)
	view.Buttons = composition.Buttons
	view.Criterion = composition.Criterion
	objects := make([]fyne.CanvasObject, 0, len(composition.Components))
	placements := make([]placement, 0, len(composition.Components))
	for _, component := range composition.Components {
		objects = append(objects, component.Object)
		placements = append(placements, placement{object: component.Object, rect: component.Position})
	}
	view.Root.Layout = &surfaceLayout{items: placements, mode: state.DisplayMode}
	view.Root.Objects = objects
	view.Root.Refresh()
	mode := state.DisplayMode.Normalize()
	if mode != view.displayMode {
		view.displayMode = mode
		if !view.Window.FullScreen() {
			view.Window.Resize(mode.Size())
		}
	}
}
