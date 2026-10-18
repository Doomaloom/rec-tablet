package actionbutton

import (
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"rectablet/src/shared/components/design"
)

// SurfaceStyle lets specialized touch controls share the button surface while
// retaining their own content layout, gestures, and logical display scaling.
type SurfaceStyle struct {
	Fill, Ink                                      color.Color
	Selected, Outlined, Focused, Hovered, Disabled bool
	Selectable                                     bool
	Scale                                          float32
}

// PaintSurface applies the library's rounded border and interaction treatments.
// It returns the foreground color for content rendered above the surface.
func PaintSurface(bg *canvas.Rectangle, style SurfaceStyle) color.Color {
	scale := style.Scale
	if scale <= 0 {
		scale = 1
	}
	fill, ink := style.Fill, style.Ink
	if fill == nil {
		fill = design.Secondary
	}
	if ink == nil {
		ink = design.White
	}
	bg.FillColor, bg.StrokeWidth = fill, 0
	bg.CornerRadius = design.ControlRadius * scale
	if style.Hovered {
		bg.FillColor = design.Opacity(fill, 225)
	}
	if style.Selected || style.Outlined {
		bg.StrokeColor, bg.StrokeWidth = design.Secondary, 2*scale
	}
	if style.Selectable {
		bg.FillColor = design.White
		ink = design.Secondary
		bg.StrokeColor, bg.StrokeWidth = design.Alpha(design.Secondary, 51), scale
		if style.Hovered {
			bg.StrokeColor = design.Secondary
		}
		if style.Selected {
			bg.FillColor, ink = design.Secondary, design.White
			bg.StrokeColor = design.Secondary
		}
	}
	if style.Focused {
		bg.StrokeColor, bg.StrokeWidth = design.Primary, 3*scale
	}
	if style.Disabled {
		bg.FillColor = design.Alpha(design.Secondary, 25)
		ink = design.Alpha(design.Secondary, 153)
	}
	return ink
}
