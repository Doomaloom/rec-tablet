// Package curriculum defines the versioned, data-driven assessment catalog.
package curriculum

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

//go:embed catalog.json
var assets embed.FS

type Catalog struct {
	SchemaVersion int     `json:"schemaVersion"`
	ProgramID     string  `json:"programId"`
	Version       string  `json:"version"`
	Levels        []Level `json:"levels"`
}
type Level struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Source       string   `json:"source"`
	SourceSHA256 string   `json:"sourceSha256"`
	Skills       []Skill  `json:"skills"`
	Kind         string   `json:"kind"`
	References   []string `json:"references,omitempty"`
}
type Skill struct {
	ID          string     `json:"id"`
	Number      string     `json:"number"`
	Name        string     `json:"name"`
	SourceLines []string   `json:"sourceLines"`
	Subskills   []Subskill `json:"subskills"`
}
type Subskill struct {
	ID    string   `json:"id"`
	Text  string   `json:"text"`
	Group []string `json:"group"`
}

func Bundled() Catalog {
	data, _ := assets.ReadFile("catalog.json")
	var c Catalog
	if err := json.Unmarshal(data, &c); err != nil {
		panic(err)
	}
	return c
}
func (c Catalog) Level(id string) (Level, bool) {
	for _, l := range c.Levels {
		if l.ID == id {
			return l, true
		}
	}
	return Level{}, false
}
func (c Catalog) Skill(id string) (Skill, error) {
	for _, l := range c.Levels {
		for _, s := range l.Skills {
			if s.ID == id {
				return s, nil
			}
		}
	}
	return Skill{}, fmt.Errorf("missing curriculum for this skill; import a supported swimmer level")
}
func (c Catalog) NormalizeLevel(value string) string {
	normalize := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, s)
	}
	for _, l := range c.Levels {
		if normalize(value) == normalize(l.ID) || normalize(value) == normalize(l.Name) {
			return l.ID
		}
	}
	return value
}
