// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"encoding/json"
	"errors"

	"rectablet/internal/workflows/swimming/domain"
)

func progress(q queryer, person, skill, version string) (domain.Progress, error) {
	r, e := one(q, `SELECT COUNT(*) AS assessed,COALESCE(SUM(success),0) AS successes FROM attempts WHERE enrollment_id=? AND skill_id=? AND curriculum_version=? AND submitted_at IS NOT NULL`, person, skill, version)
	if e != nil {
		return domain.Progress{}, e
	}
	return domain.CalculateProgress(r.Int("assessed"), r.Int("successes")), nil
}
func (s *Store) Progress(person, skill, version string) (domain.Progress, error) {
	return progress(s.DB, person, skill, version)
}
func (s *Store) attemptRecords(person, skill, version string) ([]record, error) {
	return s.queryRows(`SELECT a.*,l.date,EXISTS(SELECT 1 FROM drafts d WHERE d.attempt_id=a.id) AS editing FROM attempts a JOIN lessons l ON l.id=a.lesson_id WHERE a.enrollment_id=? AND a.skill_id=? AND a.curriculum_version=? ORDER BY a.created_at DESC,a.rowid DESC`, person, skill, version)
}
func (s *Store) NewAttempt(person, lesson, skill string) (string, error) {
	id := ""
	err := s.transaction(func(q queryer) error {
		p, e := checkLesson(q, person, lesson)
		if e != nil {
			return e
		}
		c, e := s.catalog(q, p.String("curriculum_version"))
		if e != nil {
			return e
		}
		if e = domain.ValidateNewAttempt(c, p.String("level"), skill, p.Int("active") != 0); e != nil {
			return e
		}
		r, e := one(q, `SELECT id FROM attempts WHERE enrollment_id=? AND lesson_id=? AND skill_id=? AND curriculum_version=? AND submitted_at IS NULL`, person, lesson, skill, p.String("curriculum_version"))
		if e == nil {
			id = r.String("id")
			return nil
		}
		if !errors.Is(e, ErrNotFound) {
			return e
		}
		id = ID()
		if _, e = q.Exec("INSERT INTO attempts VALUES (?,?,?,?,?,?,NULL,0)", id, person, lesson, skill, p.String("curriculum_version"), Now()); e != nil {
			return e
		}
		_, e = q.Exec("INSERT INTO drafts VALUES (?,?,?)", id, "{}", Now())
		return e
	})
	return id, err
}

