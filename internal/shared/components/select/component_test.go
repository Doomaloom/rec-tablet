package selectcontrol

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestNativeSelectionAndPopup(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	changed := ""
	s := New(Options{Options: []string{"One", "Two"}, Value: "One", OnChanged: func(value string) { changed = value }})
	w := test.NewWindow(s.Object())
	defer w.Close()
	w.Resize(fyne.NewSize(300, 100))
	s.control.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if changed != "Two" || s.Selected != "Two" {
		t.Fatal("native keyboard selection did not change")
	}
	test.Tap(s.control)
	if len(w.Canvas().Overlays().List()) != 1 {
		t.Fatal("dropdown did not open")
	}
	w.Canvas().Overlays().Top().Hide()
	s.Disable()
	test.Tap(s.control)
	if len(w.Canvas().Overlays().List()) != 0 {
		t.Fatal("disabled dropdown opened")
	}
}
