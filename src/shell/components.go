package shell

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"rectablet/src/shared/components/card"
	"rectablet/src/shared/ui"
)

// LayoutComponent is one named visual piece in the composed screen.
type LayoutComponent struct {
	Name     string
	Object   fyne.CanvasObject
	Position LayoutSlot
}

// ScreenComposition is the complete set of visual objects for one state.
type ScreenComposition struct {
	Components []LayoutComponent
	Buttons    map[string]*ui.TouchButton
	Criterion  *ui.TextBlock
}

type componentBuilder struct {
	view       *View
	controller Presenter
	state      State
	components []LayoutComponent
	buttons    map[string]*ui.TouchButton
}

func newComponentBuilder(view *View, state State) *componentBuilder {
	return &componentBuilder{
		view:       view,
		controller: view.Controller,
		state:      state,
		components: []LayoutComponent{},
		buttons:    map[string]*ui.TouchButton{},
	}
}

func (builder *componentBuilder) addComponent(name string, object fyne.CanvasObject, position LayoutSlot) {
	builder.components = append(builder.components, LayoutComponent{Name: name, Object: object, Position: position})
}

func (builder *componentBuilder) addBodyComponent(object fyne.CanvasObject, position LayoutSlot) {
	name := fmt.Sprintf("feature-body-element-%d", len(builder.components))
	builder.addComponent(name, object, position)
}

func (builder *componentBuilder) newTouchButton(id, label string, callback func()) *ui.TouchButton {
	button := ui.NewButton(label, callback)
	button.CanActivate = builder.view.Ready
	builder.buttons[id] = button
	return button
}

func newBackgroundComponent(builder *componentBuilder) {
	background := canvas.NewRectangle(ui.Background)
	builder.addComponent("background", background, backgroundSlot)
}

func newActionRailComponents(builder *componentBuilder) {
	count := len(builder.state.LeftRailActions)
	for i, action := range builder.state.LeftRailActions {
		index, action := i, action
		button := builder.newTouchButton(action.ID, action.Label, func() { _ = action.Run() })
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
		builder.addComponent(action.ID, button, func(_, height float32) (float32, float32, float32, float32) {
			return railSlot(height, index, count)
		})
	}
}

func newHomeRailComponent(builder *componentBuilder) {
	action := builder.state.HomeRailAction
	label := "Home"
	callback := func() { _ = builder.controller.Home() }
	var fill, ink color.Color = ui.Rail, ui.Blue
	fontSize := float32(17)
	if action != nil {
		label = action.Label
		callback = func() { _ = action.Run() }
		fill, ink, fontSize = action.Fill, action.Ink, action.FontSize
		if fill == nil {
			fill = ui.Rail
		}
		if ink == nil {
			ink = ui.Blue
		}
		if fontSize <= 0 {
			fontSize = 32
		}
	}
	homeButton := builder.newTouchButton("home", label, callback)
	homeButton.Fill = fill
	homeButton.Ink = ink
	homeButton.FontSize = fontSize
	if action == nil {
		homeButton.Icon = theme.NewPrimaryThemedResource(theme.HomeIcon())
		homeButton.IconOnly = true
	} else if action.Icon != nil {
		homeButton.Icon = action.Icon
		homeButton.IconOnly = true
	}
	homePosition := homeRailSlot
	if builder.state.CanBack || builder.state.BackRailAction != nil {
		homePosition = splitHomeRailSlot
	}
	builder.addComponent("home-rail", homeButton, homePosition)
}

func newPagingRailComponents(builder *componentBuilder) {
	state := builder.state
	previousButton := builder.newTouchButton("up", "↑", func() { _ = builder.controller.Page(-1) })
	previousButton.Fill = ui.Rail
	previousButton.Ink = ui.Blue
	previousButton.FontSize = 58
	previousButton.Disabled = state.Page == 0
	previousButton.IconOnly = true
	previousButton.Icon = chevronIcon("up", previousButton.Disabled)
	builder.addComponent("previous-page", previousButton, previousPageSlot)

	nextButton := builder.newTouchButton("down", "↓", func() { _ = builder.controller.Page(1) })
	nextButton.Fill = ui.Rail
	nextButton.Ink = ui.Blue
	nextButton.FontSize = 58
	nextButton.Disabled = state.Page+1 >= state.PageCount
	nextButton.IconOnly = true
	nextButton.Icon = chevronIcon("down", nextButton.Disabled)
	builder.addComponent("next-page", nextButton, nextPageSlot)
}

