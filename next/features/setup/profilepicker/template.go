package profilepicker

import (
	"rectablet/next/contracts"
)

func (c *ProfilePicker) Rows(profiles []ProfileOption) []contracts.Row {
	rows := make([]contracts.Row, 0, len(profiles))
	for _, profile := range profiles {
		rows = append(rows, contracts.Row{ID: "profile:" + profile.Key, Label: "Switch Session · " + profile.Label})
	}
	return rows
}
