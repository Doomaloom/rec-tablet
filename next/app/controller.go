// Package app hosts the persistent native touchscreen layout.
package app

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

type View struct {
	Controller  contracts.Presenter
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
	rect   contracts.LayoutSlot
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

func New(controller contracts.Presenter, window fyne.Window) *View {
	mode := controller.Snapshot().DisplayMode.Normalize()
	view := &View{Controller: controller, Window: window, Buttons: map[string]*ui.TouchButton{}, displayMode: mode}
	window.SetPadded(false)
	view.Root = container.New(&surfaceLayout{mode: mode})
	window.SetContent(view.Root)
	window.Resize(mode.Size())
	controller.Subscribe(view.Render)
	if edits, ok := controller.(interface {
		SubscribeEdits(func(contracts.ButtonUpdate))
	}); ok {
		edits.SubscribeEdits(view.applyButtonUpdate)
	}
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, func(fyne.Shortcut) {
		if !view.closed {
			_ = controller.OpenSetup()
		}
	})
	window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}, func(fyne.Shortcut) {
		if !view.closed {
			_ = controller.Back()
		}
	})
	window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		if view.closed {
			return
		}
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
	if controller, ok := view.Controller.(interface{ Deactivate() }); ok {
		controller.Deactivate()
	}
	// Hide through each overlay's lifecycle so modal key handlers and focus are
	// restored before a replacement view installs its own handlers.
	for _, overlay := range view.Window.Canvas().Overlays().List() {
		if modal, ok := overlay.(interface{ Hide() }); ok {
			modal.Hide()
		}
	}
	view.Window.Canvas().RemoveShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift})
	view.Window.Canvas().RemoveShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt})
	view.Window.Canvas().SetOnTypedKey(nil)
	view.Controller.Subscribe(nil)
	if edits, ok := view.Controller.(interface {
		SubscribeEdits(func(contracts.ButtonUpdate))
	}); ok {
		edits.SubscribeEdits(nil)
	}
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
func (view *View) Ready() bool { return !view.closed && time.Now().After(view.readyAt) }

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

func (view *View) applyButtonUpdate(update contracts.ButtonUpdate) {
	if view.closed {
		return
	}
	if button := view.Buttons[update.ID]; button != nil {
		button.Text, button.Fill, button.Ink, button.Selected = update.Text, update.Fill, update.Ink, update.Selected
		button.Refresh()
	}
}
