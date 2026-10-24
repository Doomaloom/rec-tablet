package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxPackageBytes = 10 * 1024 * 1024

type ClassPackage struct {
	SchemaVersion int          `json:"schemaVersion"`
	Kind          string       `json:"kind"`
	DatasetID     string       `json:"datasetId"`
	Instructor    string       `json:"instructor"`
	Courses       []ClassInput `json:"courses"`
}
type ClassInput struct {
	LessonPlans *[]LessonPlan  `json:"lessonPlans,omitempty"`
	SourceID    string         `json:"sourceId"`
	Name        string         `json:"name"`
	Level       string         `json:"level"`
	StartTime   string         `json:"startTime"`
	Location    string         `json:"location"`
	LessonDates []string       `json:"lessonDates"`
	Swimmers    []SwimmerInput `json:"swimmers"`
}
type SwimmerInput struct {
	SourceID string  `json:"sourceId"`
	Name     string  `json:"name"`
	Level    *string `json:"level,omitempty"`
}

func Text(value, key string) (string, error) {
	if strings.TrimSpace(value) == "" || utf8.RuneCountInString(value) > 1000 {
		return "", fmt.Errorf("%s must be a nonempty string (maximum 1000 characters)", key)
	}
	return strings.TrimSpace(value), nil
}
func Date(value string) error {
	d, e := time.Parse("2006-01-02", value)
	if e != nil || d.Year() < 1 || d.Format("2006-01-02") != value {
		return errors.New("lesson dates must use YYYY-MM-DD")
	}
	return nil
}
func DecodePackage(data []byte) (ClassPackage, error) {
	var p ClassPackage
	if len(data) > MaxPackageBytes {
		return p, errors.New("class package is too large (maximum 10 MB)")
	}
	if !utf8.Valid(data) {
		return p, errors.New("the selected file is not valid UTF-8 JSON")
	}
	if e := json.Unmarshal(data, &p); e != nil {
		return p, fmt.Errorf("invalid class package JSON: %w", e)
	}
	// Distinguish explicit null from an omitted optional string.
	var raw struct {
		Courses []struct {
			Location json.RawMessage `json:"location"`
			Swimmers []struct {
				Level json.RawMessage `json:"level"`
			} `json:"swimmers"`
		} `json:"courses"`
	}
	if e := json.Unmarshal(data, &raw); e != nil {
		return p, e
	}
	for _, c := range raw.Courses {
		if string(c.Location) == "null" {
			return p, errors.New("location must be text")
		}
		for _, s := range c.Swimmers {
			if string(s.Level) == "null" {
				return p, errors.New("level must be nonempty text when provided")
			}
		}
	}
	return p, nil
}
func ReadPackage(path string) (ClassPackage, error) {
	f, e := os.Open(path)
	if e != nil {
		return ClassPackage{}, e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, MaxPackageBytes+1))
	if e != nil {
		return ClassPackage{}, e
	}
	return DecodePackage(data)
}

// LessonPlan is a read-only snapshot for an exact scheduled lesson date.
type LessonPlan struct {
	Date            string       `json:"date"`
	CurriculumLevel string       `json:"curriculumLevel"`
	Rows            []PlannedRow `json:"rows"`
}
type PlannedRow struct {
	Skill      string            `json:"skill"`
	SkillID    string            `json:"skillId,omitempty"`
	Activity   string            `json:"activity"`
	Activities []PlannedActivity `json:"activities,omitempty"`
	Location   string            `json:"location"`
	Duration   int               `json:"duration"`
	Equipment  []string          `json:"equipment,omitempty"`
	Workout    json.RawMessage   `json:"workout,omitempty"`
}
type PlannedActivity struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

func ValidateLessonPlans(course ClassInput) error {
	if course.LessonPlans == nil {
		return nil
	}
	if len(*course.LessonPlans) > 366 {
		return errors.New("too many lesson plans")
	}
	dates := map[string]bool{}
	for _, d := range course.LessonDates {
		dates[d] = true
	}
	seen := map[string]bool{}
	for _, plan := range *course.LessonPlans {
		if Date(plan.Date) != nil || !dates[plan.Date] || seen[plan.Date] || len(plan.CurriculumLevel) > 100 || plan.Rows == nil || len(plan.Rows) > 200 {
			return errors.New("invalid lesson plan date or rows")
		}
		seen[plan.Date] = true
		for _, r := range plan.Rows {
			if len(r.Skill) > 2000 || len(r.SkillID) > 200 || len(r.Activity) > 10000 || len(r.Location) > 100 || r.Duration < 1 || r.Duration > 240 || len(r.Equipment) > 50 || len(r.Activities) > 100 {
				return errors.New("invalid planned skill")
			}
			equipment := map[string]bool{}
			for _, item := range r.Equipment {
				key := strings.ToLower(strings.TrimSpace(item))
				if key == "" || utf8.RuneCountInString(item) > 100 || equipment[key] {
					return errors.New("invalid planned equipment")
				}
				equipment[key] = true
			}
			texts := []string{}
			for _, a := range r.Activities {
				if (a.Kind != "library" && a.Kind != "custom") || strings.TrimSpace(a.Text) == "" || len(a.Text) > 10000 {
					return errors.New("invalid planned activity")
				}
				texts = append(texts, a.Text)
			}
			if len(texts) > 0 && (strings.Join(texts, "\n\n") != r.Activity || len(r.Workout) > 0) {
				return errors.New("inconsistent planned activities")
			}
		}
	}
	return nil
}
