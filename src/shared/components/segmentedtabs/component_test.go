package segmentedtabs

import (
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestSelectionAndDisabledOptions(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	calls := 0
	tabs := New([]Item{{Key: "a", Label: "One"}, {Key: "b", Label: "Two"}, {Key: "c", Label: "Locked", Disabled: true}}, Options{ActiveKey: "a", OnChanged: func(key string) {
		calls++
		if key != "b" {
			t.Errorf("unexpected key %q", key)
		}
	}})
	w := test.NewWindow(tabs)
	defer w.Close()
	test.Tap(tabs.buttons[1])
	test.Tap(tabs.buttons[1])
	test.Tap(tabs.buttons[2])
	if tabs.ActiveKey() != "b" || calls != 1 {
		t.Fatalf("key=%q, calls=%d", tabs.ActiveKey(), calls)
	}
	tabs.SetActive("a")
	tabs.SetActive("missing")
	tabs.SetActive("c")
	if tabs.ActiveKey() != "a" || calls != 1 {
		t.Fatal("programmatic selection changed callback or accepted invalid key")
	}
}
