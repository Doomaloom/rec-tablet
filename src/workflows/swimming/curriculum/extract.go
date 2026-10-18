package curriculum

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"

	"golang.org/x/net/html"
)

var headingPattern = regexp.MustCompile(`^([0-9]+[a-z]?)\s*\.\s*(.+)`)
var displayPattern = regexp.MustCompile(`([a-z])([A-Z0-9])`)

func clean(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\x1e", " ")), " ")
}
func attr(t html.Token, key string) string {
	for _, a := range t.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func hasClass(t html.Token, class string) bool {
	for _, c := range strings.Fields(attr(t, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

type section struct {
	number, name string
	lines        []string
}

func Extract(path string) (Level, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return Level{}, e
	}
	id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	hash := sha256.Sum256(raw)
	l := Level{ID: id, Name: displayPattern.ReplaceAllString(id, "${1} ${2}"), Source: filepath.Base(path), SourceSHA256: fmt.Sprintf("%x", hash), Skills: []Skill{}, Kind: "curriculum"}
	z := html.NewTokenizer(bytes.NewReader(raw))
	after, inParagraph := false, false
	boldDepth := 0
	var body, bold strings.Builder
	headings := [][2]string{}
	privateBlock, privateStrong := false, false
	var reference strings.Builder
	finish := func() error {
		if !inParagraph {
			return nil
		}
		inParagraph = false
		sections := []section{}
		for _, rawLine := range strings.Split(body.String(), "\x1e") {
			line := clean(rawLine)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "__SKILL_") {
				var index int
				if _, err := fmt.Sscanf(line, "__SKILL_%d__", &index); err != nil || index >= len(headings) {
					return fmt.Errorf("%s: invalid heading", id)
				}
				sections = append(sections, section{headings[index][0], headings[index][1], []string{}})
			} else if len(sections) > 0 {
				last := &sections[len(sections)-1]
				last.lines = append(last.lines, line)
			}
		}
		for _, part := range sections {
			if id == "Splash7" && part.number == "16" && len(part.lines) > 0 {
				part.name += " " + part.lines[0]
				part.lines = part.lines[1:]
			}
			skillID := id + ":" + part.number
			duplicate := 0
			for _, s := range l.Skills {
				if s.Number == part.number {
					duplicate++
				}
			}
			if duplicate > 0 {
				skillID += fmt.Sprintf("-%d", duplicate+1)
			}
			s := Skill{ID: skillID, Number: part.number, Name: part.name, SourceLines: part.lines, Subskills: []Subskill{}}
			group := []string{}
			groupMarker := ""
			newGroup := false
			waterSmart := strings.Contains(strings.ToLower(part.name), "water smart")
			plainGroups := map[string]bool{}
			for _, name := range []string{"front crawl", "back crawl", "breaststroke", "warm-up:", "work set:", "cool down:", "swim with a buddy", "wear a lifejacket", "check the ice", "swim to survive", "within arms' reach", "forward roll entry wearing pfd", "front crawl wearing pfd (5 m)", "back crawl wearing pfd (5 m)"} {
				plainGroups[name] = true
			}
			markerOf := func(line string) string {
				if strings.HasPrefix(line, "•") {
					return "•"
				}
				if strings.HasPrefix(line, "-") {
					return "-"
				}
				return ""
			}
			for i, line := range part.lines {
				marker := markerOf(line)
				label := line
				if marker != "" {
					label = clean(strings.TrimPrefix(line, marker))
				}
				nextMarker := ""
				if i+1 < len(part.lines) {
					nextMarker = markerOf(part.lines[i+1])
				}
				if plainGroups[strings.ToLower(label)] || (marker == "•" && nextMarker == "-" && !waterSmart) {
					group = []string{label}
					groupMarker = marker
					newGroup = true
					continue
				}
				if marker == "" && len(s.Subskills) > 0 && !newGroup && !strings.HasPrefix(strings.ToLower(label), "participation") {
					s.Subskills[len(s.Subskills)-1].Text += " " + label
					continue
				}
				if marker != "" && marker == groupMarker && !waterSmart {
					group = []string{}
					groupMarker = ""
				}
				s.Subskills = append(s.Subskills, Subskill{ID: fmt.Sprintf("%s:%d", skillID, len(s.Subskills)+1), Text: label, Group: append([]string{}, group...)})
				newGroup = false
			}
			if len(s.Subskills) == 0 {
				return fmt.Errorf("%s: missing assessable criteria", skillID)
			}
			for _, c := range s.Subskills {
				if c.Text == "" {
					return fmt.Errorf("%s: empty criterion", skillID)
				}
			}
			l.Skills = append(l.Skills, s)
		}
		return nil
	}
	for {
		kind := z.Next()
		if kind == html.ErrorToken {
			if e = z.Err(); e != io.EOF {
				return l, e
			}
			break
		}
		t := z.Token()
		if id == "SplashPrivate" {
			if kind == html.StartTagToken && t.Data == "div" && hasClass(t, "private-skill-block") {
				privateBlock = true
			}
			if privateBlock && kind == html.StartTagToken && t.Data == "strong" {
				privateStrong = true
				reference.Reset()
			}
			if privateStrong && kind == html.TextToken {
				reference.WriteString(t.Data)
			}
			if privateStrong && kind == html.EndTagToken && t.Data == "strong" {
				l.References = append(l.References, clean(reference.String()))
				privateStrong = false
			}
			if privateBlock && kind == html.EndTagToken && t.Data == "div" {
				privateBlock = false
			}
		}
		if kind == html.StartTagToken && hasClass(t, "break-before-page") {
			after = true
		}
		if !after {
			continue
		}
		if kind == html.StartTagToken && t.Data == "p" {
			if e = finish(); e != nil {
				return l, e
			}
			inParagraph = true
			body.Reset()
			bold.Reset()
			boldDepth = 0
			headings = nil
			continue
		}
		if kind == html.EndTagToken && t.Data == "p" {
			if e = finish(); e != nil {
				return l, e
			}
			continue
		}
		if !inParagraph {
			continue
		}
		if kind == html.StartTagToken && (t.Data == "b" || t.Data == "strong") {
			if boldDepth == 0 {
				bold.Reset()
			}
			boldDepth++
			continue
		}
		if kind == html.EndTagToken && (t.Data == "b" || t.Data == "strong") && boldDepth > 0 {
			boldDepth--
			if boldDepth == 0 {
				value := bold.String()
				match := headingPattern.FindStringSubmatch(clean(value))
				if len(match) > 0 {
					headings = append(headings, [2]string{match[1], match[2]})
					fmt.Fprintf(&body, "\x1e__SKILL_%d__\x1e", len(headings)-1)
				} else {
					body.WriteString(value)
				}
			}
			continue
		}
		value := ""
		if kind == html.TextToken {
			value = t.Data
		}
		if (kind == html.StartTagToken || kind == html.SelfClosingTagToken) && t.Data == "br" {
			value = "\x1e"
		}
		if boldDepth > 0 {
			bold.WriteString(value)
		} else {
			body.WriteString(value)
		}
	}
	if e = finish(); e != nil {
		return l, e
	}
	if id == "SplashPrivate" {
		l.Kind = "level-reference"
		return l, nil
	}
	if len(l.Skills) == 0 {
		return l, fmt.Errorf("%s: no skills", id)
	}
	return l, nil
}

// canonicalPython preserves the prior extractor's version hash (sort_keys=True,
// ensure_ascii=True, default JSON separators), independent of Go struct ordering.
func canonicalPython(value any) string {
	switch v := value.(type) {
	case map[string]any:
		keys := []string{}
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := []string{}
		for _, k := range keys {
			parts = append(parts, canonicalPython(k)+": "+canonicalPython(v[k]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []any:
		parts := []string{}
		for _, x := range v {
			parts = append(parts, canonicalPython(x))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case string:
		var b strings.Builder
		b.WriteByte('"')
		for _, r := range v {
			if r > 127 {
				if r > 0xffff {
					a, c := utf16.EncodeRune(r)
					fmt.Fprintf(&b, `\u%04x\u%04x`, a, c)
				} else {
					fmt.Fprintf(&b, `\u%04x`, r)
				}
			} else {
				switch r {
				case '"':
					b.WriteString(`\"`)
				case '\\':
					b.WriteString(`\\`)
				case '\n':
					b.WriteString(`\n`)
				case '\r':
					b.WriteString(`\r`)
				case '\t':
					b.WriteString(`\t`)
				case '\b':
					b.WriteString(`\b`)
				case '\f':
					b.WriteString(`\f`)
				default:
					if r < 32 {
						fmt.Fprintf(&b, `\u%04x`, r)
					} else {
						b.WriteRune(r)
					}
				}
			}
		}
		b.WriteByte('"')
		return b.String()
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}
func ExtractDirectory(dir string) (Catalog, error) {
	paths, e := filepath.Glob(filepath.Join(dir, "*.html"))
	if e != nil {
		return Catalog{}, e
	}
	if len(paths) == 0 {
		return Catalog{}, fmt.Errorf("no HTML templates in %s", dir)
	}
	sort.Strings(paths)
	levels := []Level{}
	for _, p := range paths {
		l, e := Extract(p)
		if e != nil {
			return Catalog{}, e
		}
		levels = append(levels, l)
	}
	raw, _ := json.Marshal(levels)
	var generic any
	_ = json.Unmarshal(raw, &generic)
	hash := sha256.Sum256([]byte(canonicalPython(generic)))
	return Catalog{SchemaVersion: 1, ProgramID: "aquatics", Version: fmt.Sprintf("templates-%x", hash[:8]), Levels: levels}, nil
}
