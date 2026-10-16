package shell

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/ui"
)

func TestDisplayProfileChangesLogicalSizeAndDensity(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("display profiles")
	defer w.Close()
	p := &compositionPresenter{state: State{Screen: "test", PageCount: 1}}
	p.state.ComposeBody = func(surface *ui.Surface) {
		ui.List(surface, []ui.Item{{ID: "one", Label: "One"}}, func(string) {})
	}
	v := New(p, w)
	defer v.Close()
	w.Show()
	compactHeight := v.Buttons["row:one"].Size().Height
	for _, mode := range []ui.DisplayMode{ui.DisplayLarge, ui.DisplayCompact} {
		p.state.DisplayMode = mode
		p.changed()
		if v.Root.MinSize() != mode.Size() || w.Canvas().Size() != mode.Size() {
			t.Fatalf("%s size: min %v canvas %v", mode, v.Root.MinSize(), w.Canvas().Size())
		}
		button := v.Buttons["row:one"]
		if button.ScaleFactor != 1 {
			t.Fatalf("%s scaled instead of recomposing", mode)
		}
		if mode.IsLarge() && button.Size().Height >= compactHeight {
			t.Fatal("large list did not increase density")
		}
	}
	p.state.DisplayMode = ui.DisplayLarge
	p.changed()
	w.Resize(fyne.NewSize(1536, 900))
	if scale := v.Buttons["row:one"].ScaleFactor; scale != 1.5 {
		t.Fatalf("large profile resize scale = %v", scale)
	}
}
