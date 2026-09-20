package assessment

import (
	"strings"
	"unicode"

	"rectablet/internal/workflows/swimming/domain"
)

// Abbreviations are computed once for the whole session, not the shrinking queue.
// Students with matching first names get just enough surname letters to differ.
func abbreviateClassNames(students []domain.MultiAttempt) map[string]string {
	names := make([][]string, len(students))
	for i, student := range students {
		names[i] = strings.Fields(student.Name)
	}
	labels := make(map[string]string, len(students))
	for i, student := range students {
		parts := names[i]
		labels[student.PersonID] = strings.Join(parts, " ")
		if len(parts) < 2 {
			continue
		}
		first, surname := parts[0], parts[len(parts)-1]
		letters, prefix, fullName := []rune(surname), 1, false
		for j, other := range names {
			if i == j || len(other) < 2 || !strings.EqualFold(first, other[0]) {
				continue
			}
			otherSurname := other[len(other)-1]
			if strings.EqualFold(surname, otherSurname) {
				fullName = true
				break
			}
			common := 0
			otherLetters := []rune(otherSurname)
			for common < len(letters) && common < len(otherLetters) && unicode.ToLower(letters[common]) == unicode.ToLower(otherLetters[common]) {
				common++
			}
			prefix = max(prefix, common+1)
		}
		if !fullName {
			label := first + " " + string(letters[:min(prefix, len(letters))])
			if prefix < len(letters) {
				label += "."
			}
			labels[student.PersonID] = label
		}
	}
	return labels
}
