package setup

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"net"
	"net/url"
	"path/filepath"
	"rectablet/internal/shared/components/actionbutton"
	"rectablet/internal/shared/components/design"
	"rectablet/internal/shared/components/field"
	"rectablet/internal/shared/components/modalshell"
	"rectablet/internal/shared/components/textinput"
	"strings"
)

type dialogs struct {
	Controller *Controller
	Window     fyne.Window
}

func localPath(uri fyne.URI) (string, error) {
	if uri == nil {
		return "", fmt.Errorf("no file selected")
	}
	if uri.Scheme() != "file" {
		return "", fmt.Errorf("choose a local file")
	}
	u, e := url.Parse(uri.String())
	if e != nil {
		return "", e
	}
	return u.Path, nil
}
func (v *dialogs) importDialog() {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			v.Controller.Fail(err)
			return
		}
		if reader == nil {
			return
		}
		path, e := localPath(reader.URI())
		closeErr := reader.Close()
		if e == nil {
			e = closeErr
		}
		if e != nil {
			v.Controller.Fail(e)
			return
		}
		_ = v.Controller.Import(path)
	}, v.Window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	// Fyne creates a FileDialog's underlying popup in Show. Resize calls
	// MinSize, which needs that popup to exist (including in Fyne 2.8.1).
	d.Show()
	d.Resize(fyne.NewSize(700, 410))
}

func (v *dialogs) shareDialog() {
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
		v.shareServerDialog()
	})
	dlg = confirmModal("Import Session", "Download", "Cancel", container.NewVBox(label, container.NewGridWithColumns(3, buttons...), serverButton), func(submit bool) {
		if !submit {
			return
		}
		if len(code) != 6 {
			v.Controller.Fail(fmt.Errorf("enter the six digit instructor code"))
			return
		}
		if v.Controller.JoinShare == nil {
			v.Controller.Fail(fmt.Errorf("session sharing is not configured"))
			return
		}
		v.Controller.State.Notice = "Connecting to the sharing server…"
		v.Controller.State.Error = false
		_ = v.Controller.Refresh()
		v.Controller.JoinShare(code, func(err error) {
			if err != nil {
				v.Controller.Fail(err)
			}
		})
	}, v.Window)
	dlg.Show()
}

func (v *dialogs) shareServerDialog() {
	entry := textinput.New(textinput.Options{})
	entry.SetText(v.Controller.ShareServer)
	form := confirmModal("Share Server", "Save", "Cancel", field.New(entry.Object(), field.Options{Label: "Server URL"}), func(save bool) {
		if !save {
			return
		}
		parsed, err := url.Parse(strings.TrimSpace(entry.Text))
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			v.Controller.Fail(fmt.Errorf("enter a server URL beginning with https:// (localhost may use http://)"))
			return
		}
		if parsed.Scheme == "http" && strings.ToLower(parsed.Hostname()) != "localhost" && (net.ParseIP(parsed.Hostname()) == nil || !net.ParseIP(parsed.Hostname()).IsLoopback()) {
			v.Controller.Fail(fmt.Errorf("use https for network servers; http is allowed only for localhost development"))
			return
		}
		if v.Controller.SaveShareServer == nil {
			v.Controller.Fail(fmt.Errorf("server settings are unavailable"))
			return
		}
		if err := v.Controller.SaveShareServer(parsed.String()); err != nil {
			v.Controller.Fail(err)
			return
		}
		v.Controller.ShareServer = strings.TrimRight(parsed.String(), "/")
		_ = v.Controller.Refresh()
	}, v.Window)
	form.Show()
}
func (v *dialogs) exportDialog() {
	// Fyne's FileSave opens/truncates the destination before its callback. Use a
	// directory chooser plus a name/overwrite confirmation so Store.Export owns
	// the atomic write and can reject database paths before touching the file.
	chooser := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			v.Controller.Fail(err)
			return
		}
		if uri == nil {
			return
		}
		folder, e := localPath(uri)
		if e != nil {
			v.Controller.Fail(e)
			return
		}
		entry := textinput.New(textinput.Options{})
		entry.SetText("rec-tablet-results.json")
		form := confirmModal("Export Data", "Save", "Cancel", field.New(entry.Object(), field.Options{Label: "File name"}), func(save bool) {
			if !save {
				return
			}
			name := strings.TrimSpace(entry.Text)
			if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
				v.Controller.Fail(fmt.Errorf("enter a file name without directories"))
				return
			}
			if strings.ToLower(filepath.Ext(name)) != ".json" {
				name = strings.TrimSuffix(name, filepath.Ext(name)) + ".json"
			}
			path := filepath.Join(folder, name)
			v.confirmExport(path)
		}, v.Window)
		form.Show()
	}, v.Window)
	chooser.Show()
	chooser.Resize(fyne.NewSize(700, 410))
}
func (v *dialogs) confirmExport(path string) {
	exists, e := storage.Exists(storage.NewFileURI(path))
	if e != nil {
		v.Controller.Fail(e)
		return
	}
	if exists {
		confirmModal("Replace export?", "Replace", "Cancel", design.Label("Replace "+filepath.Base(path)+"?", 16, design.Secondary, false), func(ok bool) {
			if ok {
				_ = v.Controller.Export(path)
			}
		}, v.Window).Show()
	} else {
		_ = v.Controller.Export(path)
	}
}

// File-system selection stays with Fyne; app-owned confirmations use ModalShell.
func confirmModal(title, accept, cancel string, content fyne.CanvasObject, callback func(bool), window fyne.Window) *modalshell.Modal {
	var modal *modalshell.Modal
	finish := func(ok bool) { modal.Hide(); callback(ok) }
	actions := container.NewGridWithColumns(2,
		actionbutton.New(cancel, actionbutton.Options{Variant: actionbutton.Outline}, func() { finish(false) }),
		actionbutton.New(accept, actionbutton.Options{Variant: actionbutton.Primary}, func() { finish(true) }),
	)
	modal = modalshell.New(window, content, modalshell.Options{Title: title, MaxWidth: 600, Footer: actions})
	return modal
}
