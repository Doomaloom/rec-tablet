package gallery

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/components/design"
)

func TestGalleryInteractionsAndRendering(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(design.Theme{})
	w := a.NewWindow("gallery")
	defer w.Close()
	g := &gallery{window: w}
	content := g.build()
	w.SetContent(content)
	w.Resize(fyne.NewSize(1000, 800))
	w.Show()
	g.save()
	test.Type(g.name.Entry, "Alex")
	g.save()
	if g.name.Text != "Alex" {
		t.Fatalf("input lost its value: %q", g.name.Text)
	}
	test.Tap(g.choiceB)
	if g.choiceA.Selected() || !g.choiceB.Selected() {
		t.Fatal("card selection did not change")
	}
	w.Resize(fyne.NewSize(1000, 800))
	g.openModal()
	if len(w.Canvas().Overlays().List()) != 1 {
		t.Fatal("modal did not open")
	}
	capture(t, w, "gallery-modal.png")
	g.modal.Hide()
	if len(w.Canvas().Overlays().List()) != 0 {
		t.Fatal("modal did not close")
	}
}

func TestGalleryResponsiveRendering(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(design.Theme{})
	for _, viewport := range []struct {
		name  string
		width float32
	}{{"desktop", 1000}, {"narrow", 360}} {
		t.Run(viewport.name, func(t *testing.T) {
			w := a.NewWindow("gallery")
			defer w.Close()
			content := New(w)
			w.SetContent(content)
			w.Resize(fyne.NewSize(viewport.width, 800))
			w.Show()
			capture(t, w, "gallery-"+viewport.name+".png")
			scroll := content.(*container.ThemeOverride).Content.(*container.Scroll)
			page := scroll.Content
			// Replace the scroll viewport with its content in the same window for a full-page render.
			w.SetContent(design.Scope(page))
			for i := 0; i < 4; i++ {
				w.Resize(fyne.NewSize(viewport.width, page.MinSize().Height))
				page.Refresh()
			}
			sections := page.(*fyne.Container).Objects[0].(*fyne.Container).Objects
			previousBottom := float32(0)
			for i, section := range sections {
				if section.Position().Y < previousBottom {
					t.Fatalf("section %d overlaps the previous section", i)
				}
				if section.Size().Height+0.5 < section.MinSize().Height {
					t.Fatalf("section %d clips its content: height %.1f, minimum %.1f", i, section.Size().Height, section.MinSize().Height)
				}
				previousBottom = section.Position().Y + section.Size().Height
			}
			capture(t, w, "gallery-full-"+viewport.name+".png")
		})
	}
}

func capture(t *testing.T, w fyne.Window, name string) {
	t.Helper()
	dir := os.Getenv("REC_UI_SCREENSHOTS")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
