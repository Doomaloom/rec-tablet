// Package repository persists swimming data and applies domain rules within transactions.
package repository

import (
	"errors"
)

func (s *Store) courseRecords(date string) ([]record, error) {
	return s.queryRows(`SELECT c.*,l.id AS lesson_id FROM courses c JOIN lessons l ON l.course_id=c.id WHERE l.date=? ORDER BY c.time,c.name`, date)
}
func (s *Store) Dates() ([]string, error) {
	rs, e := s.queryRows("SELECT DISTINCT date FROM lessons ORDER BY date")
	out := []string{}
	for _, r := range rs {
		out = append(out, r.String("date"))
	}
	return out, e
}
func (s *Store) rosterRecords(course string) ([]record, error) {
	return s.queryRows(`SELECT e.*,s.name FROM enrollments e JOIN swimmers s ON s.id=e.swimmer_id WHERE e.course_id=? ORDER BY e.active DESC,s.name COLLATE NOCASE`, course)
}
func enrollment(q queryer, id string) (record, error) {
	return one(q, `SELECT e.*,s.name,c.curriculum_version FROM enrollments e JOIN swimmers s ON s.id=e.swimmer_id JOIN courses c ON c.id=e.course_id WHERE e.id=?`, id)
}
func (s *Store) enrollmentRecord(id string) (record, error) { return enrollment(s.DB, id) }
func checkLesson(q queryer, person, lesson string) (record, error) {
	p, e := enrollment(q, person)
	if e != nil {
		return nil, e
	}
	l, e := one(q, "SELECT * FROM lessons WHERE id=?", lesson)
	if e != nil {
		return nil, e
	}
	if p.String("course_id") != l.String("course_id") {
		return nil, errors.New("the swimmer and lesson belong to different courses")
	}
	return p, nil
}
