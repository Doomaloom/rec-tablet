package panels

import (
	"rectablet/src/shared/ui"
)

// Split keeps feature panels inside the shell's body rectangle.
func Split(surface *ui.Surface, leftFraction float32) (*ui.Surface, *ui.Surface) {
	left := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, y, (width - ui.Gap) * leftFraction, height
	})
	right := ui.Panel(surface, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		leftWidth := (width - ui.Gap) * leftFraction
		return x + leftWidth + ui.Gap, y, width - leftWidth - ui.Gap, height
	})
	return left, right
}
