package profilepicker

import (
	"strings"
)

// ProfilePicker presents saved instructor sessions and emits their keys.
type ProfilePicker struct {
	OnSelect func(string) error
}

func (c *ProfilePicker) Select(id string) error {
	if !strings.HasPrefix(id, "profile:") {
		return nil
	}
	return c.OnSelect(strings.TrimPrefix(id, "profile:"))
}

type ProfileOption struct{ Key, Label string }
