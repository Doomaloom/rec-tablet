package classes

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/next/contracts"
	"rectablet/next/features/classes/classlist"
	"rectablet/next/features/classes/datepicker"
	"rectablet/next/features/classes/lessonmenu"
	"rectablet/next/navigation"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/ui"
)

func (c *Component) composeFeatureBody(state *contracts.State) {
	switch c.navigation.Frame().Screen {
	case navigation.Classes:
		state.HoldTitle = func() { _ = c.navigation.Maintenance() }
		compose := c.classList.Compose(*state)
		state.ComposeBody = compose
		if c.Logout != nil {
			c.configureLogout(state, compose)
		}

	case navigation.Dates:
		state.ComposeBody = c.datePicker.Compose(*state)
	case navigation.Class:
		state.ComposeBody = c.lessonMenu.Compose(*state)
	}
}

func (c *Component) configureLayout(mode ui.DisplayMode) {
	c.classList.Layout = classlist.Layout{Columns: 1, Rows: 3}
	c.lessonMenu.Layout = lessonmenu.Layout{Columns: 1, Rows: 3}
	c.datePicker.Layout = datepicker.Layout{}
	if mode.IsLarge() {
		c.classList.Layout = classlist.Layout{Columns: 2, Rows: 3}
		c.lessonMenu.Layout = lessonmenu.Layout{Columns: 3, Rows: 1}
		c.datePicker.Layout = datepicker.Layout{Calendar: true}
	}
}

// Bind the window at composition time while keeping the home action stable
// across the host's rail/body composition order.
func (c *Component) configureLogout(state *contracts.State, compose ui.BodyComposer) {
	var window fyne.Window
	var modal *modalshell.Modal
	state.HomeRailAction = &contracts.RailAction{ID: "logout", Label: "Log out", Icon: ui.RailIcon("logout", false), Run: func() error {
		if window == nil || c.navigation.Closed || c.navigation.Frame().Screen != navigation.Classes || (modal != nil && modal.Visible()) {
			return nil
		}
		var confirmation *modalshell.Modal
		cancel := actionbutton.New("Cancel", actionbutton.Options{Variant: actionbutton.Outline, Size: actionbutton.Large}, func() { confirmation.Hide() })
		accept := actionbutton.New("Log out", actionbutton.Options{Variant: actionbutton.Primary, Size: actionbutton.Large}, func() {
			if c.navigation.Closed || !confirmation.Visible() {
				return
			}
			confirmation.Hide()
			c.Logout()
		})
		confirmation = modalshell.New(window, nil, modalshell.Options{
			Title: "Confirm Log Out", TitleSize: 32, HideClose: true, InitialFocus: cancel, MaxWidth: 520,
			Footer: container.NewGridWithColumns(2, cancel, accept),
		})
		modal = confirmation
		confirmation.Show()
		return nil
	}}
	state.OnLeave = func() {
		if modal != nil {
			modal.Hide()
		}
	}
	state.ComposeBody = func(surface *ui.Surface) { window = surface.Window; compose(surface) }
}