// OpenMultiAttempts prepares one draft for every selected swimmer atomically.
// Existing unfinished drafts are reused and returned with a copy of their
// original answers so a later discard can restore them.
func (s *Store) OpenMultiAttempts(people []string, lesson, skill string) ([]domain.MultiAttempt, error) {
	if len(people) == 0 {
		return nil, errors.New("select at least one swimmer")
	}
	opened := make([]domain.MultiAttempt, 0, len(people))
	seen := map[string]bool{}
	err := s.transaction(func(q queryer) error {
		for _, person := range people {
			if seen[person] {
				return errors.New("a swimmer was selected more than once")
			}
			seen[person] = true
			p, e := checkLesson(q, person, lesson)
			if e != nil {
				return e
			}
			catalog, e := s.catalog(q, p.String("curriculum_version"))
			if e != nil {
				return e
			}
			if e = domain.ValidateNewAttempt(catalog, p.String("level"), skill, p.Int("active") != 0); e != nil {
				return e
			}
			attempt, e := one(q, `SELECT id FROM attempts WHERE enrollment_id=? AND lesson_id=? AND skill_id=? AND curriculum_version=? AND submitted_at IS NULL`, person, lesson, skill, p.String("curriculum_version"))
			created := false
			attemptID := ""
			answers := map[string]string{}
			if e == nil {
				attemptID = attempt.String("id")
				answers, e = draft(q, attemptID)
				if e != nil {
					return e
				}
			} else if errors.Is(e, ErrNotFound) {
				created = true
				attemptID = ID()
				if _, e = q.Exec("INSERT INTO attempts VALUES (?,?,?,?,?,?,NULL,0)", attemptID, person, lesson, skill, p.String("curriculum_version"), Now()); e != nil {
					return e
				}
				if _, e = q.Exec("INSERT INTO drafts VALUES (?,?,?)", attemptID, "{}", Now()); e != nil {
					return e
				}
			} else {
				return e
			}
			opened = append(opened, domain.MultiAttempt{AttemptID: attemptID, PersonID: person, Name: p.String("name"), OriginalAnswers: cloneAnswers(answers), Created: created})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return opened, nil
}

func cloneAnswers(values map[string]string) map[string]string {
	copyValues := map[string]string{}
	for key, value := range values {
		copyValues[key] = value
	}
	return copyValues
}

// NewTimedHeat creates one timed attempt per selected swimmer and initializes
// the shared lap target. The whole heat is created atomically.
func (s *Store) NewTimedHeat(people []string, lesson, skill string, laps int) ([]domain.TimedParticipant, error) {
	if len(people) == 0 {
		return nil, errors.New("select at least one swimmer")
	}
	if laps < 1 || laps > 999 {
		return nil, errors.New("laps must be between 1 and 999")
	}
	participants := []domain.TimedParticipant{}
	err := s.transaction(func(q queryer) error {
		for _, person := range people {
			p, e := checkLesson(q, person, lesson)
			if e != nil {
				return e
			}
			if p.Int("active") == 0 {
				return errors.New("this swimmer is no longer enrolled")
			}
			catalog, e := s.catalog(q, p.String("curriculum_version"))
			if e != nil {
				return e
			}
			if e = domain.ValidateNewAttempt(catalog, p.String("level"), skill, true); e != nil {
				return e
			}
			_, e = one(q, `SELECT id FROM attempts WHERE enrollment_id=? AND lesson_id=? AND skill_id=? AND curriculum_version=? AND submitted_at IS NULL`, person, lesson, skill, p.String("curriculum_version"))
			if e == nil {
				return errors.New("this swimmer already has an unfinished attempt")
			}
			if !errors.Is(e, ErrNotFound) {
				return e
			}
			attemptID := ID()
			if _, e = q.Exec("INSERT INTO attempts VALUES (?,?,?,?,?,?,NULL,0)", attemptID, person, lesson, skill, p.String("curriculum_version"), Now()); e != nil {
				return e
			}
			if _, e = q.Exec("INSERT INTO drafts VALUES (?,?,?)", attemptID, "{}", Now()); e != nil {
				return e
			}
			if _, e = q.Exec("INSERT INTO timed_attempts(attempt_id,laps_required) VALUES (?,?)", attemptID, laps); e != nil {
				return e
			}
			participants = append(participants, domain.TimedParticipant{AttemptID: attemptID, EnrollmentID: p.String("id"), PersonID: person, Name: p.String("name"), LapsRequired: laps})
		}
		return nil
	})
	return participants, err
}

// RecordTimedLap advances one participant up to the configured lap target.
func (s *Store) RecordTimedLap(attemptID string) (int, error) {
	completed := 0
	err := s.transaction(func(q queryer) error {
		r, e := one(q, "SELECT laps_completed,laps_required FROM timed_attempts WHERE attempt_id=?", attemptID)
		if e != nil {
			return e
		}
		completed = r.Int("laps_completed")
		if completed >= r.Int("laps_required") {
			return nil
		}
		completed++
		_, e = q.Exec("UPDATE timed_attempts SET laps_completed=? WHERE attempt_id=?", completed, attemptID)
		return e
	})
	return completed, err
}

// ResetTimedHeat clears lap progress before a heat is started again.
func (s *Store) ResetTimedHeat(attemptIDs []string) error {
	return s.transaction(func(q queryer) error {
		for _, id := range attemptIDs {
			if _, e := q.Exec("UPDATE timed_attempts SET laps_completed=0,elapsed_ms=0 WHERE attempt_id=?", id); e != nil {
				return e
			}
		}
		return nil
	})
}

// SaveTimedHeat submits every participant with the measured elapsed time.
// A participant passes only when the caller marks it successful after all
// required laps were completed within the benchmark limit.
func (s *Store) SaveTimedHeat(results []domain.TimedResult) error {
	if len(results) == 0 {
		return errors.New("heat has no swimmers")
	}
	return s.transaction(func(q queryer) error {
		for _, result := range results {
			if result.LapsRequired < 1 || result.LapsCompleted < 0 || result.LapsCompleted > result.LapsRequired || result.ElapsedMS < 0 {
				return errors.New("invalid timed result")
			}
			a, e := one(q, "SELECT * FROM attempts WHERE id=? AND submitted_at IS NULL", result.AttemptID)
			if e != nil {
				return e
			}
			if _, e = one(q, "SELECT attempt_id FROM timed_attempts WHERE attempt_id=? AND laps_required=?", result.AttemptID, result.LapsRequired); e != nil {
				return e
			}
			skill, e := s.skill(q, a.String("curriculum_version"), a.String("skill_id"))
			if e != nil {
				return e
			}
			status := "not_yet"
			if result.Success {
				status = "successful"
			}
			if _, e = q.Exec("DELETE FROM results WHERE attempt_id=?", result.AttemptID); e != nil {
				return e
			}
			for _, subskill := range skill.Subskills {
				if _, e = q.Exec("INSERT INTO results VALUES (?,?,?)", result.AttemptID, subskill.ID, status); e != nil {
					return e
				}
			}
			if _, e = q.Exec("UPDATE timed_attempts SET laps_completed=?,elapsed_ms=? WHERE attempt_id=?", result.LapsCompleted, result.ElapsedMS, result.AttemptID); e != nil {
				return e
			}
			success := 0
			if result.Success {
				success = 1
			}
			if _, e = q.Exec("UPDATE attempts SET success=?,submitted_at=? WHERE id=?", success, Now(), result.AttemptID); e != nil {
				return e
			}
			if _, e = q.Exec("DELETE FROM drafts WHERE attempt_id=?", result.AttemptID); e != nil {
				return e
			}
		}
		return nil
	})
}

// DiscardTimedHeat removes all unfinished attempts created for a heat.
func (s *Store) DiscardTimedHeat(attemptIDs []string) error {
	return s.transaction(func(q queryer) error {
		for _, id := range attemptIDs {
			if _, e := one(q, "SELECT attempt_id FROM timed_attempts WHERE attempt_id=?", id); e != nil {
				return e
			}
			for _, table := range []string{"revisions", "results", "drafts", "timed_attempts"} {
				if _, e := q.Exec("DELETE FROM "+table+" WHERE attempt_id=?", id); e != nil {
					return e
				}
			}
			if _, e := q.Exec("DELETE FROM attempts WHERE id=? AND submitted_at IS NULL", id); e != nil {
				return e
			}
		}
		return nil
	})
}
func answers(q queryer, id string) (map[string]string, error) {
	rs, e := rows(q, "SELECT * FROM results WHERE attempt_id=?", id)
	out := map[string]string{}
	for _, r := range rs {
		out[r.String("subskill_id")] = r.String("status")
	}
	return out, e
}
func draft(q queryer, id string) (map[string]string, error) {
	r, e := one(q, "SELECT answers FROM drafts WHERE attempt_id=?", id)
	if e != nil {
		return nil, e
	}
	out := map[string]string{}
	e = json.Unmarshal([]byte(r.String("answers")), &out)
	return out, e
}
func (s *Store) Draft(id string) (map[string]string, error) { return draft(s.DB, id) }

// ReplaceDraftAnswers updates a pending student's entire assessment atomically.
// It is deliberately separate from correction editing of submitted attempts.
func (s *Store) ReplaceDraftAnswers(id string, values map[string]string) error {
	return s.transaction(func(q queryer) error {
		a, err := one(q, "SELECT * FROM attempts WHERE id=? AND submitted_at IS NULL", id)
		if err != nil {
			return err
		}
		skill, err := s.skill(q, a.String("curriculum_version"), a.String("skill_id"))
		if err != nil {
			return err
		}
		answers := make(map[string]string, len(values))
		for criterion, status := range values {
			if err := domain.ValidateMark(skill, criterion, status); err != nil {
				return err
			}
			if status != "unmarked" {
				answers[criterion] = status
			}
		}
		if _, err := draft(q, id); err != nil {
			return err
		}
		body, err := json.Marshal(answers)
		if err != nil {
			return err
		}
		_, err = q.Exec("UPDATE drafts SET answers=?,updated_at=? WHERE attempt_id=?", string(body), Now(), id)
		return err
	})
}

func (s *Store) EditAttempt(id string) error {
	return s.transaction(func(q queryer) error {
		if _, e := one(q, "SELECT id FROM attempts WHERE id=?", id); e != nil {
			return e
		}
		a, e := answers(q, id)
		if e != nil {
			return e
		}
		b, _ := json.Marshal(a)
		_, e = q.Exec("INSERT OR IGNORE INTO drafts VALUES (?,?,?)", id, string(b), Now())
		return e
	})
}
func (s *Store) Mark(id, subskill, status string) error {
	if status != "successful" && status != "not_yet" && status != "unmarked" {
		return errors.New("invalid subskill result")
	}
	return s.transaction(func(q queryer) error {
		a, e := one(q, "SELECT * FROM attempts WHERE id=?", id)
		if e != nil {
			return e
		}
		k, e := s.skill(q, a.String("curriculum_version"), a.String("skill_id"))
		if e != nil {
			return e
		}
		if e = domain.ValidateMark(k, subskill, status); e != nil {
			return e
		}
		values, e := draft(q, id)
		if e != nil {
			return e
		}
		if status == "unmarked" {
			delete(values, subskill)
		} else {
			values[subskill] = status
		}
		body, _ := json.Marshal(values)
		_, e = q.Exec("UPDATE drafts SET answers=?,updated_at=? WHERE attempt_id=?", string(body), Now(), id)
		return e
	})
}
func (s *Store) submitAttempt(q queryer, id string) (bool, error) {
	a, e := one(q, "SELECT * FROM attempts WHERE id=?", id)
	if e != nil {
		return false, e
	}
	values, e := draft(q, id)
	if errors.Is(e, ErrNotFound) {
		return a.Int("success") == 1, nil
	}
	if e != nil {
		return false, e
	}
	skill, e := s.skill(q, a.String("curriculum_version"), a.String("skill_id"))
	if e != nil {
		return false, e
	}
	normalized := map[string]string{}
	for subskillID, value := range values {
		if value == "" {
			continue
		}
		if e = domain.ValidateMark(skill, subskillID, value); e != nil {
			return false, e
		}
		normalized[subskillID] = value
	}
	success := true
	for _, subskill := range skill.Subskills {
		if normalized[subskill.ID] != "successful" {
			success = false
			if normalized[subskill.ID] == "" {
				normalized[subskill.ID] = "not_yet"
			}
		}
	}
	if a["submitted_at"] != nil {
		old, e := answers(q, id)
		if e != nil {
			return false, e
		}
		body, _ := json.Marshal(old)
		if _, e = q.Exec("INSERT INTO revisions(attempt_id,answers,saved_at) VALUES (?,?,?)", id, string(body), Now()); e != nil {
			return false, e
		}
	}
	if _, e = q.Exec("DELETE FROM results WHERE attempt_id=?", id); e != nil {
		return false, e
	}
	for _, subskill := range skill.Subskills {
		if _, e = q.Exec("INSERT INTO results VALUES (?,?,?)", id, subskill.ID, normalized[subskill.ID]); e != nil {
			return false, e
		}
	}
	successValue := 0
	if success {
		successValue = 1
	}
	if _, e = q.Exec("UPDATE attempts SET success=?,submitted_at=? WHERE id=?", successValue, Now(), id); e != nil {
		return false, e
	}
	if _, e = q.Exec("DELETE FROM drafts WHERE attempt_id=?", id); e != nil {
		return false, e
	}
	return success, nil
}

func (s *Store) Submit(id string) (bool, error) {
	var success bool
	err := s.transaction(func(q queryer) error {
		var e error
		success, e = s.submitAttempt(q, id)
		return e
	})
	return success, err
}

// SubmitAttempts submits every attempt in one transaction.
func (s *Store) SubmitAttempts(ids []string) error {
	if len(ids) == 0 {
		return errors.New("multi-student assessment has no attempts")
	}
	return s.transaction(func(q queryer) error {
		seen := map[string]bool{}
		for _, id := range ids {
			if seen[id] {
				return errors.New("an attempt was selected more than once")
			}
			seen[id] = true
			if _, e := one(q, "SELECT id FROM attempts WHERE id=? AND submitted_at IS NULL", id); e != nil {
				return e
			}
			if _, e := one(q, "SELECT attempt_id FROM drafts WHERE attempt_id=?", id); e != nil {
				return e
			}
			if _, e := s.submitAttempt(q, id); e != nil {
				return e
			}
		}
		return nil
	})
}

// DiscardMultiAttempts restores reused drafts and removes newly-created ones.
func (s *Store) DiscardMultiAttempts(attempts []domain.MultiAttempt) error {
	if len(attempts) == 0 {
		return errors.New("multi-student assessment has no attempts")
	}
	return s.transaction(func(q queryer) error {
		for _, attempt := range attempts {
			if _, e := one(q, "SELECT id FROM attempts WHERE id=? AND submitted_at IS NULL", attempt.AttemptID); e != nil {
				return e
			}
			if attempt.Created {
				for _, table := range []string{"revisions", "results", "drafts", "timed_attempts"} {
					if _, e := q.Exec("DELETE FROM "+table+" WHERE attempt_id=?", attempt.AttemptID); e != nil {
						return e
					}
				}
				if _, e := q.Exec("DELETE FROM attempts WHERE id=?", attempt.AttemptID); e != nil {
					return e
				}
				continue
			}
			body, _ := json.Marshal(attempt.OriginalAnswers)
			if _, e := q.Exec("UPDATE drafts SET answers=?,updated_at=? WHERE attempt_id=?", string(body), Now(), attempt.AttemptID); e != nil {
				return e
			}
		}
		return nil
	})
}

// DiscardAttempt removes an unfinished skill observation and its draft.
func (s *Store) DiscardAttempt(id string) error {
	return s.DiscardAttempts([]string{id})
}

// DiscardAttempts removes pending attempts atomically, never submitted results.
func (s *Store) DiscardAttempts(ids []string) error {
	return s.transaction(func(q queryer) error {
		for _, id := range ids {
			if _, e := one(q, "SELECT id FROM attempts WHERE id=? AND submitted_at IS NULL", id); e != nil {
				return e
			}
			for _, table := range []string{"revisions", "results", "drafts", "timed_attempts", "attempts"} {
				column := "attempt_id"
				if table == "attempts" {
					column = "id"
				}
				if _, e := q.Exec("DELETE FROM "+table+" WHERE "+column+"=?", id); e != nil {
					return e
				}
			}
		}
		return nil
	})
}

// RemoveLatestSuccess removes the newest successful observation for one skill.
// Its result rows and revision history are removed in the same transaction so
// the progress count and stored attempt history stay consistent.
func (s *Store) RemoveLatestSuccess(person, skill, version string) error {
	return s.transaction(func(q queryer) error {
		a, e := one(q, `SELECT id FROM attempts WHERE enrollment_id=? AND skill_id=? AND curriculum_version=? AND submitted_at IS NOT NULL AND success=1 ORDER BY submitted_at DESC,created_at DESC,rowid DESC LIMIT 1`, person, skill, version)
		if e != nil {
			return e
		}
		id := a.String("id")
		for _, table := range []string{"revisions", "results", "drafts", "timed_attempts"} {
			if _, e = q.Exec("DELETE FROM "+table+" WHERE attempt_id=?", id); e != nil {
				return e
			}
		}
		_, e = q.Exec("DELETE FROM attempts WHERE id=?", id)
		return e
	})
}
