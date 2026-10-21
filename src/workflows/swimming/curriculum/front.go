package curriculum

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type frontSkill struct{ number, name string }

// frontSkills reads the numbered marking columns, never the back-page headings.
func frontSkills(raw []byte) ([]frontSkill, error) {
	z := html.NewTokenizer(bytes.NewReader(raw))
	var result []frontSkill
	var label strings.Builder
	inCell := false
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			if z.Err() != io.EOF {
				return nil, z.Err()
			}
			break
		}
		t := z.Token()
		if kind == html.StartTagToken && hasClass(t, "break-before-page") {
			break
		}
		if kind == html.StartTagToken && t.Data == "td" && hasClass(t, "rotate-cell") {
			inCell = true
			label.Reset()
		} else if inCell && kind == html.TextToken {
			label.WriteString(t.Data)
		} else if inCell && kind == html.EndTagToken && t.Data == "td" {
			if match := headingPattern.FindStringSubmatch(clean(label.String())); len(match) > 0 {
				result = append(result, frontSkill{match[1], match[2]})
			}
			inCell = false
		}
	}
	return result, nil
}

var numericPart = regexp.MustCompile(`^[0-9]+`)
var componentSlug = regexp.MustCompile(`[^a-z0-9]+`)
var positionWord = regexp.MustCompile(`\b(front|back|side)\b`)

// These components already had stable internal IDs. The displayed number and
// name still come directly from the front page, including its a/b/c suffixes.
var legacySplitBases = map[string]bool{
	"Splash1:10": true, "Splash2A:8": true, "Splash2B:8": true,
	"Splash3:10": true, "Splash4:7": true, "Splash5:8": true, "Splash6:8": true,
	"Splash7:8": true, "Splash8:7": true, "Splash9:9": true,
	"TeenAdult1:12": true, "TeenAdult1:15": true, "TeenAdult2:14": true, "TeenAdult3:9": true,
}

func componentFor(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "water smart") {
		for _, message := range []string{"within arms", "wear a lifejacket", "swim with a buddy", "check the ice", "swim to survive"} {
			if strings.Contains(lower, message) {
				return message
			}
		}
	}
	if strings.Contains(lower, "breaststroke") {
		return "breaststroke"
	}
	return positionWord.FindString(lower)
}

func matchesComponent(text, component string) bool {
	lower := strings.ToLower(text)
	if component == "front" || component == "back" || component == "side" {
		return positionWord.FindString(lower) == component
	}
	return strings.Contains(lower, component)
}

// Extract uses front-page columns for skill identity and back-page requirements
// for assessment. A shared back-page requirement is copied to each relevant
// column; requirements for another stroke, position or message are excluded.
func Extract(path string) (Level, error) {
	back, err := extractBack(path)
	if err != nil {
		return back, err
	}
	if back.Kind == "level-reference" {
		return back, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return back, err
	}
	columns, err := frontSkills(raw)
	if err != nil {
		return back, err
	}
	if len(columns) == 0 {
		return back, fmt.Errorf("%s: no front-page skill columns", back.ID)
	}
	sources := make([]Skill, len(columns))
	uses := map[string]int{}
	for i, column := range columns {
		sourceID := back.ID + ":" + column.number
		// 8c is the second skill labelled 8 on Splash2B's back page.
		if back.ID == "Splash2B" && column.number == "8c" {
			sourceID = "Splash2B:8-2"
		}
		var source Skill
		for _, candidate := range back.Skills {
			if candidate.ID == sourceID {
				source = candidate
				break
			}
		}
		if source.ID == "" {
			sourceID = back.ID + ":" + numericPart.FindString(column.number)
			for _, candidate := range back.Skills {
				if candidate.ID == sourceID {
					source = candidate
					break
				}
			}
		}
		if source.ID == "" {
			return back, fmt.Errorf("%s:%s: no back-page requirements", back.ID, column.number)
		}
		sources[i] = source
		uses[source.ID]++
	}
	result := back
	result.Skills = []Skill{}
	used := map[string]bool{}
	seen := map[string]int{}
	for i, column := range columns {
		source := sources[i]
		component := componentFor(column.name)
		skill := Skill{ID: back.ID + ":" + column.number, Number: column.number, Name: column.name, SourceSkillID: source.ID, SourceName: source.Name, SourceLines: source.SourceLines, Subskills: []Subskill{}}
		if uses[source.ID] > 1 && legacySplitBases[source.ID] {
			for _, sub := range source.Subskills {
				if len(sub.Group) > 0 && matchesComponent(sub.Group[0], component) {
					skill.ID = source.ID + "-" + strings.Trim(componentSlug.ReplaceAllString(strings.ToLower(sub.Group[0]), "-"), "-")
					break
				}
			}
		}
		seen[skill.ID]++
		if seen[skill.ID] > 1 {
			skill.ID += fmt.Sprintf("-%d", seen[skill.ID])
		}
		for j, sub := range source.Subskills {
			include := true
			if uses[source.ID] > 1 {
				if back.ID == "Splash2B" && source.Number == "2" {
					// PFD fit is common; the two entry descriptions belong to their own column.
					include = j == 0 || (column.number == "2a" && j == 1) || (column.number == "2b" && j == 2)
				} else if len(sub.Group) > 0 && component != "" {
					group := sub.Group[0]
					// Position/Water Smart groups are specific; generic body-position
					// groups remain shared and their leaf text selects front or back.
					groupComponent := componentFor(group)
					if groupComponent != "" || strings.Contains(strings.ToLower(column.name), "water smart") {
						include = matchesComponent(group, component)
					} else if positionWord.FindString(strings.ToLower(sub.Text)) != "" {
						include = matchesComponent(sub.Text, component)
					}
				} else if component == "front" || component == "back" || component == "side" {
					lower := strings.ToLower(sub.Text)
					// Explicit position-specific leaves are selected. Wording about both
					// front and back (e.g. recovery) remains a shared source requirement.
					words := positionWord.FindAllString(lower, -1)
					distinct := map[string]bool{}
					for _, word := range words {
						distinct[word] = true
					}
					if len(distinct) == 1 {
						include = matchesComponent(sub.Text, component)
					}
				}
			}
			if include {
				used[sub.ID] = true
				sub.ID = fmt.Sprintf("%s:%d", skill.ID, len(skill.Subskills)+1)
				skill.Subskills = append(skill.Subskills, sub)
			}
		}
		if len(skill.Subskills) == 0 {
			return back, fmt.Errorf("%s: no matching back-page subskills", skill.ID)
		}
		result.Skills = append(result.Skills, skill)
	}
	// Keep unmatched source requirements visible in provenance rather than
	// assigning them to an unrelated front-page skill or silently dropping them.
	for _, source := range back.Skills {
		for _, sub := range source.Subskills {
			if !used[sub.ID] {
				result.UnmappedSubskills = append(result.UnmappedSubskills, sub)
			}
		}
	}
	return result, nil
}
