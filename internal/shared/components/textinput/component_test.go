package textinput

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/components/design"
	"testing"
)

func TestNativeEditingAndEscape(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	changed := ""
	input := New(Options{OnChanged: func(s string) { changed = s }})
	w := test.NewWindow(input.Object())
	defer w.Close()
	test.Type(input.control, "Alex")
	if changed != "Alex" || input.Text != "Alex" {
		t.Fatal("native editing callback failed")
	}
	closed := false
	design.SetEscapeHandler(w.Canvas(), func() { closed = true })
	defer design.SetEscapeHandler(w.Canvas(), nil)
	input.control.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if !closed {
		t.Fatal("focused entry swallowed Escape")
	}
	input.Disable()
	test.Type(input.control, "X")
	if input.Text != "Alex" {
		t.Fatal("disabled input accepted typing")
	}
	if input.Object().MinSize().Height < 44 {
		t.Fatal("input is smaller than touch target")
	}
}
