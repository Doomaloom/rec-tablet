package repository

import (
	"strings"

	"rectablet/src/workflows/swimming/curriculum"
)

// translatedAttempt is a submitted attempt from the previous catalog version
// seen through the current catalog's skill. It is derived on read and never
// stored, edited or uploaded; SourceID is the stored attempt it comes from.
type translatedAttempt struct {
	curriculum.Translated
	Date, CreatedAt, SubmittedAt string
}

// legacySkills returns the previous-version skill ids whose attempts count for
// a skill of the given version. It is empty unless that version is the bundled
// catalog with a migration from the previous one.
func (s *Store) legacySkills(skill, version string) []string {
	if s.migration.To == "" || version != s.migration.To {
		return nil
	}
	return s.migration.OldSkills(skill)
}

// translatedAttempts returns the previous-version attempts that count toward a
// current skill, newest data unchanged in the database.
func (s *Store) translatedAttempts(q queryer, person, skill, version string) ([]translatedAttempt, error) {
	olds := s.legacySkills(skill, version)
	if len(olds) == 0 {
		return nil, nil
	}
	catalog, err := s.catalog(q, version)
	if err != nil {
		return nil, err
	}
	target, err := catalog.Skill(skill)
	if err != nil {
		return nil, nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(olds)), ",")
	args := []any{person, s.migration.From}
	for _, o := range olds {
		args = append(args, o)
	}
	rs, err := rows(q, `SELECT a.*,l.date FROM attempts a JOIN lessons l ON l.id=a.lesson_id WHERE a.enrollment_id=? AND a.curriculum_version=? AND a.skill_id IN (`+marks+`) AND a.submitted_at IS NOT NULL ORDER BY l.date DESC,a.created_at DESC,a.rowid DESC`, args...)
	if err != nil {
		return nil, err
	}
	out := []translatedAttempt{}
	for _, r := range rs {
		values, err := answers(q, r.String("id"))
		if err != nil {
			return nil, err
		}
		t, ok := s.migration.TranslateTo(target, curriculum.OldAttempt{ID: r.String("id"), SkillID: r.String("skill_id"), Success: r.Int("success") != 0, Results: values})
		if !ok {
			continue
		}
		out = append(out, translatedAttempt{Translated: t, Date: r.String("date"), CreatedAt: r.String("created_at"), SubmittedAt: r.String("submitted_at")})
	}
	return out, nil
}
