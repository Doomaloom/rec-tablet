package repository

import "testing"

func TestDiscardAttemptsIsAtomicAndProtectsSubmittedResults(t *testing.T) {
	f := setup(t)
	people := must(f.s.ListRoster(f.course.String("id")))
	opened := must(f.s.OpenMultiAttempts([]string{people[0].ID, people[1].ID}, f.course.String("lesson_id"), "Splash1:1"))
	first, second := opened[0].AttemptID, opened[1].AttemptID
	check(t, f.s.Mark(first, "Splash1:1:1", "successful"))
	if err := f.s.DiscardAttempts([]string{first, "missing"}); err == nil {
		t.Fatal("batch discard accepted a missing attempt")
	}
	if draft := must(f.s.Draft(first)); draft["Splash1:1:1"] != "successful" {
		t.Fatal("failed batch discard did not roll back")
	}
	check(t, f.s.SubmitAttempts([]string{second}))
	if err := f.s.DiscardAttempts([]string{first, second}); err == nil {
		t.Fatal("batch discard accepted a submitted attempt")
	}
	if draft := must(f.s.Draft(first)); draft["Splash1:1:1"] != "successful" {
		t.Fatal("submitted attempt guard did not roll back earlier deletion")
	}
	check(t, f.s.DiscardAttempts([]string{first}))
	if _, err := f.s.GetAttempt(first); err == nil {
		t.Fatal("pending attempt was not removed")
	}
	if attempt := must(f.s.GetAttempt(second)); attempt.SubmittedAt == nil {
		t.Fatal("discard changed submitted results")
	}
}
