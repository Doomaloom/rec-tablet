package ui

// Shared dimensions in the selected display profile's logical design space.
const (
	OuterMargin   float32 = 4
	Gap           float32 = 4
	RailWidth     float32 = 52
	ContentInset          = OuterMargin + RailWidth + Gap
	ButtonPadding float32 = 8
	HeaderHeight  float32 = 44
)

type Bounds func(width, height float32) (x, y, bodyWidth, bodyHeight float32)

// BodyBounds is also usable by isolated feature views without a shell.
func (surface *Surface) BodyBounds(width, height float32) (float32, float32, float32, float32) {
	if surface.Bounds != nil {
		return surface.Bounds(width, height)
	}
	top := OuterMargin + HeaderHeight + Gap
	return ContentInset, top, width - 2*ContentInset, height - top - OuterMargin
}
