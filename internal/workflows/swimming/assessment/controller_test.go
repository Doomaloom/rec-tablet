package assessment

import (
	"errors"
	"math"
	"rectablet/internal/shared/ui"
	"rectablet/internal/workflows/swimming/navigation"
	"strings"
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
func (s *attemptStore) DiscardAttempt(string) error                      { return s.err }
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
	if c.Submit() != nil || nav.Frame().Screen != navigation.Assessment || nav.State.Error {
		t.Fatal("failed submission was not handled quietly")
	}
	s.err = nil
	if err := c.Submit(); err != nil {
		t.Fatal(err)
	}
	if nav.Frame().Screen != navigation.Attempts || nav.State.Notice != "" {
		t.Fatal("successful submission did not return to attempts")
	}
}

func TestMultiHeaderTextPreservesFullCriterion(t *testing.T) {
	criterion := "Front glide – face in water; underwater exhalation; arms extended beyond head"
	text := multiHeaderText(1, "  "+criterion+"\n")
	if strings.TrimPrefix(text, "2 ") != criterion {
		t.Fatalf("multi header cut off the full criterion before wrapping: %q", text)
	}
}

func TestMultiHeaderPlacementsUseEachHeaderHeight(t *testing.T) {
	short := ui.NewAngledLabel("Short", multiHeaderLength, multiHeaderFont, 0, ui.White, ui.Navy, ui.Navy)
	long := ui.NewAngledLabel("A criterion that definitely wraps onto a second line", multiHeaderLength, multiHeaderFont, 0, ui.White, ui.Navy, ui.Navy)
	headerGap := float32(4)
	placements, totalHeight := multiHeaderPlacements([]*ui.AngledLabel{short, long}, headerGap)
	if len(placements) != 2 {
		t.Fatalf("got %d header placements, want 2", len(placements))
	}
	if placements[0].Height >= placements[1].Height {
		t.Fatalf("test headers did not produce different heights: short=%v long=%v", placements[0].Height, placements[1].Height)
	}
	wantSecondOffset := placements[0].Height + headerGap
	if math.Abs(float64(placements[1].Offset-wantSecondOffset)) > 0.001 {
		t.Fatalf("second header offset = %v, want %v", placements[1].Offset, wantSecondOffset)
	}
	wantTotal := placements[0].Height + headerGap + placements[1].Height
	if math.Abs(float64(totalHeight-wantTotal)) > 0.001 {
		t.Fatalf("header stack height = %v, want %v", totalHeight, wantTotal)
	}
}
