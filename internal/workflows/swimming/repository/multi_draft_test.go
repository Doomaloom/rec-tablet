package repository

import (
	"reflect"
	"testing"
)

func TestReplaceDraftAnswersIsValidatedAtomicAndPendingOnly(t *testing.T) {
	f := setup(t)
	id := must(f.s.NewAttempt(f.person.String("id"), f.course.String("lesson_id"), "Splash1:1"))
	original := map[string]string{"Splash1:1:1": "successful"}
	check(t, f.s.ReplaceDraftAnswers(id, original))
	for _, invalid := range []map[string]string{
		{"Splash1:1:1": "not_yet", "unknown": "successful"},
		{"Splash1:1:1": "invalid"},
	} {
		if err := f.s.ReplaceDraftAnswers(id, invalid); err == nil {
			t.Fatal("invalid draft accepted")
		}
		if !reflect.DeepEqual(must(f.s.Draft(id)), original) {
			t.Fatal("invalid replacement changed draft")
		}
	}
	replacement := map[string]string{"Splash1:1:2": "not_yet"}
	_, err := f.s.DB.Exec("CREATE TRIGGER draft_write_failure BEFORE UPDATE ON drafts BEGIN SELECT RAISE(ABORT, 'disk full'); END")
	check(t, err)
	if f.s.ReplaceDraftAnswers(id, replacement) == nil {
		t.Fatal("expected write failure")
	}
	if !reflect.DeepEqual(must(f.s.Draft(id)), original) {
		t.Fatal("failed write changed draft")
	}
	_, err = f.s.DB.Exec("DROP TRIGGER draft_write_failure")
	check(t, err)
	check(t, f.s.ReplaceDraftAnswers(id, replacement))
	if !reflect.DeepEqual(must(f.s.Draft(id)), replacement) {
		t.Fatal("replacement merged old marks")
	}
	check(t, f.s.ReplaceDraftAnswers(id, map[string]string{"Splash1:1:1": "unmarked"}))
	if len(must(f.s.Draft(id))) != 0 {
		t.Fatal("unmarked criterion was not cleared")
	}
	check(t, f.s.ReplaceDraftAnswers(id, original))
	_ = must(f.s.Submit(id))
	check(t, f.s.EditAttempt(id))
	before := must(f.s.Draft(id))
	if f.s.ReplaceDraftAnswers(id, replacement) == nil {
		t.Fatal("submitted attempt allowed through pending API")
	}
	if !reflect.DeepEqual(must(f.s.Draft(id)), before) {
		t.Fatal("submitted correction draft changed")
	}
	if f.s.ReplaceDraftAnswers("missing", original) == nil {
		t.Fatal("missing attempt accepted")
	}
}
