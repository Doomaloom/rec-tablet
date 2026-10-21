package datepicker

import (
	"testing"

	"rectablet/next/contracts"
)

func TestCalendarBrowsingAndReopening(t *testing.T) {
	picker := New(Options{})
	dates := []string{"2026-12-04", "2026-12-11", "2026-12-18", "2026-12-25"}
	picker.Layout = Layout{Calendar: true}
	state := contracts.State{DisplayMode: "large", PageSize: 3}
	rows := picker.Configure(&state, dates, "2026-12-11")
	if state.PageSize != len(dates) || len(rows) != len(dates) || !rows[1].Selected {
		t.Fatalf("calendar lost scheduled dates or selection: state %+v, rows %+v", state, rows)
	}
	picker.MoveMonth(1, "2026-12-11")
	if got := picker.Month().Format("2006-01-02"); got != "2027-01-01" {
		t.Fatalf("month crossing year boundary = %s", got)
	}
	picker.Configure(&state, dates, "2026-12-11")
	if got := picker.Month().Format("2006-01"); got != "2027-01" {
		t.Fatalf("refresh reset browsed month to %s", got)
	}
	picker.Reset()
	picker.Configure(&state, dates, "2026-12-18")
	if got := picker.Month().Format("2006-01"); got != "2026-12" {
		t.Fatalf("reopened picker month = %s", got)
	}
}

func TestSmallDateListKeepsPagination(t *testing.T) {
	picker := New(Options{})
	state := contracts.State{DisplayMode: "small", PageSize: 3}
	rows := picker.Configure(&state, []string{"2026-09-04"}, "2026-09-04")
	if state.PageSize != 3 || !picker.Month().IsZero() || len(rows) != 1 || rows[0].Label != "Friday, Sep 04, 2026" || !rows[0].Selected {
		t.Fatalf("small date list = %+v, rows %+v, month %s", state, rows, picker.Month())
	}
}
