CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE curricula (version TEXT PRIMARY KEY, body TEXT NOT NULL);
CREATE TABLE courses (
 id TEXT PRIMARY KEY, dataset_id TEXT NOT NULL, source_id TEXT NOT NULL,
 name TEXT NOT NULL, level TEXT NOT NULL, time TEXT NOT NULL, location TEXT NOT NULL,
 curriculum_version TEXT NOT NULL REFERENCES curricula(version),
 UNIQUE(dataset_id, source_id));
CREATE TABLE swimmers (
 id TEXT PRIMARY KEY, dataset_id TEXT NOT NULL, source_id TEXT NOT NULL, name TEXT NOT NULL,
 UNIQUE(dataset_id, source_id));
CREATE TABLE enrollments (
 id TEXT PRIMARY KEY, course_id TEXT NOT NULL REFERENCES courses(id),
 swimmer_id TEXT NOT NULL REFERENCES swimmers(id), level TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 1,
 UNIQUE(course_id, swimmer_id));
CREATE TABLE lessons (
 id TEXT PRIMARY KEY, course_id TEXT NOT NULL REFERENCES courses(id), date TEXT NOT NULL,
 UNIQUE(course_id, date));
CREATE TABLE attendance (
 lesson_id TEXT NOT NULL REFERENCES lessons(id), enrollment_id TEXT NOT NULL REFERENCES enrollments(id),
 status TEXT NOT NULL CHECK(status IN ('present','absent','unmarked')), updated_at TEXT NOT NULL,
 PRIMARY KEY(lesson_id, enrollment_id));
CREATE TABLE attempts (
 id TEXT PRIMARY KEY, enrollment_id TEXT NOT NULL REFERENCES enrollments(id),
 lesson_id TEXT NOT NULL REFERENCES lessons(id), skill_id TEXT NOT NULL,
 curriculum_version TEXT NOT NULL REFERENCES curricula(version),
 created_at TEXT NOT NULL, submitted_at TEXT, success INTEGER NOT NULL DEFAULT 0 CHECK(success IN (0,1)));
CREATE TABLE results (
 attempt_id TEXT NOT NULL REFERENCES attempts(id), subskill_id TEXT NOT NULL,
 status TEXT NOT NULL CHECK(status IN ('successful','not_yet')),
 PRIMARY KEY(attempt_id, subskill_id));
CREATE TABLE drafts (
 attempt_id TEXT PRIMARY KEY REFERENCES attempts(id), answers TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE revisions (
 id INTEGER PRIMARY KEY, attempt_id TEXT NOT NULL REFERENCES attempts(id),
 answers TEXT NOT NULL, saved_at TEXT NOT NULL);
CREATE INDEX attempt_progress ON attempts(enrollment_id, skill_id, submitted_at);
PRAGMA user_version = 1;
