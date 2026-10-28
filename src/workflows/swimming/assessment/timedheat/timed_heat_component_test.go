package timedheat

import (
	"testing"
	"time"

	"rectablet/src/workflows/swimming/assessment/internal/common"
	"rectablet/src/workflows/swimming/assessment/view/small"
	"rectablet/src/workflows/swimming/navigation"
)

func TestTimedHeatOwnsTickerCleanup(t *testing.T) {
	stop := make(chan struct{})
	component := Component{
		State:    State{Heat: &Heat{SkillID: "Splash7:10", Stop: stop}},
		paginate: func(state *navigation.State, rows []navigation.Row) { state.Rows = rows; state.RowCount = len(rows) },
		small:    &small.View{},
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
