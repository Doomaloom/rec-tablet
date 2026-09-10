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
func (s *attemptStore) Submit(string) (bool, error)                      { return true, s.err }
func (s *attemptStore) RemoveLatestSuccess(string, string, string) error { return s.err }

func TestAssessmentActionsUseCurrentDraftAndKeepFailedSubmission(t *testing.T) {
	nav := navigation.New()
	s := &attemptStore{}
	c := New(nav, s)
	nav.Push(navigation.Attempts, nil)
	nav.Push(navigation.Assessment, func(f *navigation.Frame) { f.Attempt = "draft"; f.Review = true })
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
