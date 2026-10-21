package header

import (
	"fmt"

	"fyne.io/fyne/v2/container"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/contracts"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
)

func renderHeaderComponents(builder *contracts.RenderContext) {
	state := builder.State
	if state.SplitHeader != nil {
		renderSplitHeaderComponents(builder)
		return
	}
	actionWidth := float32(0)
	actions := state.HeaderActions
	if len(actions) > 0 {
		actionWidth = float32(len(actions))*98 + float32(len(actions)-1)*ui.Gap
		for i, action := range actions {
			index, action := i, action
			button := builder.Button(fmt.Sprintf("header-action-%d", i), action.Label, func() { _ = action.Run() })
			button.Fill, button.Ink, button.FontSize = action.Fill, action.Ink, action.FontSize
			if button.Fill == nil {
				button.Fill = ui.Rail
			}
			if button.Ink == nil {
				button.Ink = ui.Blue
			}
			if button.FontSize <= 0 {
				button.FontSize = 18
			}
			position := func(width, height float32) (float32, float32, float32, float32) {
				x := width - ui.ContentInset - actionWidth + float32(index)*(98+ui.Gap)
				return x, ui.OuterMargin, 98, geometry.HeaderHeight(state)
			}
			builder.Add(fmt.Sprintf("header-action-%d", i), button, position)
		}
	} else if state.HeaderAction != "" {
		actionWidth = 98
		if state.HeaderWidth > 0 {
			actionWidth = state.HeaderWidth
		}
		headerAction := builder.Button("header", state.HeaderAction, func() { _ = builder.Controller.Header() })
		headerAction.FontSize = 18
		position := func(width, height float32) (float32, float32, float32, float32) {
			return geometry.HeaderActionSlot(width, height, actionWidth)
		}
		builder.Add("header-action", headerAction, position)
	}

	if len(state.HeaderSwitch) > 0 {
		const optionWidth = float32(92)
		const optionGap = ui.Gap
		switchWidth := float32(len(state.HeaderSwitch))*optionWidth + float32(len(state.HeaderSwitch)-1)*optionGap
		for i, option := range state.HeaderSwitch {
			index, option := i, option
			button := builder.Button(fmt.Sprintf("header-switch-%d", i), option.Label, func() { _ = option.Run() })
			button.FontSize = 18
			button.Fill, button.Ink = ui.White, ui.Navy
			if option.Selected {
				button.Fill, button.Ink = ui.Blue, ui.White
				button.Selected = true
			}
			position := func(width, _ float32) (float32, float32, float32, float32) {
				return width - ui.ContentInset - switchWidth + float32(index)*(optionWidth+optionGap), ui.OuterMargin, optionWidth, ui.HeaderHeight
			}
			builder.Add(fmt.Sprintf("header-switch-%d", i), button, position)
		}
		actionWidth = switchWidth
	}

	left := ui.ContentInset
	fontSize := float32(25)
	if state.HeaderTitleFontSize > 0 {
		fontSize = state.HeaderTitleFontSize
	}
	position := func(width, height float32) (float32, float32, float32, float32) {
		x, y, w, _ := geometry.HeaderTitleSlot(width, height, left, actionWidth)
		return x, y, w, geometry.HeaderHeight(state)
	}
	if state.HoldTitle != nil {
		titleButton := builder.Button("title", state.Title, func() {})
		titleButton.Fill, titleButton.Ink, titleButton.FontSize = ui.White, ui.Navy, fontSize
		titleButton.OnHold = state.HoldTitle
		if state.FitHeader {
			titleButton.Compact, titleButton.MinimumFont, titleButton.MaxLabelLines = true, 10, -1
		}
		builder.Add("screen-title", titleButton, position)
		return
	}
	background := card.New(container.NewWithoutLayout(), card.Panel)
	builder.Add("screen-title-background", background, position)
	title := ui.NewText(state.Title, fontSize, ui.Navy)
	title.Bold, title.Center = true, true
	if state.FitHeader {
		title.MinimumFont, title.MaxLines = 10, -1
	}
	builder.Add("screen-title", title, position)
}

func renderSplitHeaderComponents(builder *contracts.RenderContext) {
	header := builder.State.SplitHeader
	actionWidth := float32(0)
	if builder.State.HeaderAction != "" {
		actionWidth = builder.State.HeaderWidth
		if actionWidth <= 0 {
			actionWidth = 98
		}
	}
	for i, label := range []string{header.Left, header.Right} {
		index := i
		id := fmt.Sprintf("header-half-%d", index)
		position := func(width, _ float32) (float32, float32, float32, float32) {
			if actionWidth > 0 {
				return geometry.SplitHeaderHalfWithActionSlot(width, index, actionWidth)
			}
			x, y, w, _ := geometry.SplitHeaderSlot(width, index)
			return x, y, w, geometry.HeaderHeight(builder.State)
		}
		if index == 0 && builder.State.HoldTitle != nil {
			button := builder.Button("title", label, func() {})
			button.OnHold = builder.State.HoldTitle
			button.Fill, button.Ink, button.FontSize = ui.White, ui.Navy, 22
			button.Compact, button.MinimumFont, button.MaxLabelLines = true, 16, -1
			builder.Add(id, button, position)
			continue
		}
		background := card.New(container.NewWithoutLayout(), card.Panel)
		builder.Add(id+"-background", background, position)
		text := ui.NewText(label, 22, ui.Navy)
		text.Bold, text.Center, text.MinimumFont, text.MaxLines = true, true, 16, -1
		builder.Add(id, text, position)
	}
	if builder.State.HeaderAction != "" {
		button := builder.Button("header", builder.State.HeaderAction, func() { _ = builder.Controller.Header() })
		button.FontSize = 18
		position := func(width, _ float32) (float32, float32, float32, float32) {
			return geometry.SplitHeaderActionSlot(width, actionWidth)
		}
		builder.Add("header-action", button, position)
	}
}

func renderPrimaryInfoComponent(builder *contracts.RenderContext) {
	text := geometry.PrimaryInfoText(builder.State)
	if text == "" {
		return
	}
	info := ui.NewText(text, geometry.PrimaryInfoFontSize(builder.State), ui.Navy)
	info.Bold = true
	builder.Add("primary-info", info, func(width, _ float32) (float32, float32, float32, float32) {
		return ui.ContentInset, ui.OuterMargin + geometry.HeaderHeight(builder.State) + ui.Gap, width - 2*ui.ContentInset, geometry.PrimaryInfoHeight(builder.State, width)
	})
}
