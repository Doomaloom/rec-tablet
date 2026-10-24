package icons

import (
	"fyne.io/fyne/v2"
	"rectablet/src/shared/ui"
)

func Chevron(direction string, disabled bool) fyne.Resource {
	return ui.RailIcon(direction, disabled)
}
