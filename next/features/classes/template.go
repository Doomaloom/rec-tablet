package classes

import (
	"rectablet/next/contracts"
	"rectablet/next/features/classes/classlist"
	"rectablet/next/features/classes/datepicker"
	"rectablet/next/features/classes/lessonmenu"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
)

func (c *Component) composeFeatureBody(state *contracts.State) {
	switch c.navigation.Frame().Screen {
	case navigation.Classes:
		state.HoldTitle = func() { _ = c.navigation.Maintenance() }
		state.ComposeBody = c.classList.Compose(*state)
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
