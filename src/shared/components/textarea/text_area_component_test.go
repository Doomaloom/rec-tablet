package textarea

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestMultilineEditing(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	value := ""
	c := New(Options{OnChanged: func(text string) { value = text }})
	w := test.NewWindow(c.Object())
	defer w.Close()
	test.Type(c.control, "First")
	c.control.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	test.Type(c.control, "Second")
	if value != "First\nSecond" {
		t.Fatalf("multiline value=%q", value)
	}
	if c.Object().MinSize().Height < 120 {
		t.Fatal("textarea too short")
	}
}
