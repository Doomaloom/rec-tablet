package leftrail

import (
	"image/color"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/app/internal/icons"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func renderActionRailComponents(builder *contracts.RenderContext) {
	count := len(builder.State.LeftRailActions)
	for i, action := range builder.State.LeftRailActions {
		index, action := i, action
		button := builder.Button(action.ID, action.Label, func() { _ = action.Run() })
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
		builder.Add(action.ID, button, func(_, height float32) (float32, float32, float32, float32) {
			return geometry.RailSlot(height, index, count)
		})
	}
}

func renderHomeRailComponent(builder *contracts.RenderContext) {
	action := builder.State.HomeRailAction
	label := "Home"
	callback := func() { _ = builder.Controller.Home() }
	var fill, ink color.Color = ui.NavigationRailFill, ui.Blue
	fontSize := float32(17)
	if action != nil {
		label = action.Label
		callback = func() { _ = action.Run() }
		fill, ink, fontSize = action.Fill, action.Ink, action.FontSize
		if fill == nil {
			fill = ui.NavigationRailFill
		}
		if ink == nil {
			ink = ui.Blue
		}
		if fontSize <= 0 {
			fontSize = 32
		}
	}
	homeButton := builder.Button("home", label, callback)
	homeButton.Fill = fill
	homeButton.Ink = ink
	homeButton.FontSize = fontSize
	if action == nil {
		homeButton.Icon = ui.RailIcon("home", false)
		homeButton.IconOnly = true
	} else if action.Icon != nil {
		homeButton.Icon = action.Icon
		homeButton.IconOnly = true
	}
	homePosition := geometry.HomeRailSlot
	if builder.State.CanBack || builder.State.BackRailAction != nil {
		homePosition = geometry.SplitHomeRailSlot
	}
	builder.Add("home-rail", homeButton, homePosition)
}

func renderBackButtonComponent(builder *contracts.RenderContext) {
	if !builder.State.CanBack && builder.State.BackRailAction == nil {
		return
	}
	action := builder.State.BackRailAction
	label := "‹"
	callback := func() { _ = builder.Controller.Back() }
	var fill, ink color.Color = ui.NavigationRailFill, ui.Blue
	fontSize := float32(36)
	if action != nil {
		label = action.Label
		callback = func() { _ = action.Run() }
		fill, ink, fontSize = action.Fill, action.Ink, action.FontSize
		if fill == nil {
			fill = ui.NavigationRailFill
		}
		if ink == nil {
			ink = ui.Blue
		}
		if fontSize <= 0 {
			fontSize = 32
		}
	}
	backButton := builder.Button("back", label, callback)
	backButton.Fill = fill
	backButton.Ink = ink
	backButton.FontSize = fontSize
	if action == nil {
		backButton.Icon = icons.Chevron("left", false)
		backButton.IconOnly = true
	}
	if action != nil && action.Icon != nil {
		backButton.Icon = action.Icon
		backButton.IconOnly = true
	}
	builder.Add("back-button", backButton, geometry.SplitBackRailSlot)
}
