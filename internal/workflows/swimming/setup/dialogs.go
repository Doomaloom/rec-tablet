package setup

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"path/filepath"
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
		entry := widget.NewEntry()
		entry.SetText("rec-tablet-results.json")
		form := dialog.NewForm("Export Data", "Save", "Cancel", []*widget.FormItem{widget.NewFormItem("File name", entry)}, func(save bool) {
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
		form.Resize(fyne.NewSize(650, 280))
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
		dialog.ShowConfirm("Replace export?", "Replace "+filepath.Base(path)+"?", func(ok bool) {
			if ok {
				_ = v.Controller.Export(path)
			}
		}, v.Window)
	} else {
		_ = v.Controller.Export(path)
	}
}
