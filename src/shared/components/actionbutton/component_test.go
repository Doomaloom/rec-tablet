package actionbutton

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"rectablet/src/shared/components/design"
	"testing"
)

func TestActivationAndVisualFeedback(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	calls := 0
	b := New("Save", Options{Variant: Primary}, func() { calls++ })
	w := test.NewWindow(b)
	defer w.Close()
	w.Resize(fyne.NewSize(200, 56))
	test.Tap(b)
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeySpace})
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if calls != 3 {
		t.Fatalf("activation calls=%d", calls)
	}
	b.Disable()
	test.Tap(b)
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if calls != 3 {
		t.Fatal("disabled button activated")
	}
	b.Enable()
	w.Canvas().Focus(b)
	r := test.WidgetRenderer(b)
	if r.Objects()[0].(*canvas.Rectangle).StrokeWidth == 0 {
		t.Fatal("focus is invisible")
	}
	if b.MinSize().Height < design.Target {
		t.Fatal("touch target too small")
	}
}
func TestLongLabelFits(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	b := New("A long action label that must wrap at tablet width", Options{}, nil)
	w := test.NewWindow(b)
	defer w.Close()
	w.Resize(fyne.NewSize(220, 160))
	r := test.WidgetRenderer(b)
	r.Layout(fyne.NewSize(220, 160))
	if r.MinSize().Height > 160 {
		t.Fatal("long label exceeds available height")
	}
}

func TestDisabledTransparentVariants(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	for _, variant := range []Variant{Ghost, Outline} {
		b := New("Disabled", Options{Variant: variant, Disabled: true}, nil)
		background := test.WidgetRenderer(b).Objects()[1].(*canvas.Rectangle)
		_, _, _, alpha := background.FillColor.RGBA()
		if alpha != 0 {
			t.Fatalf("disabled %s gained an opaque background", variant)
		}
	}
}
