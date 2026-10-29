package classes

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"rectablet/next/contracts"
	"rectablet/next/features/classes/classlist"
	"rectablet/next/features/classes/datepicker"
	"rectablet/next/features/classes/lessonmenu"
	"rectablet/next/navigation"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/design"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/ui"
	"time"
)

func (c *Component) composeFeatureBody(state *contracts.State) {
	switch c.navigation.Frame().Screen {
	case navigation.Classes:
		state.HoldTitle = func() { _ = c.navigation.Maintenance() }
		bindReset := c.configureClassProgressReset(state)
		compose := c.classList.Compose(*state)
		state.ComposeBody = func(surface *ui.Surface) { bindReset(surface); compose(surface) }
		if c.Logout != nil {
			c.configureLogout(state, state.ComposeBody)
		}

	case navigation.Dates:
		state.ComposeBody = c.datePicker.Compose(*state)
	case navigation.Class:
		state.ComposeBody = c.lessonMenu.Compose(*state)
	}
}

func (c *Component) configureClassProgressReset(state *contracts.State) func(*ui.Surface) {
	var window fyne.Window
	var modal *modalshell.Modal
	active := true
	for i := range state.Rows {
		course := state.Rows[i]
		state.Rows[i].HoldDuration = 5 * time.Second
		state.Rows[i].OnHold = func() {
			if !active || window == nil || c.navigation.Closed || c.navigation.Frame().Screen != navigation.Classes || window.Canvas().Overlays().Top() != nil {
				return
			}
			var confirmation *modalshell.Modal
			cancel := actionbutton.New("Cancel", actionbutton.Options{Variant: actionbutton.Outline, Size: actionbutton.Large}, func() { confirmation.Hide() })
			accept := actionbutton.New("Delete progress", actionbutton.Options{Variant: actionbutton.Danger, Size: actionbutton.Large}, func() {
				if !active || c.navigation.Closed || !confirmation.Visible() || c.navigation.Frame().Screen != navigation.Classes {
					return
				}
				confirmation.Hide()
				_ = c.navigation.Run(func() error {
					if err := c.store.ResetClassProgress(course.ID); err != nil {
						return err
					}
					c.navigation.Saved("")
					return nil
				})
			})
			confirmation = modalshell.New(window, design.Label(fmt.Sprintf("Delete all marked progress for all students in %s? This clears skill results, attempt history, and unfinished assessments across all lesson dates. This cannot be undone.", course.Label), 18, design.Secondary, false), modalshell.Options{
				Title: "Delete Class Progress?", TitleSize: 28, HideClose: true, InitialFocus: cancel, MaxWidth: 600,
				Footer: container.NewGridWithColumns(2, cancel, accept),
			})
			modal = confirmation
			confirmation.Show()
		}
	}
	state.OnLeave = func() {
		active = false
		if modal != nil {
			modal.Hide()
		}
	}
	return func(surface *ui.Surface) { window = surface.Window }
}

func (c *Component) configureLayout(mode ui.DisplayMode) {
	c.classList.Layout = classlist.Layout{Columns: 1, Rows: 3}
	c.lessonMenu.Layout = lessonmenu.Layout{Columns: 1, Rows: 3}
	c.datePicker.Layout = datepicker.Layout{}
	if mode.IsLarge() {
		c.classList.Layout = classlist.Layout{Columns: 2, Rows: 3}
		c.lessonMenu.Layout = lessonmenu.Layout{Columns: 2, Rows: 2, Plans: true}
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
	onLeave := state.OnLeave
	state.OnLeave = func() {
		if onLeave != nil {
			onLeave()
		}
		if modal != nil {
			modal.Hide()
		}
	}
	state.ComposeBody = func(surface *ui.Surface) { window = surface.Window; compose(surface) }
}
