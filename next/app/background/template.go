package background

import (
	"fyne.io/fyne/v2/canvas"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func renderBackgroundComponent(builder *contracts.RenderContext) {
	background := canvas.NewRectangle(ui.Background)
	builder.Add("background", background, geometry.BackgroundSlot)
}
