package navigation

import (
	"errors"
	"reflect"
	"testing"
)

type testFeature struct {
	build     func() (State, error)
	selectRow func(string) error
}

func (f *testFeature) Build() (State, error)  { return f.build() }
func (f *testFeature) Select(id string) error { return f.selectRow(id) }
func (f *testFeature) Header() error          { return nil }
func (f *testFeature) Mark(string) error      { return nil }
func (f *testFeature) Submit() error          { return nil }

func TestFailedDispatchRestoresHistoryAndDate(t *testing.T) {
	for _, failDuring := range []string{"action", "refresh"} {
		t.Run(failDuring, func(t *testing.T) {
			nav := New()
			nav.Date = "2026-09-04"
			failure := errors.New("operation failed")
			failing := true
			nav.Features[Classes] = &testFeature{
				build: func() (State, error) {
					state := nav.InitialState()
					nav.Paginate(&state, []Row{{ID: "next", Label: "Next"}})
					return state, nil
				},
				selectRow: func(string) error {
					nav.Date = "2026-09-11"
					nav.History.Frames[0].Course = "changed"
					nav.Push(Dates, nil)
					if failing && failDuring == "action" {
						return failure
					}
					return nil
				},
			}
			nav.Features[Dates] = &testFeature{build: func() (State, error) {
				if failing && failDuring == "refresh" {
					return State{}, failure
				}
				return nav.InitialState(), nil
			}}
			if err := nav.Refresh(); err != nil {
				t.Fatal(err)
			}
			before := append([]Frame(nil), nav.History.Frames...)
			if err := nav.Select("next"); !errors.Is(err, failure) {
				t.Fatalf("dispatch error = %v", err)
			}
			if !reflect.DeepEqual(nav.History.Frames, before) || nav.Date != "2026-09-04" {
				t.Fatalf("failed dispatch changed history/date: %+v, %s", nav.History.Frames, nav.Date)
			}
			if state := nav.Snapshot(); !state.Error || state.Notice != failure.Error() || state.Screen != string(Classes) || len(state.Rows) != 1 {
				t.Fatalf("failed dispatch lost the previous screen or error: %+v", state)
			}
			failing = false
			if err := nav.Select("next"); err != nil {
				t.Fatal(err)
			}
			if nav.Frame().Screen != Dates || len(nav.History.Frames) != 2 {
				t.Fatalf("retry history = %+v", nav.History.Frames)
			}
			if err := nav.Back(); err != nil {
				t.Fatal(err)
			}
			if nav.Frame().Screen != Classes || len(nav.History.Frames) != 1 {
				t.Fatalf("back history = %+v", nav.History.Frames)
			}
			if err := nav.Home(); err != nil {
				t.Fatal(err)
			}
			if nav.Frame() != before[0] || nav.Date != "2026-09-11" {
				t.Fatalf("home changed the selected date or failed to reset history: %+v, %s", nav.Frame(), nav.Date)
			}
		})
	}
}
