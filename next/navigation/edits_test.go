package navigation

import (
	"errors"
	"testing"
	"time"

	"rectablet/next/contracts"
)

type editFeature struct {
	nav       *Controller
	builds    int
	persisted string
	started   chan string
	allow     chan struct{}
	fail      bool
}

func (f *editFeature) Build() (State, error) {
	f.builds++
	state := f.nav.InitialState()
	state.Rows = []Row{{ID: "criterion", Label: "Criterion", Status: f.persisted}}
	return state, nil
}
func (f *editFeature) Select(string) error { return nil }
func (f *editFeature) Header() error       { return nil }
func (f *editFeature) Mark(string) error   { return nil }
func (f *editFeature) Submit() error       { return nil }
func (f *editFeature) PrepareEdit(id string) (Edit, bool) {
	if id != "criterion" {
		return Edit{}, false
	}
	row := f.nav.State.Rows[0]
	row.Status = "successful"
	if f.nav.State.Rows[0].Status == "successful" {
		row.Status = "unmarked"
	}
	status, fail := row.Status, f.fail
	return Edit{Row: row, Button: contracts.ButtonUpdate{ID: id, Text: row.Label}, Save: func() error {
		f.started <- status
		<-f.allow
		if fail {
			return errors.New("disk full")
		}
		f.persisted = status
		return nil
	}}, true
}
func editFixture(t *testing.T) (*Controller, *editFeature, chan func()) {
	t.Helper()
	nav := New()
	feature := &editFeature{nav: nav, persisted: "unmarked", started: make(chan string, 10), allow: make(chan struct{}, 10)}
	nav.Features[Assessment] = feature
	nav.Push(Assessment, nil)
	if err := nav.Refresh(); err != nil {
		t.Fatal(err)
	}
	callbacks := make(chan func(), 10)
	nav.EnableAsyncSaves(func(fn func()) { callbacks <- fn })
	nav.SubscribeEdits(func(contracts.ButtonUpdate) {})
	t.Cleanup(func() { close(feature.allow); nav.DrainSaves() })
	return nav, feature, callbacks
}
func callback(t *testing.T, callbacks chan func()) {
	t.Helper()
	select {
	case fn := <-callbacks:
		fn()
	case <-time.After(time.Second):
		t.Fatal("missing save completion")
	}
}
func started(t *testing.T, feature *editFeature) string {
	t.Helper()
	select {
	case status := <-feature.started:
		return status
	case <-time.After(time.Second):
		t.Fatal("save worker did not start")
		return ""
	}
}
func TestRapidEditsPaintImmediatelyAndSaveInOrderBeforeNavigation(t *testing.T) {
	nav, f, callbacks := editFixture(t)
	if err := nav.Select("criterion"); err != nil {
		t.Fatal(err)
	}
	if started(t, f) != "successful" {
		t.Fatal("wrong first saved value")
	}
	if err := nav.Select("criterion"); err != nil {
		t.Fatal(err)
	}
	if nav.State.Rows[0].Status != "unmarked" || f.builds != 1 {
		t.Fatal("tap must update state without waiting or rebuilding")
	}
	moved := false
	_ = nav.Run(func() error { moved = true; return nil })
	if moved {
		t.Fatal("navigation overtook queued marks")
	}
	f.allow <- struct{}{}
	callback(t, callbacks)
	if moved {
		t.Fatal("navigation ran before all writes")
	}
	if started(t, f) != "unmarked" {
		t.Fatal("rapid toggles saved out of order")
	}
	f.allow <- struct{}{}
	callback(t, callbacks)
	if !moved || f.persisted != "unmarked" || f.builds != 2 {
		t.Fatal("navigation did not resume after ordered saves")
	}
}
func TestFailedSaveRestoresPersistedStateAndCancelsNavigation(t *testing.T) {
	nav, f, callbacks := editFixture(t)
	f.fail = true
	_ = nav.Select("criterion")
	started(t, f)
	_ = nav.Back()
	f.allow <- struct{}{}
	callback(t, callbacks)
	if nav.Frame().Screen != Assessment || !nav.State.Error || nav.State.Rows[0].Status != "unmarked" {
		t.Fatal("failed save must stay on screen and reload saved marks")
	}
	f.fail = false
	_ = nav.Select("criterion")
	if started(t, f) != "successful" {
		t.Fatal("retry toggled from failed optimistic state")
	}
	f.allow <- struct{}{}
	callback(t, callbacks)
	if f.persisted != "successful" {
		t.Fatal("failed save could not be retried")
	}
}
func TestCloseDrainsWritesWithoutWaitingForUICallbacks(t *testing.T) {
	nav, f, _ := editFixture(t)
	_ = nav.Select("criterion")
	started(t, f)
	nav.Closed = true
	drained := make(chan struct{})
	go func() { nav.DrainSaves(); close(drained) }()
	select {
	case <-drained:
		t.Fatal("closed before pending write finished")
	default:
	}
	f.allow <- struct{}{}
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("drain waited for UI callback")
	}
	if f.persisted != "successful" {
		t.Fatal("close lost pending mark")
	}
}
