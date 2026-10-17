package field

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"testing"
)

func TestErrorReplacesHelper(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	f := New(widget.NewEntry(), Options{Label: "Name", Helper: "Use your first name"})
	f.SetError("Name is required")
	column := f.content().(*fyne.Container)
	last := column.Objects[len(column.Objects)-1].(*container.ThemeOverride).Content.(*widget.Label)
	if last.Text != "Name is required" {
		t.Fatalf("wrong field message %q", last.Text)
	}
	f.SetError("")
	column = f.content().(*fyne.Container)
	last = column.Objects[len(column.Objects)-1].(*container.ThemeOverride).Content.(*widget.Label)
	if last.Text != "Use your first name" {
		t.Fatal("clearing error did not restore helper")
	}
}