func newPagePositionIndicatorComponents(builder *componentBuilder) {
	if builder.state.PageCount <= 1 {
		return
	}
	track := canvas.NewRectangle(ui.Rail)
	track.CornerRadius = pagePositionTrackWidth / 2
	track.StrokeColor = ui.Muted
	track.StrokeWidth = 1
	builder.addComponent("page-position-track", track, pagePositionTrackSlot(builder.state))

	thumb := canvas.NewRectangle(ui.Blue)
	thumb.CornerRadius = pagePositionTrackWidth / 2
	thumb.StrokeColor = ui.Navy
	thumb.StrokeWidth = 1
	builder.addComponent("page-position-thumb", thumb, pagePositionThumbSlot(builder.state))
}

func newBackButtonComponent(builder *componentBuilder) {
	if !builder.state.CanBack && builder.state.BackRailAction == nil {
		return
	}
	action := builder.state.BackRailAction
	label := "‹"
	callback := func() { _ = builder.controller.Back() }
	var fill, ink color.Color = ui.Rail, ui.Blue
	fontSize := float32(36)
	if action != nil {
		label = action.Label
		callback = func() { _ = action.Run() }
		fill, ink, fontSize = action.Fill, action.Ink, action.FontSize
		if fill == nil {
			fill = ui.Rail
		}
		if ink == nil {
			ink = ui.Blue
		}
		if fontSize <= 0 {
			fontSize = 32
		}
	}
	backButton := builder.newTouchButton("back", label, callback)
	backButton.Fill = fill
	backButton.Ink = ink
	backButton.FontSize = fontSize
	if action == nil {
		backButton.Icon = chevronIcon("left", false)
		backButton.IconOnly = true
	}
	if action != nil && action.Icon != nil {
		backButton.Icon = action.Icon
		backButton.IconOnly = true
	}
	builder.addComponent("back-button", backButton, splitBackRailSlot)
}

func newHeaderComponents(builder *componentBuilder) {
	state := builder.state
	if state.SplitHeader != nil {
		newSplitHeaderComponents(builder)
		return
	}
	actionWidth := float32(0)
	actions := state.HeaderActions
	if len(actions) > 0 {
		actionWidth = float32(len(actions))*98 + float32(len(actions)-1)*ui.Gap
		for i, action := range actions {
			index, action := i, action
			button := builder.newTouchButton(fmt.Sprintf("header-action-%d", i), action.Label, func() { _ = action.Run() })
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
				return x, ui.OuterMargin, 98, headerHeight(state)
			}
			builder.addComponent(fmt.Sprintf("header-action-%d", i), button, position)
		}
	} else if state.HeaderAction != "" {
		actionWidth = 98
		if state.HeaderWidth > 0 {
			actionWidth = state.HeaderWidth
		}
		headerAction := builder.newTouchButton("header", state.HeaderAction, func() { _ = builder.controller.Header() })
		headerAction.FontSize = 18
		position := func(width, height float32) (float32, float32, float32, float32) {
			return headerActionSlot(width, height, actionWidth)
		}
		builder.addComponent("header-action", headerAction, position)
	}

	if len(state.HeaderSwitch) > 0 {
		const optionWidth = float32(92)
		const optionGap = ui.Gap
		switchWidth := float32(len(state.HeaderSwitch))*optionWidth + float32(len(state.HeaderSwitch)-1)*optionGap
		for i, option := range state.HeaderSwitch {
			index, option := i, option
			button := builder.newTouchButton(fmt.Sprintf("header-switch-%d", i), option.Label, func() { _ = option.Run() })
			button.FontSize = 18
			button.Fill, button.Ink = ui.White, ui.Navy
			if option.Selected {
				button.Fill, button.Ink = ui.Blue, ui.White
				button.Selected = true
			}
			position := func(width, _ float32) (float32, float32, float32, float32) {
				return width - ui.ContentInset - switchWidth + float32(index)*(optionWidth+optionGap), ui.OuterMargin, optionWidth, ui.HeaderHeight
			}
			builder.addComponent(fmt.Sprintf("header-switch-%d", i), button, position)
		}
		actionWidth = switchWidth
	}

	left := ui.ContentInset
	fontSize := float32(25)
	if state.HeaderTitleFontSize > 0 {
		fontSize = state.HeaderTitleFontSize
	}
	position := func(width, height float32) (float32, float32, float32, float32) {
		x, y, w, _ := headerTitleSlot(width, height, left, actionWidth)
		return x, y, w, headerHeight(state)
	}
	if state.HoldTitle != nil {
		titleButton := builder.newTouchButton("title", state.Title, func() {})
		titleButton.Fill, titleButton.Ink, titleButton.FontSize = ui.White, ui.Navy, fontSize
		titleButton.OnHold = state.HoldTitle
		if state.FitHeader {
			titleButton.Compact, titleButton.MinimumFont, titleButton.MaxLabelLines = true, 10, -1
		}
		builder.addComponent("screen-title", titleButton, position)
		return
	}
	background := card.New(container.NewWithoutLayout(), card.Panel)
	builder.addComponent("screen-title-background", background, position)
	title := ui.NewText(state.Title, fontSize, ui.Navy)
	if state.BindTitle != nil {
		state.BindTitle(func(value string) {
			title.Text = value
			title.Refresh()
		})
	}
	title.Bold, title.Center = true, true
	if state.FitHeader {
		title.MinimumFont, title.MaxLines = 10, -1
	}
	builder.addComponent("screen-title", title, position)
}

