package integration_test

import (
	"testing"
	"time"

	"rectablet/src/shared/ui"
)

func TestAttendanceTapsReuseMountedControls(t *testing.T) {
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		t.Run(string(mode), func(t *testing.T) {
			f := setupUI(t)
			f.c.ConfigureDisplay(mode, nil)
			f.row(t, 0)
			f.tap(t, "row:attendance")
			row := f.c.Snapshot().Rows[0]
			button := f.view.Buttons["row:"+row.ID]
			background := f.view.Root.Objects[0]
			f.tap(t, "row:"+row.ID)
			if f.view.Buttons["row:"+row.ID] != button || f.view.Root.Objects[0] != background {
				t.Fatal("attendance tap rebuilt screen")
			}
			if f.c.Snapshot().Rows[0].Status != "present" || button.Fill != ui.Green {
				t.Fatal("attendance did not repaint immediately")
			}
			f.tap(t, "row:"+row.ID)
			marks := must(f.s.Attendance(f.c.Frame().Lesson))
			if marks[row.ID] != "absent" || button.Fill != ui.Amber {
				t.Fatal("repeated tap did not update persisted and visible status")
			}
		})
	}
}

func TestAsyncChecklistRemainsResponsiveWhileDatabaseIsBusy(t *testing.T) {
	f := setupUI(t)
	f.skill(t)
	frame := f.c.Frame()
	row := f.c.Snapshot().Rows[0]
	button := f.view.Buttons["row:"+row.ID]
	callbacks := make(chan func(), 10)
	f.c.EnableAsyncSaves(func(fn func()) { callbacks <- fn })
	// Occupy the sole database connection: the mark cannot complete until release.
	tx := must(f.s.DB.Begin())
	defer tx.Rollback()
	f.tap(t, "row:"+row.ID)
	if f.c.Snapshot().Rows[0].Status != "successful" || button.Fill != ui.Green {
		t.Fatal("slow storage blocked immediate tap feedback")
	}
	f.tap(t, "row:"+row.ID)
	if f.c.Snapshot().Rows[0].Status != "unmarked" || f.view.Buttons["row:"+row.ID] != button {
		t.Fatal("second tap rebuilt screen or lost optimistic state")
	}
	check(t, f.c.Home())
	if f.c.Frame().Screen != frame.Screen {
		t.Fatal("navigation raced pending saves")
	}
	check(t, tx.Rollback())
	for i := 0; i < 2; i++ {
		select {
		case fn := <-callbacks:
			fn()
		case <-time.After(5 * time.Second):
			t.Fatal("save worker did not complete")
		}
	}
	if f.c.Frame().Screen == frame.Screen {
		t.Fatal("navigation did not resume after saves")
	}
	draft := must(f.s.Draft(frame.Attempt))
	if draft[row.ID] != "" {
		t.Fatalf("queued toggles saved in wrong order: %v", draft)
	}
}
