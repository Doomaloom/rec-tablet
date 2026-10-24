package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
)

// NavigationRailFill gives navigation controls a visible, quiet surface.
var NavigationRailFill = color.NRGBA{R: Navy.R, G: Navy.G, B: Navy.B, A: 25}

// RailIcon keeps navigation glyphs on the same canvas and stroke weight.
func RailIcon(name string, disabled bool) fyne.Resource {
	path := "M12 24L32 44L52 24"
	switch name {
	case "up":
		path = "M12 40L32 20L52 40"
	case "left":
		path = "M40 12L20 32L40 52"
	case "home":
		path = "M10 30L32 10L54 30M17 25V53H47V25M27 53V37H37V53"
	case "logout":
		path = "M28 12H12V52H28M34 20L46 32L34 44M22 32H46"
	}
	ink := Blue
	if disabled {
		ink = color.NRGBA{R: 102, G: 159, B: 189, A: 255}
	}
	return fyne.NewStaticResource(fmt.Sprintf("rail-%s-%t.svg", name, disabled), []byte(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="%s" fill="none" stroke="#%02x%02x%02x" stroke-width="7"/></svg>`,
		path, ink.R, ink.G, ink.B)))
}
