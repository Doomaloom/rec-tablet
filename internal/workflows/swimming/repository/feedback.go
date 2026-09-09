// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"rectablet/internal/workflows/swimming/domain"
)

func (s *Store) Feedback(person string) ([]domain.Feedback, error) {
	out := []domain.Feedback{}
	err := s.transaction(func(q queryer) error {
		pairs, e := rows(q, `SELECT DISTINCT skill_id,curriculum_version FROM attempts WHERE enrollment_id=? AND submitted_at IS NOT NULL ORDER BY curriculum_version,skill_id`, person)
		if e != nil {
			return e
		}
		for _, pair := range pairs {
			skillID, version := pair.String("skill_id"), pair.String("curriculum_version")
			p, e := progress(q, person, skillID, version)
			if e != nil {
				return e
			}
			if p.Complete {
				continue
			}
			skill, e := s.skill(q, version, skillID)
			if e != nil {
				return e
			}
			latest, e := one(q, `SELECT a.* FROM attempts a JOIN lessons l ON l.id=a.lesson_id WHERE a.enrollment_id=? AND a.skill_id=? AND a.curriculum_version=? AND a.submitted_at IS NOT NULL ORDER BY l.date DESC,a.created_at DESC,a.rowid DESC LIMIT 1`, person, skillID, version)
			if e != nil {
				return e
			}
			values, e := answers(q, latest.String("id"))
			if e != nil {
				return e
			}
			remaining := domain.Remaining(skill, values)
			out = append(out, domain.Feedback{Skill: skill, Version: version, Progress: p, Remaining: remaining})
		}
		return nil
	})
	return out, err
}
