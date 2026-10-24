package rightrail

import (
	"fmt"

	"fyne.io/fyne/v2/canvas"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/app/internal/icons"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func renderActionRailComponents(c *contracts.RenderContext) {
	for index, action := range c.State.RightRailActions {
		button := c.Button(action.ID, action.Label, func() {
			if action.Run != nil {
				_ = action.Run()
			}
		})
		button.Fill, button.Ink, button.FontSize = action.Fill, action.Ink, action.FontSize
		if button.Fill == nil {
			button.Fill = ui.Rail
		}
		if button.Ink == nil {
			button.Ink = ui.Blue
		}
		if button.FontSize <= 0 {
			button.FontSize = 32
		}
		button.Icon, button.IconOnly = action.Icon, action.Icon != nil
		c.Add(fmt.Sprintf("right-rail-%s", action.ID), button, func(width, height float32) (float32, float32, float32, float32) {
			_, y, w, h := geometry.RailSlot(height, index, len(c.State.RightRailActions))
			return width - ui.OuterMargin - ui.RailWidth, y, w, h
		})
	}
}

func renderPagingRailComponents(builder *contracts.RenderContext) {
	state := builder.State
	previousButton := builder.Button("up", "↑", func() { _ = builder.Controller.Page(-1) })
	previousButton.Fill = ui.NavigationRailFill
	previousButton.Ink = ui.Blue
	previousButton.FontSize = 58
	previousButton.Disabled = state.Page == 0
	previousButton.IconOnly = true
	previousButton.Icon = icons.Chevron("up", previousButton.Disabled)
	builder.Add("previous-page", previousButton, geometry.PreviousPageSlot)

	nextButton := builder.Button("down", "↓", func() { _ = builder.Controller.Page(1) })
	nextButton.Fill = ui.NavigationRailFill
	nextButton.Ink = ui.Blue
	nextButton.FontSize = 58
	nextButton.Disabled = state.Page+1 >= state.PageCount
	nextButton.IconOnly = true
	nextButton.Icon = icons.Chevron("down", nextButton.Disabled)
	builder.Add("next-page", nextButton, geometry.NextPageSlot)
}

func renderPagePositionIndicatorComponents(builder *contracts.RenderContext) {
	if builder.State.PageCount <= 1 {
		return
	}
	track := canvas.NewRectangle(ui.Rail)
	track.CornerRadius = geometry.PagePositionTrackWidth / 2
	track.StrokeColor = ui.Muted
	track.StrokeWidth = 1
	builder.Add("page-position-track", track, geometry.PagePositionTrackSlot(builder.State))

	thumb := canvas.NewRectangle(ui.Blue)
	thumb.CornerRadius = geometry.PagePositionTrackWidth / 2
	thumb.StrokeColor = ui.Navy
	thumb.StrokeWidth = 1
	builder.Add("page-position-thumb", thumb, geometry.PagePositionThumbSlot(builder.State))
}
