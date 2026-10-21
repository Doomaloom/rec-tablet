package filetransfer

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"

	"rectablet/next/contracts"
)

// FileTransfer owns file selection and export confirmation, with persistence callbacks.
type FileTransfer struct {
	OnImport func(string) error
	OnExport func(string) error
	OnError  func(error)
}

func (c *FileTransfer) ImportRow() contracts.Row {
	return contracts.Row{ID: "import", Label: "Import Class Data · Class package (.json)"}
}
func (c *FileTransfer) ExportRow() contracts.Row {
	return contracts.Row{ID: "export", Label: "Export Data · Attendance, attempts, and drafts"}
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
