package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"rectablet/src/workflows/swimming/transfer"
)

// LessonPlan returns nil for lessons with no imported plan.
func (s *Store) LessonPlan(lesson string) (*transfer.LessonPlan, error) {
	var body string
	err := s.DB.QueryRow("SELECT body FROM lesson_plans WHERE lesson_id=?", lesson).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var plan transfer.LessonPlan
	if err = json.Unmarshal([]byte(body), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
