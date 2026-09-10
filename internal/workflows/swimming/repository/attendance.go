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

// MarkPastAttendanceAbsent closes a lesson's unchecked active roster entries.
// Existing present or absent marks are preserved. The date comparison is
// supplied by the caller so the device's local date is used consistently.
func (s *Store) MarkPastAttendanceAbsent(lesson, today string) error {
	return s.transaction(func(q queryer) error {
		l, e := one(q, "SELECT course_id,date FROM lessons WHERE id=?", lesson)
		if e != nil {
			return e
		}
		if l.String("date") >= today {
			return nil
		}
		updatedAt := Now()
		if _, e = q.Exec(`
			UPDATE attendance SET status='absent',updated_at=?
			WHERE lesson_id=? AND status='unmarked'
			  AND enrollment_id IN (SELECT id FROM enrollments WHERE active=1)`, updatedAt, lesson); e != nil {
			return e
		}
		_, e = q.Exec(`
			INSERT INTO attendance(lesson_id,enrollment_id,status,updated_at)
			SELECT ?,e.id,'absent',?
			FROM enrollments e
			WHERE e.course_id=? AND e.active=1
			  AND NOT EXISTS (
				SELECT 1 FROM attendance a
				WHERE a.lesson_id=? AND a.enrollment_id=e.id
			  )`, lesson, updatedAt, l.String("course_id"), lesson)
		return e
	})
}
