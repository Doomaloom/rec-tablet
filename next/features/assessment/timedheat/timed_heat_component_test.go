package timedheat

import (
	"testing"
	"time"

	"rectablet/next/features/assessment/internal/common"
	"rectablet/next/navigation"
	"rectablet/src/workflows/swimming/domain"
)

func TestTimedHeatOwnsTickerCleanup(t *testing.T) {
	stop := make(chan struct{})
	component := Component{
		State:    State{Heat: &Heat{SkillID: "Splash7:10", Stop: stop}},
		paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows; state.RowCount = len(rows) },
		small:    &CompactTemplate{},
	}
	state, err := component.Build(common.Input{Frame: navigation.Frame{Screen: navigation.TimedHeat}})
	if err != nil {
		t.Fatal(err)
	}
	if state.OnLeave == nil {
		t.Fatal("heat has no cleanup callback")
	}
	state.OnLeave()
	select {
	case <-stop:
	default:
		t.Fatal("leaving heat did not stop its ticker")
	}
	state.OnLeave() // Leaving again is harmless.
}

func TestClockTickDoesNotRebuildScreen(t *testing.T) {
	stop := make(chan struct{})
	heat := &Heat{Running: true, StartedAt: time.Now().Add(-time.Second), Stop: stop}
	updates := 0
	c := Component{State: State{Heat: heat}, navigation: common.Navigation{
		Refresh: func() error { t.Fatal("clock tick rebuilt the screen during a press"); return nil },
	}, updateTitle: func(title string) {
		if title == "" {
			t.Fatal("empty clock")
		}
		updates++
	}}
	c.tick(heat, stop)
	if updates != 1 {
		t.Fatal("clock did not update")
	}
	heat.Running = false
	c.tick(heat, stop)
	heat.Running = true
	heat.Stop = make(chan struct{})
	c.tick(heat, stop) // A queued tick from the previous run must be ignored.
	if updates != 1 {
		t.Fatal("stale or paused clock updated")
	}
	c.StopTicker()
	c.tick(heat, heat.Stop)
	if updates != 1 {
		t.Fatal("clock updated after leaving")
	}
}

func TestTimedOutcomesUseLapsAndCompletionTime(t *testing.T) {
	for _, tc := range []struct {
		name      string
		laps      int
		completed bool
		at        time.Duration
		want      bool
	}{
		{"before limit", 3, true, 179 * time.Second, true},
		{"at limit", 3, true, 180 * time.Second, true},
		{"after limit", 3, true, 180*time.Second + time.Millisecond, false},
		{"incomplete laps", 2, false, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := Component{
				State: State{Heat: &Heat{SkillID: "Splash7:10", Elapsed: 4 * time.Minute, Participants: []ParticipantState{{
					Participant: Participant{AttemptID: "attempt", LapsRequired: 3, LapsCompleted: tc.laps}, Completed: tc.completed, CompletedAt: tc.at,
				}}}},
				navigation: common.Navigation{
					Frame: func() navigation.Frame { return navigation.Frame{Screen: navigation.TimedHeat} },
					Run:   func(fn func() error) error { return fn() }, Pop: func() {},
				},
				paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows },
				small:    &CompactTemplate{},
			}
			state, err := c.Build(common.Input{Frame: navigation.Frame{Screen: navigation.TimedHeat}})
			if err != nil {
				t.Fatal(err)
			}
			if got := state.Rows[0].Status == "successful"; got != tc.want {
				t.Fatalf("visible success = %v, want %v", got, tc.want)
			}
			called := false
			c.storage.SaveTimedHeat = func(results []domain.TimedResult) error {
				called = true
				if len(results) != 1 || results[0].Success != tc.want {
					t.Fatalf("results = %+v", results)
				}
				if tc.completed && results[0].ElapsedMS != tc.at.Milliseconds() {
					t.Fatal("save used heat time instead of finish time")
				}
				return nil
			}
			if err := c.SaveTimedHeat(); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("heat was not saved")
			}
		})
	}
}
