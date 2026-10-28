// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"rectablet/src/workflows/swimming/domain"
)

func (s *Store) Feedback(person string) ([]domain.Feedback, error) {
	return s.feedbackByCompletion(person, false)
}

// CompletedFeedback returns assessed skills with three successful observations.
func (s *Store) CompletedFeedback(person string) ([]domain.Feedback, error) {
	return s.feedbackByCompletion(person, true)
}

func (s *Store) feedbackByCompletion(person string, complete bool) ([]domain.Feedback, error) {
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
			if p.Complete != complete {
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
			entry := domain.Feedback{Skill: skill, Version: version, Progress: p, Remaining: remaining}
			misses, e := rows(q, `SELECT r.subskill_id, COUNT(*) AS missed FROM results r JOIN attempts a ON a.id=r.attempt_id WHERE a.enrollment_id=? AND a.skill_id=? AND a.curriculum_version=? AND a.submitted_at IS NOT NULL AND r.status='not_yet' GROUP BY r.subskill_id`, person, skillID, version)
			if e != nil {
				return e
			}
			counts := map[string]int{}
			for _, miss := range misses {
				counts[miss.String("subskill_id")] = miss.Int("missed")
			}
			for _, subskill := range skill.Subskills {
				if count := counts[subskill.ID]; count > 0 {
					entry.MissedSubskills = append(entry.MissedSubskills, domain.SubskillMiss{Subskill: subskill, Count: count})
				}
			}
			if timed, timedErr := one(q, "SELECT laps_required,laps_completed,elapsed_ms FROM timed_attempts WHERE attempt_id=?", latest.String("id")); timedErr == nil {
				entry.Timed = true
				entry.TimedRequired = timed.Int("laps_required")
				entry.TimedLaps = timed.Int("laps_completed")
				entry.TimedElapsedMS = int64(timed.Int("elapsed_ms"))
			}
			out = append(out, entry)
		}
		return nil
	})
	return out, err
}
