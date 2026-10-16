package ui

import "fyne.io/fyne/v2"

// DisplayMode selects a layout independently of the window's current size.
// The zero value preserves the original five-inch layout.
type DisplayMode string

const (
	DisplayCompact DisplayMode = "compact"
	DisplayLarge   DisplayMode = "large"
)

func (mode DisplayMode) IsLarge() bool { return mode == DisplayLarge }

func (mode DisplayMode) Normalize() DisplayMode {
	if mode.IsLarge() {
		return DisplayLarge
	}
	return DisplayCompact
}

func (mode DisplayMode) Size() fyne.Size {
	if mode.IsLarge() {
		return fyne.NewSize(1024, 600)
	}
	return fyne.NewSize(800, 480)
}

func (mode DisplayMode) Label() string {
	if mode.IsLarge() {
		return "10.1-inch · 1024×600"
	}
	return "5-inch · 800×480"
}

// Panel shares the parent's component registration while replacing body bounds.
func Panel(surface *Surface, bounds Bounds) *Surface {
	panel := *surface
	panel.Bounds = bounds
	return &panel
}
