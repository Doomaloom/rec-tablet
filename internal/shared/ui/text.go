package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strings"
	"unicode/utf8"
)

// TextBlock wraps at physical layout width, with bounded fitting for long criteria.
type TextBlock struct {
	widget.BaseWidget
	Text                               string
	FontSize, MinimumFont, ScaleFactor float32
	Color                              color.Color
	Bold, Center, Wrap                 bool
	MaxLines                           int
	Truncated                          bool
}

func NewText(text string, size float32, ink color.Color) *TextBlock {
	t := &TextBlock{Text: text, FontSize: size, MinimumFont: size, ScaleFactor: 1, Color: ink, Wrap: true}
	t.ExtendBaseWidget(t)
	return t
}
func (t *TextBlock) SetScale(scale float32) {
	if t.ScaleFactor != scale {
		t.ScaleFactor = scale
		t.Refresh()
	}
}
func (t *TextBlock) MinSize() fyne.Size                  { return fyne.NewSize(1, 1) }
func (t *TextBlock) CreateRenderer() fyne.WidgetRenderer { return &textRenderer{t: t} }

type textRenderer struct {
	t       *TextBlock
	objects []fyne.CanvasObject
}

func (r *textRenderer) MinSize() fyne.Size           { return fyne.NewSize(1, 1) }
func (r *textRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *textRenderer) Destroy()                     {}
func (r *textRenderer) Refresh()                     { r.Layout(r.t.Size()); canvas.Refresh(r.t) }

// WrappedTextHeight measures the same wrapping and line spacing used by TextBlock.
func WrappedTextHeight(text string, width, size float32, bold bool) float32 {
	lines := wrap(text, max(1, width), size, bold, true)
	return float32(len(lines)) * fyne.MeasureText("Ag", size, fyne.TextStyle{Bold: bold}).Height * 1.12
}

func wrap(text string, width, size float32, bold, wrapping bool) []string {
	return wrapMeasured(text, width, size, bold, wrapping, fyne.MeasureText)
}

// wrapScaled measures at the logical font size before applying the canvas scale.
// Fyne's text metrics quantize fractional physical font sizes, which can move a
// borderline word to another line even though both the font and available width
// were scaled by the same factor.
func wrapScaled(text string, width, size, scale float32, bold, wrapping bool) []string {
	if scale <= 0 {
		scale = 1
	}
	measure := func(text string, physicalSize float32, style fyne.TextStyle) fyne.Size {
		logical := fyne.MeasureText(text, physicalSize/scale, style)
		return fyne.NewSize(logical.Width*scale, logical.Height*scale)
	}
	return wrapMeasured(text, width, size, bold, wrapping, measure)
}

func wrapMeasured(text string, width, size float32, bold, wrapping bool, measure func(string, float32, fyne.TextStyle) fyne.Size) []string {
	lines := []string{}
	style := fyne.TextStyle{Bold: bold}
	for _, paragraph := range strings.Split(text, "\n") {
		if !wrapping {
			lines = append(lines, paragraph)
			continue
		}
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if line != "" && measure(candidate, size, style).Width > width {
				lines = append(lines, line)
				line = ""
			}
			if line != "" {
				line += " "
			}
			for _, r := range word {
				candidate = line + string(r)
				if line != "" && measure(candidate, size, style).Width > width {
					lines = append(lines, line)
					line = ""
				}
				line += string(r)
			}
		}
		lines = append(lines, line)
	}
	return lines
}
func (r *textRenderer) Layout(bounds fyne.Size) {
	t := r.t
	size := t.FontSize * t.ScaleFactor
	minimum := t.MinimumFont * t.ScaleFactor
	scale := max(float32(0.01), t.ScaleFactor)
	var lines []string
	var lineHeight float32
	for {
		lines = wrapScaled(t.Text, max(1, bounds.Width), size, scale, t.Bold, t.Wrap)
		lineHeight = fyne.MeasureText("Ag", size/scale, fyne.TextStyle{Bold: t.Bold}).Height * scale * 1.12
		if float32(len(lines))*lineHeight <= bounds.Height || size <= minimum {
			break
		}
		size = max(minimum, size-t.ScaleFactor)
	}
	maxLines := max(1, int(bounds.Height/max(1, lineHeight)))
	if t.MaxLines > 0 {
		maxLines = min(maxLines, t.MaxLines)
	}
	t.Truncated = len(lines) > maxLines
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		tooWide := fyne.MeasureText(line, size/scale, fyne.TextStyle{Bold: t.Bold}).Width*scale > bounds.Width
		if tooWide || (t.Truncated && i == len(lines)-1) {
			for len(line) > 0 && fyne.MeasureText(line+"…", size/scale, fyne.TextStyle{Bold: t.Bold}).Width*scale > bounds.Width {
				_, n := utf8.DecodeLastRuneInString(line)
				line = line[:len(line)-n]
			}
			lines[i] = line + "…"
			t.Truncated = true
		}
	}
	objects := make([]fyne.CanvasObject, 0, len(lines))
	y := max(0, (bounds.Height-float32(len(lines))*lineHeight)/2)
	for _, line := range lines {
		label := canvas.NewText(line, t.Color)
		label.TextSize = size
		label.TextStyle.Bold = t.Bold
		if t.Center {
			label.Alignment = fyne.TextAlignCenter
		}
		label.Move(fyne.NewPos(0, y))
		label.Resize(fyne.NewSize(bounds.Width, lineHeight))
		objects = append(objects, label)
		y += lineHeight
	}
	r.objects = objects
}
