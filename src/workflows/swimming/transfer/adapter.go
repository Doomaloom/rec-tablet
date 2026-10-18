// Package transfer adapts annotated sister-app rosters without inventing identity.
package transfer

import (
	"encoding/json"
	"errors"
)

func Adapt(data []byte, dataset, instructor string) (ClassPackage, error) {
	var rosters []struct {
		SourceID    string   `json:"sourceId"`
		ServiceName string   `json:"serviceName"`
		Level       string   `json:"level"`
		StartTime   string   `json:"startTime24"`
		Location    string   `json:"location"`
		LessonDates []string `json:"lessonDates"`
		Students    []struct {
			SourceID string `json:"sourceId"`
			Name     string `json:"name"`
			Level    string `json:"level"`
			Waitlist bool   `json:"waitlist"`
		} `json:"students"`
	}
	p := ClassPackage{SchemaVersion: 1, Kind: "rec-tablet-classes", DatasetID: dataset, Instructor: instructor, Courses: []ClassInput{}}
	if e := json.Unmarshal(data, &rosters); e != nil {
		return p, e
	}
	if rosters == nil {
		return p, errors.New("expected a JSON array of annotated ClassRoster objects")
	}
	for _, r := range rosters {
		for key, value := range map[string]string{"sourceId": r.SourceID, "serviceName": r.ServiceName, "level": r.Level, "startTime24": r.StartTime} {
			if _, e := Text(value, key); e != nil {
				return p, e
			}
		}
		if r.Students == nil {
			return p, errors.New("each ClassRoster needs a students array")
		}
		c := ClassInput{SourceID: r.SourceID, Name: r.ServiceName, Level: r.Level, StartTime: r.StartTime, Location: r.Location, LessonDates: r.LessonDates, Swimmers: []SwimmerInput{}}
		for _, s := range r.Students {
			if s.Waitlist {
				continue
			}
			if _, e := Text(s.SourceID, "sourceId"); e != nil {
				return p, e
			}
			if _, e := Text(s.Name, "name"); e != nil {
				return p, e
			}
			level := s.Level
			if level == "" {
				level = r.Level
			}
			c.Swimmers = append(c.Swimmers, SwimmerInput{SourceID: s.SourceID, Name: s.Name, Level: &level})
		}
		p.Courses = append(p.Courses, c)
	}
	return p, nil
}
