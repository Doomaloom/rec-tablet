package login

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"rectablet/next/contracts"
	"rectablet/next/features/setup/displaypicker"
	"rectablet/next/features/setup/sessionsharing"
	"rectablet/src/shared/ui"
)

var backspaceIcon = fyne.NewStaticResource("backspace.svg", []byte(fmt.Sprintf(
	`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M9 4h11a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H9l-7-8 7-8Zm3 5 6 6m0-6-6 6" fill="none" stroke="#%02x%02x%02x" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg>`,
	ui.Navy.R, ui.Navy.G, ui.Navy.B)))

func (c *Component) Snapshot() contracts.State {
	state := contracts.State{
		DisplayMode: c.options.DisplayMode, Screen: "login", HideHeader: true,
		HideLeftRail: true, HideFooter: true, PageCount: 1, RightRailActions: []contracts.RailAction{},
		Notice: c.notice, Error: c.error,
	}
	if c.settings {
		state.Screen, state.Title, state.PrimaryInfo = "login_setup", "Device Setup", ""
		state.HideLeftRail, state.CanBack, state.HideHeader = false, true, false
		state.HoldTitle = nil
		state.ComposeBody = c.composeSettings
	} else {
		keypadWidth := float32(420)
		if state.DisplayMode.IsLarge() {
			keypadWidth = 520
		}
		code, busy := c.code, c.busy
		state.ComposeBody = func(surface *ui.Surface) { c.composeKeypad(surface, code, busy, keypadWidth); c.composeNotice(surface) }
	}
	return state
}

func (c *Component) composeKeypad(surface *ui.Surface, code string, busy bool, maxWidth float32) {
	position := func(row int, column int, columns int) ui.Bounds {
		return func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := surface.BodyBounds(w, h)
			width = min(width, maxWidth)
			x = (w - width) / 2
			// Center the six control rows with equal space above and below,
			// leaving room beneath them for login errors.
			const verticalMargin = float32(64)
			y += verticalMargin
			rowHeight := (height - 2*verticalMargin - 5*ui.Gap) / 6
			cellWidth := (width - float32(columns-1)*ui.Gap) / float32(columns)
			return x + float32(column)*(cellWidth+ui.Gap), y + float32(row)*(rowHeight+ui.Gap), cellWidth, rowHeight
		}
	}
	for i := range 6 {
		index := i
		slotPosition := func(w, h float32) (float32, float32, float32, float32) {
			x, y, width, height := position(0, 0, 1)(w, h)
			const gap = float32(8)
			slotWidth := (width - 5*gap) / 6
			return x + float32(index)*(slotWidth+gap), y, slotWidth, height
		}
		box := canvas.NewRectangle(ui.White)
		box.StrokeColor, box.StrokeWidth, box.CornerRadius = ui.Blue, 2, 8
		surface.AddNamed(fmt.Sprintf("code-slot-%d", i), box, slotPosition)
		digit := ""
		if i < len(code) {
			digit = code[i : i+1]
		}
		label := ui.NewText(digit, 30, ui.Navy)
		label.Center, label.Bold = true, true
		surface.Add(label, slotPosition)
	}
	for i, digit := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "Clear", "0", "Delete"} {
		digit := digit
		run := func() { c.Digit(digit) }
		if digit == "Clear" {
			run = c.Clear
		}
		if digit == "Delete" {
			run = c.Delete
		}
		button := surface.Button("login:"+digit, digit, run)
		button.Disabled = busy
		button.FontSize = 26
		if digit == "Clear" {
			button.OnHold = func() { _ = c.OpenSetup() }
		}
		if digit == "Clear" || digit == "Delete" {
			button.Fill, button.Ink, button.FontSize = ui.White, ui.Navy, 18
		}
		if digit == "Delete" {
			button.Icon, button.IconOnly = backspaceIcon, true
			if c.options.Resume != nil {
				button.OnHold = c.Resume
				button.HoldDuration = 3 * time.Second
			}
		}
		surface.Add(button, position(1+i/3, i%3, 3))
	}
	title := "Log in"
	if busy {
		title = "Logging in…"
	}
	button := surface.Button("login:submit", title, func() { _ = c.Submit() })
	button.Disabled, button.Fill, button.FontSize = busy || len(code) != 6, ui.Navy, 22
	surface.Add(button, position(5, 0, 1))
}

func (c *Component) composeSettings(surface *ui.Surface) {
	c.composeNotice(surface)
	ui.Grid(surface, []ui.Item{
		{ID: "server", Label: "Configure Share Server"},
		{ID: "display", Label: "Display · " + c.options.DisplayMode.Label()},
	}, 1, 3, func(id string) {
		if c.closed {
			return
		}
		switch id {
		case "server":
			share := sessionsharing.SessionSharing{
				Server: c.options.Server,
				OnSaveServer: func(value string) error {
					if c.closed || c.options.SaveServer == nil {
						return nil
					}
					return c.options.SaveServer(value)
				},
				OnError: c.Fail,
			}
			share.OpenServer(surface.Window)
		case "display":
			picker := displaypicker.DisplayPicker{OnSelect: func(mode ui.DisplayMode) error {
				if c.closed {
					return nil
				}
				c.options.DisplayMode = mode.Normalize()
				if c.options.SaveDisplayMode != nil {
					c.options.SaveDisplayMode(c.options.DisplayMode)
				}
				c.refresh()
				return nil
			}}
			picker.Open(surface.Window)
		}
	})
}

func (c *Component) composeNotice(surface *ui.Surface) {
	if c.notice == "" {
		return
	}
	text := ui.NewText(c.notice, 16, ui.Red)
	text.Center, text.MinimumFont, text.MaxLines = true, 14, 3
	surface.AddNamed("login-notice", text, func(w, h float32) (float32, float32, float32, float32) {
		x, y, width, height := surface.BodyBounds(w, h)
		return x, y + height - 60, width, 60
	})
}
