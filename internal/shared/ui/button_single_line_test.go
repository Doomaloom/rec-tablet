package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

func TestSingleLineButtonEllipsizesAtFixedFont(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(AppTheme{})
	for _, scale := range []float32{1, 1.5} {
		button := NewButton(strings.Repeat("Long student name ", 8), nil)
		button.FontSize, button.Compact, button.SingleLine = 18, true, true
		button.SetScale(scale)
		button.Resize(fyne.NewSize(296*scale, 32*scale))
		button.Refresh()
		renderer := test.WidgetRenderer(button).(*buttonRenderer)
		label := renderer.label
		if label.Wrap || label.MaxLines != 1 || !label.Truncated {
			t.Fatal("overflow must ellipsize without wrapping")
		}
		lines := test.WidgetRenderer(label).Objects()
		if len(lines) != 1 {
			t.Fatalf("single-line button rendered %d lines", len(lines))
		}
		line := lines[0].(*canvas.Text)
		if !strings.HasSuffix(line.Text, "…") || line.TextSize != 18*scale {
			t.Fatal("overflow should retain 18pt text and end with an ellipsis")
		}
	}
}

func TestScaledWrappedTextKeepsLogicalBreaks(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	app.Settings().SetTheme(AppTheme{})
	button := NewButton("Both legs kick at same time and recover at the same time", nil)
	button.FontSize, button.Compact, button.MaxLabelLines = 22, true, -1
	button.SetScale(1.6)
	button.Resize(fyne.NewSize(656*1.6, 59.5*1.6))
	button.Refresh()
	label := test.WidgetRenderer(button).(*buttonRenderer).label
	if label.Truncated {
		t.Fatal("fractional display scaling changed a fitting logical line into a truncated line")
	}
	lines := test.WidgetRenderer(label).Objects()
	if len(lines) != 1 || lines[0].(*canvas.Text).Text != button.Text {
		t.Fatalf("scaled text wrapping changed unexpectedly: %+v", lines)
	}
}
