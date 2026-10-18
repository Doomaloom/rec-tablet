package selectablecard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"rectablet/src/shared/components/design"
	"testing"
)

func TestSelectionPaintAndDisabledActivation(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	calls := 0
	c := New(widget.NewLabel("Afternoon"), Options{}, func() { calls++ })
	w := test.NewWindow(c)
	defer w.Close()
	w.Resize(fyne.NewSize(300, 64))
	c.SetSelected(true)
	r := test.WidgetRenderer(c.button)
	r.Layout(c.Size())
	bg := r.Objects()[1].(*canvas.Rectangle)
	if bg.FillColor != design.Secondary || bg.Size().Height == 0 {
		t.Fatal("selected card background was not painted")
	}
	c.button.TypedKey(&fyne.KeyEvent{Name: fyne.KeySpace})
	c.Disable()
	test.Tap(c)
	if calls != 1 {
		t.Fatal("disabled card activated or keyboard activation failed")
	}
}
