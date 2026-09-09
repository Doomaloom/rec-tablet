package domain

import (
	"errors"
	"rectablet/internal/workflows/swimming/curriculum"
)

const RequiredSuccesses = 3

func CalculateProgress(assessed, successes int) Progress {
	return Progress{assessed, successes, successes >= RequiredSuccesses}
}
func Assess(skill curriculum.Skill, answers map[string]string) (bool, error) {
	if len(answers) != len(skill.Subskills) {
		return false, errors.New("assess every subskill before submitting this attempt")
	}
	success := true
	for _, criterion := range skill.Subskills {
		value := answers[criterion.ID]
		if value != "successful" && value != "not_yet" {
			return false, errors.New("assess every subskill before submitting this attempt")
		}
		if value != "successful" {
			success = false
		}
	}
	return success, nil
}
func Remaining(skill curriculum.Skill, answers map[string]string) []curriculum.Subskill {
	remaining := []curriculum.Subskill{}
	for _, criterion := range skill.Subskills {
		if answers[criterion.ID] == "not_yet" {
			remaining = append(remaining, criterion)
		}
	}
	return remaining
}
func ValidateMark(skill curriculum.Skill, id, status string) error {
	if status != "successful" && status != "not_yet" && status != "unmarked" {
		return errors.New("invalid subskill result")
	}
	for _, criterion := range skill.Subskills {
		if criterion.ID == id {
			return nil
		}
	}
	return errors.New("unknown subskill for this attempt")
}
func ValidateNewAttempt(catalog curriculum.Catalog, levelID, skillID string, active bool) error {
	level, ok := catalog.Level(levelID)
	valid := false
	for _, skill := range level.Skills {
		if skill.ID == skillID {
			valid = true
		}
	}
	if !ok || !valid {
		return errors.New("this skill is not part of the swimmer's assigned level")
	}
	if !active {
		return errors.New("this swimmer is no longer enrolled; existing results remain available")
	}
	return nil
}