func newSplitHeaderComponents(builder *componentBuilder) {
	header := builder.state.SplitHeader
	actionWidth := float32(0)
	if builder.state.HeaderAction != "" {
		actionWidth = builder.state.HeaderWidth
		if actionWidth <= 0 {
			actionWidth = 98
		}
	}
	for i, label := range []string{header.Left, header.Right} {
		index := i
		id := fmt.Sprintf("header-half-%d", index)
		position := func(width, _ float32) (float32, float32, float32, float32) {
			if actionWidth > 0 {
				return splitHeaderHalfWithActionSlot(width, index, actionWidth)
			}
			x, y, w, _ := splitHeaderSlot(width, index)
			return x, y, w, headerHeight(builder.state)
		}
		if index == 0 && builder.state.HoldTitle != nil {
			button := builder.newTouchButton("title", label, func() {})
			button.OnHold = builder.state.HoldTitle
			button.Fill, button.Ink, button.FontSize = ui.White, ui.Navy, 22
			button.Compact, button.MinimumFont, button.MaxLabelLines = true, 16, -1
			builder.addComponent(id, button, position)
			continue
		}
		background := card.New(container.NewWithoutLayout(), card.Panel)
		builder.addComponent(id+"-background", background, position)
		text := ui.NewText(label, 22, ui.Navy)
		text.Bold, text.Center, text.MinimumFont, text.MaxLines = true, true, 16, -1
		builder.addComponent(id, text, position)
	}
	if builder.state.HeaderAction != "" {
		button := builder.newTouchButton("header", builder.state.HeaderAction, func() { _ = builder.controller.Header() })
		button.FontSize = 18
		position := func(width, _ float32) (float32, float32, float32, float32) {
			return splitHeaderActionSlot(width, actionWidth)
		}
		builder.addComponent("header-action", button, position)
	}
}

func newPrimaryInfoComponent(builder *componentBuilder) {
	text := primaryInfoText(builder.state)
	if text == "" {
		return
	}
	info := ui.NewText(text, primaryInfoFontSize(builder.state), ui.Navy)
	info.Bold = true
	builder.addComponent("primary-info", info, func(width, _ float32) (float32, float32, float32, float32) {
		return ui.ContentInset, ui.OuterMargin + headerHeight(builder.state) + ui.Gap, width - 2*ui.ContentInset, primaryInfoHeight(builder.state, width)
	})
}

func newBodyComponent(builder *componentBuilder) {
	if builder.state.ComposeBody == nil {
		return
	}
	surface := &ui.Surface{
		DisplayMode: builder.state.DisplayMode,
		Window:      builder.view.Window,
		Bounds:      bodySlot(builder.state),
		Add: func(object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32)) {
			builder.addBodyComponent(object, position)
		},
		AddNamed: func(name string, object fyne.CanvasObject, position func(float32, float32) (float32, float32, float32, float32)) {
			builder.addComponent("feature-"+name, object, position)
		},
		Button: builder.newTouchButton,
	}
	builder.state.ComposeBody(surface)
	builder.view.Criterion = surface.Criterion
}

func newFooterComponents(builder *componentBuilder) {
	state := builder.state
	if state.HideFooter || state.FooterAction == "" {
		return
	}
	if state.FooterAction != "" {
		submitButton := builder.newTouchButton("submit", state.FooterAction, func() { _ = builder.controller.Submit() })
		submitButton.Disabled = !state.CanSubmit
		submitButton.Fill = ui.Navy
		submitButton.FontSize = 20
		builder.addComponent("submit-footer", submitButton, submitFooterSlot)
	}
}
