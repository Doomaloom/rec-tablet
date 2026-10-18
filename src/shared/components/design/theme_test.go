package design

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"testing"
)

func TestThemeScopeAndFontStyles(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	original := a.Settings().Theme()
	inside := widget.NewLabel("Scoped")
	scope := Scope(inside)
	w := test.NewWindow(scope)
	defer w.Close()
	_ = test.WidgetRenderer(inside)
	if a.Settings().Theme() != original {
		t.Fatal("scoped theme changed application settings")
	}
	th := Theme{}
	seen := map[string]bool{}
	for _, s := range []fyne.TextStyle{{}, {Bold: true}, {Italic: true}, {Bold: true, Italic: true}} {
		font := th.Font(s)
		if len(font.Content()) == 0 || seen[font.Name()] {
			t.Fatal("missing or duplicate font style")
		}
		seen[font.Name()] = true
	}
	if th.Color(theme.ColorNamePrimary, theme.VariantDark) != Primary {
		t.Fatal("palette changes in dark mode")
	}
}
