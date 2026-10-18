package timedheat

import (
	"testing"

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
