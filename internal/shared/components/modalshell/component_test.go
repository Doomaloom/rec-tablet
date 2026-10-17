package modalshell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"rectablet/internal/shared/components/design"
	"rectablet/internal/shared/components/textinput"
	"testing"
)

func TestDismissalFocusAndNestedLifecycle(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	launcher := widget.NewButton("Open", nil)
	w := test.NewWindow(launcher)
	defer w.Close()
	w.Resize(fyne.NewSize(800, 600))
	w.Canvas().Focus(launcher)
	calls := 0
	input := textinput.New(textinput.Options{})
	m := New(w, input.Object(), Options{Title: "Edit", OnClose: func() { calls++ }})
	for _, method := range []string{"button", "escape", "backdrop"} {
		m.Show()
		m.Show()
		if len(w.Canvas().Overlays().List()) != 1 {
			t.Fatal("duplicate modal overlay")
		}
		switch method {
		case "button":
			test.Tap(m.close)
		case "escape":
			w.Canvas().Focus(input)
			if !design.HandleEscape(input.Entry, &fyne.KeyEvent{Name: fyne.KeyEscape}) {
				t.Fatal("focused input did not handle Escape")
			}
		case "backdrop":
			test.TapAt(m, fyne.NewPos(0, 0))
		}
		if m.Visible() || len(w.Canvas().Overlays().List()) != 0 {
			t.Fatalf("%s did not dismiss", method)
		}
		m.Hide()
		if w.Canvas().Focused() != launcher {
			t.Fatal("launcher focus was not restored")
		}
	}
	if calls != 3 {
		t.Fatalf("close callbacks=%d", calls)
	}
	m.Show()
	nested := New(w, widget.NewLabel("Nested"), Options{Title: "Nested"})
	nested.Show()
	m.Hide()
	if nested.Visible() || len(w.Canvas().Overlays().List()) != 0 {
		t.Fatal("closing parent left a child modal")
	}
}
func TestNoFocusEscapeAndPanelTap(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(widget.NewLabel("Background"))
	defer w.Close()
	w.Resize(fyne.NewSize(500, 600))
	oldCalls := 0
	w.Canvas().SetOnTypedKey(func(*fyne.KeyEvent) { oldCalls++ })
	m := New(w, widget.NewLabel("Content"), Options{Title: "Test"})
	m.Show()
	test.TapAt(m, fyne.NewPos(m.panelPos.X+10, m.panelPos.Y+10))
	if !m.Visible() {
		t.Fatal("panel tap dismissed modal")
	}
	w.Canvas().Unfocus()
	w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if m.Visible() {
		t.Fatal("unfocused Escape did not dismiss")
	}
	w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyA})
	if oldCalls != 1 {
		t.Fatal("original key handler not restored")
	}
}
