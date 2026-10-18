package setup

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"rectablet/src/shared/components/design"
	"rectablet/src/shared/components/field"
	"rectablet/src/shared/components/textinput"
	"rectablet/src/shell"
)

// FileTransfer owns file selection and export confirmation, with persistence callbacks.
type FileTransfer struct {
	onImport func(string) error
	onExport func(string) error
	onError  func(error)
}

func (c *FileTransfer) ImportRow() shell.Row {
	return shell.Row{ID: "import", Label: "Import Class Data · Class package (.json)"}
}

func (c *FileTransfer) ExportRow() shell.Row {
	return shell.Row{ID: "export", Label: "Export Data · Attendance, attempts, and drafts"}
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

func (c *FileTransfer) OpenImport(window fyne.Window) {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			c.onError(err)
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
			c.onError(e)
			return
		}
		_ = c.onImport(path)
	}, window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	// Fyne creates a FileDialog's underlying popup in Show. Resize calls
	// MinSize, which needs that popup to exist (including in Fyne 2.8.1).
	d.Show()
	d.Resize(fyne.NewSize(700, 410))
}

func (c *FileTransfer) OpenExport(window fyne.Window) {
	// Fyne's FileSave opens/truncates the destination before its callback. Use a
	// directory chooser plus a name/overwrite confirmation so the export callback owns
	// the atomic write and can reject database paths before touching the file.
	chooser := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			c.onError(err)
			return
		}
		if uri == nil {
			return
		}
		folder, e := localPath(uri)
		if e != nil {
			c.onError(e)
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
				c.onError(fmt.Errorf("enter a file name without directories"))
				return
			}
			if strings.ToLower(filepath.Ext(name)) != ".json" {
				name = strings.TrimSuffix(name, filepath.Ext(name)) + ".json"
			}
			path := filepath.Join(folder, name)
			c.confirmExport(path, window)
		}, window)
		form.Show()
	}, window)
	chooser.Show()
	chooser.Resize(fyne.NewSize(700, 410))
}

func (c *FileTransfer) confirmExport(path string, window fyne.Window) {
	exists, e := storage.Exists(storage.NewFileURI(path))
	if e != nil {
		c.onError(e)
		return
	}
	if exists {
		confirmModal("Replace export?", "Replace", "Cancel", design.Label("Replace "+filepath.Base(path)+"?", 16, design.Secondary, false), func(ok bool) {
			if ok {
				_ = c.onExport(path)
			}
		}, window).Show()
	} else {
		_ = c.onExport(path)
	}
}
