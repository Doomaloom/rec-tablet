package content

import (
	"fyne.io/fyne/v2"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func renderBodyComponent(builder *contracts.RenderContext) {
	if builder.State.ComposeBody == nil {
		return
	}
	surface := &ui.Surface{
		DisplayMode: builder.State.DisplayMode,
		Window:      builder.Window,
		Bounds:      geometry.BodySlot(builder.State),
		Add: func(object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32)) {
			builder.AddBody(object, position)
		},
		AddNamed: func(name string, object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32)) {
			builder.Add("feature-"+name, object, position)
		},
		Button: builder.Button,
	}
	builder.State.ComposeBody(surface)
	builder.Criterion = surface.Criterion
}
