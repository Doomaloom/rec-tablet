package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestStudentSuccessChecksAlignRightAndLeaveRoomForName(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	for _, scale := range []float32{1, 1.5} {
		button := NewButton("Alex Smi", nil)
		button.Selectable = true
		button.SetScale(scale)
		button.Resize(fyne.NewSize(308*scale, 45*scale))
		for _, count := range []int{0, 1, 2, 3, 5, -1, 0} {
			button.SuccessChecks = count
			button.Refresh()
			renderer := test.WidgetRenderer(button).(*buttonRenderer)
			if renderer.label.Center {
				t.Fatal("student name should align left")
			}
			want := min(3, max(0, count))
			for i, check := range renderer.successChecks {
				if check.Visible() != (i < want) {
					t.Fatalf("count %d: check %d visibility = %v", count, i, check.Visible())
				}
				if i >= want {
					continue
				}
				if renderer.label.Position().X+renderer.label.Size().Width >= check.Position().X {
					t.Fatal("name overlaps checkmarks")
				}
				if i == want-1 && check.Position().X+check.Size().Width != button.Size().Width-ButtonPadding*scale {
					t.Fatal("checkmarks must align to the right padding")
				}
			}
		}
	}
}
