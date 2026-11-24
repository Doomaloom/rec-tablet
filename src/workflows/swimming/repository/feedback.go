// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"errors"
	"sort"

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
		// Attempts of the previous catalog are reported under the current skills
		// they translate to, so each skill appears once with its combined history.
		type target struct{ skill, version string }
		seen := map[target]bool{}
		targets := []target{}
		add := func(t target) {
			if !seen[t] {
				seen[t] = true
				targets = append(targets, t)
			}
		}
		for _, pair := range pairs {
			skillID, version := pair.String("skill_id"), pair.String("curriculum_version")
			if s.migration.To != "" && version == s.migration.From {
				for _, n := range s.migration.NewSkills(skillID) {
					add(target{n, s.migration.To})
				}
				continue
			}
			add(target{skillID, version})
		}
		sort.SliceStable(targets, func(i, j int) bool {
			if targets[i].version != targets[j].version {
				return targets[i].version < targets[j].version
			}
			return targets[i].skill < targets[j].skill
		})
		for _, t := range targets {
			skillID, version := t.skill, t.version
			p, e := s.progress(q, person, skillID, version)
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
			translated, e := s.translatedAttempts(q, person, skillID, version)
			if e != nil {
				return e
			}
			// The newest attempt is the stored one or a translated one, whichever
			// is later by lesson date then creation time (stored wins ties).
			latestID, latestDate, latestCreated := "", "", ""
			var values map[string]string
			latest, e := one(q, `SELECT a.*,l.date FROM attempts a JOIN lessons l ON l.id=a.lesson_id WHERE a.enrollment_id=? AND a.skill_id=? AND a.curriculum_version=? AND a.submitted_at IS NOT NULL ORDER BY l.date DESC,a.created_at DESC,a.rowid DESC LIMIT 1`, person, skillID, version)
			if e != nil && !errors.Is(e, ErrNotFound) {
				return e
			}
			if e == nil {
				latestID, latestDate, latestCreated = latest.String("id"), latest.String("date"), latest.String("created_at")
				if values, e = answers(q, latestID); e != nil {
					return e
				}
			}
			for _, tr := range translated {
				if latestID == "" || tr.Date+"\x00"+tr.CreatedAt > latestDate+"\x00"+latestCreated {
					latestID, latestDate, latestCreated = tr.SourceID, tr.Date, tr.CreatedAt
					values = tr.Results
				}
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
			for _, tr := range translated {
				for id, status := range tr.Results {
					if status == "not_yet" {
						counts[id]++
					}
				}
			}
			for _, subskill := range skill.Subskills {
				if count := counts[subskill.ID]; count > 0 {
					entry.MissedSubskills = append(entry.MissedSubskills, domain.SubskillMiss{Subskill: subskill, Count: count})
				}
			}
			if timed, timedErr := one(q, "SELECT laps_required,laps_completed,elapsed_ms FROM timed_attempts WHERE attempt_id=?", latestID); timedErr == nil {
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
