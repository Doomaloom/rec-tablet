package shell

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"image/color"
	"rectablet/internal/shared/ui"
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
			const gap float32 = 10
			h := (height - 28 - float32(count-1)*gap) / float32(count)
			return 12, 14 + float32(index)*(h+gap), 76, h
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
		homeButton.Icon = homeIcon
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
	previousButton.Icon = arrowIcon(true, previousButton.Disabled)
	builder.addComponent("previous-page", previousButton, previousPageSlot)

	nextButton := builder.newTouchButton("down", "↓", func() { _ = builder.controller.Page(1) })
	nextButton.Fill = ui.Rail
	nextButton.Ink = ui.Blue
	nextButton.FontSize = 58
	nextButton.Disabled = state.Page+1 >= state.PageCount
	nextButton.IconOnly = true
	nextButton.Icon = arrowIcon(false, nextButton.Disabled)
	builder.addComponent("next-page", nextButton, nextPageSlot)
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
	if action != nil && action.Icon != nil {
		backButton.Icon = action.Icon
		backButton.IconOnly = true
	}
	builder.addComponent("back-button", backButton, splitBackRailSlot)
}

func newHeaderComponents(builder *componentBuilder) {
	state := builder.state
	actionWidth := float32(0)
	actions := state.HeaderActions
	if len(actions) > 0 {
		actionWidth = float32(len(actions))*98 + float32(len(actions)-1)*8
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
				x := width - 100 - actionWidth + float32(index)*106
				return x, 14, 98, 58
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
		switchWidth := float32(len(state.HeaderSwitch)) * 140
		const optionGap float32 = 6
		optionWidth := (switchWidth - float32(len(state.HeaderSwitch)-1)*optionGap) / float32(len(state.HeaderSwitch))
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
				return (width-switchWidth)/2 + float32(index)*(optionWidth+optionGap), 14, optionWidth, 58
			}
			builder.addComponent(fmt.Sprintf("header-switch-%d", i), button, position)
		}
		return
	}

	left := float32(100)
	titleButton := builder.newTouchButton("title", state.Title, func() {
		if state.TitleDetails {
			builder.view.details()
		}
	})
	titleButton.Fill = ui.White
	titleButton.Ink = ui.Navy
	titleButton.FontSize = 25
	if state.HeaderTitleFontSize > 0 {
		titleButton.FontSize = state.HeaderTitleFontSize
	}
	titleButton.OnHold = state.HoldTitle
	if state.FitHeader {
		titleButton.Compact = true
		titleButton.MinimumFont = 10
		titleButton.MaxLabelLines = -1
	}
	position := func(width, height float32) (float32, float32, float32, float32) {
		return headerTitleSlot(width, height, left, actionWidth)
	}
	builder.addComponent("screen-title", titleButton, position)
}

func newContextComponent(builder *componentBuilder) {
	if builder.state.HideContext {
		return
	}
	contextButton := builder.newTouchButton("context", builder.state.Subtitle, builder.view.details)
	contextButton.Fill = ui.Background
	contextButton.Ink = ui.Muted
	contextButton.FontSize = 17
	if builder.state.FitHeader {
		contextButton.Compact = true
		contextButton.MinimumFont = 12
		contextButton.MaxLabelLines = -1
	}
	builder.addComponent("screen-context", contextButton, contextSlot)
}

func newBodyComponent(builder *componentBuilder) {
	if builder.state.ComposeBody == nil {
		return
	}
	surface := &ui.Surface{
		Window: builder.view.Window,
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
	if state.HideFooter {
		return
	}
	progressText := state.Progress
	if progressText != "" {
		progressText += "\n"
	}
	progressText += formatPageIndicator(state.Page, state.PageCount)
	progress := ui.NewText(progressText, 16, ui.Muted)
	builder.addComponent("progress-footer", progress, progressFooterSlot)

	if state.FooterAction != "" {
		submitButton := builder.newTouchButton("submit", state.FooterAction, func() { _ = builder.controller.Submit() })
		submitButton.Disabled = !state.CanSubmit
		submitButton.Fill = ui.Navy
		submitButton.FontSize = 20
		builder.addComponent("submit-footer", submitButton, submitFooterSlot)
		return
	}

	deviceStatus := ui.NewText("ON DEVICE", 12, ui.Green)
	deviceStatus.Center = true
	builder.addComponent("device-status", deviceStatus, deviceStatusSlot)
}

func formatPageIndicator(page, pageCount int) string {
	return fmt.Sprintf("Page %d of %d", page+1, pageCount)
}
