package curriculum

import (
	_ "embed"
	"encoding/json"
)

// Display metadata is deliberately separate from the versioned curriculum:
// changing a compact label must not migrate or reset saved assessments.
//
//go:embed compact_titles.json
var compactTitlesJSON []byte

var compactTitles = func() map[string]string {
	var titles map[string]string
	if err := json.Unmarshal(compactTitlesJSON, &titles); err != nil {
		panic(err)
	}
	return titles
}()

// CompactTitle returns presentation-only wording, never changing Skill.Name.
func CompactTitle(skill Skill) string {
	if title := compactTitles[skill.ID]; title != "" {
		return title
	}
	return skill.Name
}
