package setup

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/field"
	"rectablet/src/shared/components/modalshell"
	"rectablet/src/shared/components/textinput"
	"rectablet/src/shell"
)

// SessionSharing owns code entry and server settings dialogs.
type SessionSharing struct {
	server       func() string
	onJoin       func(string) error
	onSaveServer func(string) error
	onError      func(error)
}

func (c *SessionSharing) Row() shell.Row {
	return shell.Row{ID: "share", Label: "Import from Session Code"}
}

func (c *SessionSharing) SubmitCode(code string) error {
	if len(code) != 6 {
		return fmt.Errorf("enter the six digit instructor code")
	}
	return c.onJoin(code)
}

func (c *SessionSharing) SaveServer(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("enter a server URL beginning with https:// (localhost may use http://)")
	}
	if parsed.Scheme == "http" && strings.ToLower(parsed.Hostname()) != "localhost" && (net.ParseIP(parsed.Hostname()) == nil || !net.ParseIP(parsed.Hostname()).IsLoopback()) {
		return fmt.Errorf("use https for network servers; http is allowed only for localhost development")
	}
	return c.onSaveServer(parsed.String())
}

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
	dlg = confirmModal("Import Session", "Download", "Cancel", container.NewVBox(label, container.NewGridWithColumns(3, buttons...), serverButton), func(submit bool) {
		if !submit {
			return
		}
		if err := c.SubmitCode(code); err != nil {
			c.onError(err)
		}
	}, window)
	dlg.Show()
}

func (c *SessionSharing) OpenServer(window fyne.Window) {
	entry := textinput.New(textinput.Options{})
	entry.SetText(c.server())
	form := confirmModal("Share Server", "Save", "Cancel", field.New(entry.Object(), field.Options{Label: "Server URL"}), func(save bool) {
		if !save {
			return
		}
		if err := c.SaveServer(entry.Text); err != nil {
			c.onError(err)
		}
	}, window)
	form.Show()
}
