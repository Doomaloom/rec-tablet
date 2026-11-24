package curriculum

import (
	"encoding/json"
	"sort"
	"sync"
)

// Migration translates attempts recorded against an older catalog version into
// the current one. Old data is never rewritten: translation happens when
// progress is read, so every attempt keeps its own curriculum version.
//
// Rule (identical on the website): for an old attempt a of skill X and every
// new skill S with X in Skills[S]:
//   - each subskill y of S is successful when Subskills[y].From is non-empty
//     and at least one of those old ids has a result and every old id that has
//     a result is successful; when From is empty and FromSkillSuccess is set it
//     is successful iff a.Success; otherwise it is not_yet.
//   - the translated attempt succeeds when every subskill of S is successful
//     and, for the timed benchmark skills, a.Success as well.
type Migration struct {
	From      string                    `json:"from"`
	To        string                    `json:"to"`
	Skills    map[string][]string       `json:"skills"`
	Subskills map[string]SubskillOrigin `json:"subskills"`
	inverse   map[string][]string
}

// SubskillOrigin lists the old subskills a new subskill descends from.
type SubskillOrigin struct {
	From             []string `json:"from"`
	FromSkillSuccess bool     `json:"fromSkillSuccess"`
}

// OldAttempt is the part of a submitted attempt translation needs.
type OldAttempt struct {
	ID      string
	SkillID string
	Success bool
	Results map[string]string
}

// Translated is the view of an old attempt for one new skill. It is read-only:
// it never becomes a stored attempt and is never uploaded.
type Translated struct {
	SourceID string
	SkillID  string
	Success  bool
	Results  map[string]string
}

var timedSkills = map[string]bool{"Splash7:10": true, "Splash8:9": true, "Splash9:11": true}

// IsTimedSkill reports the timed benchmark skills, whose success also needs
// the old attempt's own success (the stopwatch result).
func IsTimedSkill(id string) bool { return timedSkills[id] }

func ParseMigration(data []byte) (Migration, error) {
	var m Migration
	if err := json.Unmarshal(data, &m); err != nil {
		return Migration{}, err
	}
	m.index()
	return m, nil
}

var (
	bundledMigration     Migration
	bundledMigrationOnce sync.Once
)

// BundledMigration returns the migration from the previous catalog to Bundled.
func BundledMigration() Migration {
	bundledMigrationOnce.Do(func() {
		data, err := assets.ReadFile("catalog-migration.json")
		if err != nil {
			panic(err)
		}
		if bundledMigration, err = ParseMigration(data); err != nil {
			panic(err)
		}
	})
	return bundledMigration
}

func (m *Migration) index() {
	m.inverse = map[string][]string{}
	for n, olds := range m.Skills {
		for _, o := range olds {
			m.inverse[o] = append(m.inverse[o], n)
		}
	}
	for _, v := range m.inverse {
		sort.Strings(v)
	}
}

// NewSkills returns the new skills an old skill id feeds, sorted by id.
func (m Migration) NewSkills(oldSkillID string) []string {
	if m.inverse == nil {
		m.index()
	}
	return append([]string(nil), m.inverse[oldSkillID]...)
}

// OldSkills returns the old skill ids feeding a new skill.
func (m Migration) OldSkills(newSkillID string) []string {
	return append([]string(nil), m.Skills[newSkillID]...)
}

// TranslateTo translates one old attempt for one new skill. It reports false
// when the attempt's skill does not feed that skill.
func (m Migration) TranslateTo(skill Skill, a OldAttempt) (Translated, bool) {
	feeds := false
	for _, o := range m.Skills[skill.ID] {
		if o == a.SkillID {
			feeds = true
			break
		}
	}
	if !feeds {
		return Translated{}, false
	}
	t := Translated{SourceID: a.ID, SkillID: skill.ID, Success: true, Results: map[string]string{}}
	for _, sub := range skill.Subskills {
		origin := m.Subskills[sub.ID]
		status := "not_yet"
		if len(origin.From) > 0 {
			seen, all := false, true
			for _, id := range origin.From {
				v, ok := a.Results[id]
				if !ok {
					continue
				}
				seen = true
				if v != "successful" {
					all = false
				}
			}
			if seen && all {
				status = "successful"
			}
		} else if origin.FromSkillSuccess && a.Success {
			status = "successful"
		}
		t.Results[sub.ID] = status
		if status != "successful" {
			t.Success = false
		}
	}
	if len(skill.Subskills) == 0 || (IsTimedSkill(skill.ID) && !a.Success) {
		t.Success = false
	}
	return t, true
}

// Translate returns the translation of an old attempt for every new skill of
// the catalog it feeds, ordered by new skill id.
func (m Migration) Translate(c Catalog, a OldAttempt) []Translated {
	out := []Translated{}
	for _, id := range m.NewSkills(a.SkillID) {
		skill, err := c.Skill(id)
		if err != nil {
			continue
		}
		if t, ok := m.TranslateTo(skill, a); ok {
			out = append(out, t)
		}
	}
	return out
}
