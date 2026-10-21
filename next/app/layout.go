package app

import (
	"rectablet/next/app/background"
	"rectablet/next/app/content"
	"rectablet/next/app/footer"
	"rectablet/next/app/header"
	"rectablet/next/app/leftrail"
	"rectablet/next/app/rightrail"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

// ComposeScreen is the persistent layout's only composition point.
func ComposeScreen(view *View, state contracts.State) contracts.ScreenComposition {
	c := &contracts.RenderContext{Controller: view.Controller, State: state, Window: view.Window, Ready: view.Ready, Buttons: map[string]*ui.TouchButton{}}
	background.Controller{}.Compose(c)
	leftrail.Controller{}.Compose(c)
	rightrail.Controller{}.Compose(c)
	header.Controller{}.Compose(c)
	content.Controller{}.Compose(c)
	footer.Controller{}.Compose(c)
	return contracts.ScreenComposition{Components: c.Components, Buttons: c.Buttons, Criterion: c.Criterion}
}
