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
