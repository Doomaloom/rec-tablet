package repository

import "rectablet/internal/workflows/swimming/domain"

// Typed queries form the boundary consumed by features. Raw records stay here
// to preserve the v1 table snapshot contract and transactional SQL helpers.
func courseFrom(r record) domain.Course {
	return domain.Course{ID: r.String("id"), Name: r.String("name"), Level: r.String("level"), Time: r.String("time"), Location: r.String("location"), LessonID: r.String("lesson_id"), CurriculumVersion: r.String("curriculum_version")}
}
func enrollmentFrom(r record) domain.Enrollment {
	return domain.Enrollment{ID: r.String("id"), Name: r.String("name"), Level: r.String("level"), CourseID: r.String("course_id"), CurriculumVersion: r.String("curriculum_version"), Active: r.Int("active") != 0}
}
func attemptFrom(r record) domain.Attempt {
	a := domain.Attempt{ID: r.String("id"), EnrollmentID: r.String("enrollment_id"), LessonID: r.String("lesson_id"), SkillID: r.String("skill_id"), CurriculumVersion: r.String("curriculum_version"), CreatedAt: r.String("created_at"), Date: r.String("date"), Success: r.Int("success") != 0, Editing: r.Int("editing") != 0}
	if r["submitted_at"] != nil {
		v := r.String("submitted_at")
		a.SubmittedAt = &v
	}
	return a
}
func (s *Store) GetCourse(id string) (domain.Course, error) {
	r, e := s.queryOne("SELECT * FROM courses WHERE id=?", id)
	return courseFrom(r), e
}
func (s *Store) ListCourses(date string) ([]domain.Course, error) {
	rs, e := s.courseRecords(date)
	out := []domain.Course{}
	for _, r := range rs {
		out = append(out, courseFrom(r))
	}
	return out, e
}
func (s *Store) ListRoster(id string) ([]domain.Enrollment, error) {
	rs, e := s.rosterRecords(id)
	out := []domain.Enrollment{}
	for _, r := range rs {
		out = append(out, enrollmentFrom(r))
	}
	return out, e
}
func (s *Store) GetEnrollment(id string) (domain.Enrollment, error) {
	r, e := s.enrollmentRecord(id)
	return enrollmentFrom(r), e
}
func (s *Store) ListAttempts(person, skill, version string) ([]domain.Attempt, error) {
	rs, e := s.attemptRecords(person, skill, version)
	out := []domain.Attempt{}
	for _, r := range rs {
		out = append(out, attemptFrom(r))
	}
	return out, e
}
func (s *Store) GetAttempt(id string) (domain.Attempt, error) {
	r, e := s.queryOne("SELECT * FROM attempts WHERE id=?", id)
	return attemptFrom(r), e
}
func (s *Store) DatabasePath() string { return s.Path }
