package dialogs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/modalshell"
)

// File selection stays with Fyne; child components share app-owned confirmations.
func Confirm(title, accept, cancel string, content fyne.CanvasObject, callback func(bool), window fyne.Window) *modalshell.Modal {
	var modal *modalshell.Modal
	finish := func(ok bool) { modal.Hide(); callback(ok) }
	actions := container.NewGridWithColumns(2,
		actionbutton.New(cancel, actionbutton.Options{Variant: actionbutton.Outline}, func() { finish(false) }),
		actionbutton.New(accept, actionbutton.Options{Variant: actionbutton.Primary}, func() { finish(true) }),
	)
	modal = modalshell.New(window, content, modalshell.Options{Title: title, MaxWidth: 600, Footer: actions})
	return modal
}
