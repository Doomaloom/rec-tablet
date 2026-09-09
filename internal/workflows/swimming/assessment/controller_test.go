package assessment

import (
	"errors"
	"rectablet/internal/workflows/swimming/navigation"
	"testing"
)

type attemptStore struct {
	Repository                 // Unused reads deliberately panic if an action starts depending on them.
	err                        error
	attempt, criterion, status string
}

func (s *attemptStore) Mark(a, k, status string) error {
	s.attempt, s.criterion, s.status = a, k, status
	return s.err
}
func (s *attemptStore) Submit(string) (bool, error)                       { return true, s.err }
func (s *attemptStore) NewAttempt(string, string, string) (string, error) { return "draft", s.err }

func TestAssessmentActionsUseCurrentDraftAndKeepFailedSubmission(t *testing.T) {
	nav := navigation.New()
	s := &attemptStore{}
	c := New(nav, s)
	nav.Push(navigation.Attempts, nil)
	nav.State.HeaderAction = "New attempt"
	if err := c.Header(); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Attempt != "draft" {
		t.Fatal("did not open returned draft")
	}
	nav.State.Rows = []navigation.Row{{ID: "criterion"}}
	if err := c.Mark("successful"); err != nil {
		t.Fatal(err)
	}
	if s.attempt != "draft" || s.criterion != "criterion" || s.status != "successful" {
		t.Fatal("wrong draft mark")
	}
	s.err = errors.New("disk full")
	if c.Submit() == nil || nav.Frame().Screen != navigation.Assessment {
		t.Fatal("failed submission discarded screen")
	}
	s.err = nil
	if err := c.Submit(); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.Attempts || nav.State.Notice != "Attempt saved · Successful" {
		t.Fatal("successful submission did not return to attempts")
	}
}
