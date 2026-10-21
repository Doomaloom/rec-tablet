package sessionsharing

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"rectablet/next/features/setup/internal/dialogs"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/field"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/components/textinput"
)

func (c *SessionSharing) Open(window fyne.Window) {
	code := ""
	label := widget.NewLabel("Enter six digit code")
	buttons := make([]fyne.CanvasObject, 0, 12)
	for _, digit := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		d := digit
		buttons = append(buttons, actionbutton.New(d, actionbutton.Options{}, func() {
			if len(code) < 6 {
				code += d
				label.SetText(code)
			}
		}))
	}
	buttons = append(buttons, actionbutton.New("Clear", actionbutton.Options{Variant: actionbutton.Outline}, func() { code = ""; label.SetText(code) }))
	buttons = append(buttons, actionbutton.New("0", actionbutton.Options{}, func() {
		if len(code) < 6 {
			code += "0"
			label.SetText(code)
		}
	}))
	buttons = append(buttons, actionbutton.New("Delete", actionbutton.Options{Variant: actionbutton.Outline}, func() {
		if len(code) > 0 {
			code = code[:len(code)-1]
			label.SetText(code)
		}
	}))
	var dlg *modalshell.Modal
	serverButton := actionbutton.New("Configure Share Server", actionbutton.Options{Variant: actionbutton.Outline}, func() {
		dlg.Hide()
		c.OpenServer(window)
	})
	dlg = dialogs.Confirm("Import Session", "Download", "Cancel", container.NewVBox(label, container.NewGridWithColumns(3, buttons...), serverButton), func(submit bool) {
		if !submit {
			return
		}
		if err := c.SubmitCode(code); err != nil {
			c.OnError(err)
		}
	}, window)
	dlg.Show()
}
func (c *SessionSharing) OpenServer(window fyne.Window) {
	entry := textinput.New(textinput.Options{})
	entry.SetText(c.Server())
	form := dialogs.Confirm("Share Server", "Save", "Cancel", field.New(entry.Object(), field.Options{Label: "Server URL"}), func(save bool) {
		if !save {
			return
		}
		if err := c.SaveServer(entry.Text); err != nil {
			c.OnError(err)
		}
	}, window)
	form.Show()
}
