package setup

import (
	"strings"

	"rectablet/src/shell"
)

// ProfilePicker presents saved instructor sessions and emits their keys.
type ProfilePicker struct {
	onSelect func(string) error
}

func (c *ProfilePicker) Rows(profiles []ProfileOption) []shell.Row {
	rows := make([]shell.Row, 0, len(profiles))
	for _, profile := range profiles {
		rows = append(rows, shell.Row{ID: "profile:" + profile.Key, Label: "Switch Session · " + profile.Label})
	}
	return rows
}

func (c *ProfilePicker) Select(id string) error {
	if !strings.HasPrefix(id, "profile:") {
		return nil
	}
	return c.onSelect(strings.TrimPrefix(id, "profile:"))
}
