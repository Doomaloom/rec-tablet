// Package files provides atomic local file operations.
package files

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func WriteJSON(path string, value any) error {
	f, e := os.CreateTemp(filepath.Dir(path), ".rec-tablet-*.tmp")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if e = encoder.Encode(value); e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}
