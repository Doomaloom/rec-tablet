// Package shell hosts the shared native touchscreen composition.
package shell

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/ui"
)

var homeIcon = fyne.NewStaticResource("home.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path fill="#315daf" d="M3 29L32 3l29 26H51v32H13V29z"/><path fill="#e1e9f4" d="M27 41h11v20H27z"/></svg>`))

func arrowIcon(up, disabled bool) fyne.Resource {
	fill := "#315daf"
	if disabled {
		fill = "#9aacc5"
	}
	transform := ""
	if !up {
		transform = ` transform="rotate(180 32 48)"`
	}
	return fyne.NewStaticResource(fmt.Sprintf("arrow-%t-%t.svg", up, disabled), []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="96" viewBox="0 0 64 96"><path fill="%s"%s d="M32 2L2 34H20V94H44V34H62Z"/></svg>`, fill, transform)))
}

type View struct {
	Controller Presenter
	Window     fyne.Window
	Root       *fyne.Container
	Buttons    map[string]*ui.TouchButton
	Criterion  *ui.TextBlock
	lastScreen string
	readyAt    time.Time
	timer      *time.Timer
	lastNotice string
	closed     bool
	onLeave    func()
}

type placement struct {
	object fyne.CanvasObject
	rect   LayoutSlot
}
type surfaceLayout struct{ items []placement }

func (layout *surfaceLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.NewSize(800, 480) }
func (layout *surfaceLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	scale := min(size.Width/800, size.Height/480)
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
	view := &View{Controller: controller, Window: window, Buttons: map[string]*ui.TouchButton{}}
	window.SetPadded(false)
	view.Root = container.New(&surfaceLayout{})
	window.SetContent(view.Root)
	window.Resize(fyne.NewSize(800, 480))
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
	if view.timer != nil {
		view.timer.Stop()
	}
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
	view.Root.Layout = &surfaceLayout{items: placements}
	view.Root.Objects = objects
	view.Root.Refresh()
	view.updateNoticeTimer(state)
}

func (view *View) updateNoticeTimer(state State) {
	if state.Notice == view.lastNotice {
		return
	}
	if view.timer != nil {
		view.timer.Stop()
	}
	view.lastNotice = state.Notice
	if state.Notice == "" || state.Error {
		return
	}
	noticeMessage := state.Notice
	view.timer = time.AfterFunc(2200*time.Millisecond, func() {
		fyne.Do(func() {
			currentState := view.Controller.Snapshot()
			if !view.closed && !currentState.Error && currentState.Notice == noticeMessage {
				view.Controller.Dismiss()
			}
		})
	})
}

func (view *View) details() {
	state := view.Controller.Snapshot()
	detailsText := state.Context
	if detailsText == "" {
		detailsText = state.Subtitle
	}
	if detailsText == "" {
		detailsText = state.Title
	}
	label := widget.NewLabel(detailsText)
	label.Wrapping = fyne.TextWrapWord
	detailsDialog := dialog.NewCustom(state.DetailsTitle, "Close", container.NewVScroll(label), view.Window)
	detailsDialog.Resize(fyne.NewSize(650, 380))
	detailsDialog.Show()
}
