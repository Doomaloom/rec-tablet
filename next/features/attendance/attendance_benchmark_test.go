package attendance

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2/test"
	host "rectablet/next/app"
	"rectablet/next/navigation"
	"rectablet/src/shared/ui"
	"rectablet/src/workflows/swimming/domain"
)

// Compare the two UI paths with the same twelve-student roster and memory store,
// so storage latency does not obscure allocation and rendering costs.
func BenchmarkAttendanceTap(b *testing.B) {
	for _, incremental := range []bool{false, true} {
		name := "full-screen"
		if incremental {
			name = "single-button"
		}
		b.Run(name, func(b *testing.B) {
			app := test.NewApp()
			defer app.Quit()
			window := app.NewWindow("Attendance benchmark")
			defer window.Close()
			store := &attendanceStore{marks: map[string]string{}}
			for i := 0; i < 12; i++ {
				store.roster = append(store.roster, domain.Enrollment{ID: fmt.Sprint(i), Name: fmt.Sprintf("Swimmer %d", i), Active: true})
			}
			nav := navigation.New()
			nav.DisplayMode = ui.DisplayLarge
			feature := New(nav, store)
			nav.Features[navigation.Attendance] = feature
			nav.Push(navigation.Attendance, func(f *navigation.Frame) { f.Course = "course"; f.Lesson = "lesson" })
			if err := nav.Refresh(); err != nil {
				b.Fatal(err)
			}
			view := host.New(nav, window)
			defer view.Close()
			window.Show()
			if !incremental {
				nav.SubscribeEdits(nil)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := nav.Select("0"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
