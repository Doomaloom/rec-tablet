// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"errors"
)

func (s *Store) SetAttendance(person, lesson, status string) error {
	if status != "present" && status != "absent" && status != "unmarked" {
		return errors.New("choose Present, Absent, or Unmarked")
	}
	return s.transaction(func(q queryer) error {
		if _, e := checkLesson(q, person, lesson); e != nil {
			return e
		}
		_, e := q.Exec("INSERT OR REPLACE INTO attendance VALUES (?,?,?,?)", lesson, person, status, Now())
		return e
	})
}
func (s *Store) Attendance(lesson string) (map[string]string, error) {
	rs, e := s.queryRows("SELECT * FROM attendance WHERE lesson_id=?", lesson)
	out := map[string]string{}
	for _, r := range rs {
		out[r.String("enrollment_id")] = r.String("status")
	}
	return out, e
}
