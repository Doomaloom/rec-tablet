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
	builder.addComponent("back-button", backButton, backRailSlot)
}

func newHeaderComponents(builder *componentBuilder) {
	state := builder.state
	actionWidth := float32(0)
	if state.HeaderAction != "" {
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

	left := float32(100)
	titleButton := builder.newTouchButton("title", state.Title, func() {
		if state.TitleDetails {
			builder.view.details()
		}
	})
	titleButton.Fill = ui.White
	titleButton.Ink = ui.Navy
	titleButton.FontSize = 25
	titleButton.OnHold = state.HoldTitle
	position := func(width, height float32) (float32, float32, float32, float32) {
		return headerTitleSlot(width, height, left, actionWidth)
	}
	builder.addComponent("screen-title", titleButton, position)
}

func newContextComponent(builder *componentBuilder) {
	contextButton := builder.newTouchButton("context", builder.state.Subtitle, builder.view.details)
	contextButton.Fill = ui.Background
	contextButton.Ink = ui.Muted
	contextButton.FontSize = 17
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
