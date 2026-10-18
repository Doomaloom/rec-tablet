package roster

import (
	"testing"

	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming/domain"
)

func TestRosterPresentationAcrossDisplaySizes(t *testing.T) {
	swimmers := []domain.Enrollment{
		{ID: "present", Name: "A", Active: true},
		{ID: "absent", Name: "B", Active: true},
		{ID: "unmarked", Name: "C", Active: true},
		{ID: "former", Name: "D", Active: false},
	}
	marks := map[string]string{"present": "present", "absent": "absent"}
	for _, mode := range []ui.DisplayMode{"small", "large"} {
		t.Run(string(mode), func(t *testing.T) {
			state := shell.State{DisplayMode: mode, PageSize: 3}
			rows := New(nil).Configure(&state, swimmers, marks)
			wantPageSize := 3
			if mode.IsLarge() {
				wantPageSize = 12
			}
			if state.PageSize != wantPageSize || len(rows) != 3 {
				t.Fatalf("page size %d, rows %+v", state.PageSize, rows)
			}
			if rows[0].Label != "A · Present" || rows[0].Fill != ui.Green || rows[0].Status != "present" {
				t.Fatalf("present swimmer = %+v", rows[0])
			}
			if rows[1].Label != "B · Absent" || rows[1].Fill != ui.Amber || rows[1].Status != "absent" {
				t.Fatalf("absent swimmer = %+v", rows[1])
			}
			if rows[2].Label != "C · Unmarked" || rows[2].Status != "unmarked" || rows[2].Fill != nil {
				t.Fatalf("unmarked swimmer = %+v", rows[2])
			}
		})
	}
}
