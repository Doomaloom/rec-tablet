package assessment

import (
	"fmt"
	"time"
)

type timedBenchmark struct {
	SkillID     string
	Limit       time.Duration
	Description string
}

var timedBenchmarks = map[string]timedBenchmark{
	"Splash7:10": {SkillID: "Splash7:10", Limit: 3 * time.Minute, Description: "100 m · 3:00 limit (100 yd · 2:40)"},
	"Splash8:9":  {SkillID: "Splash8:9", Limit: 6 * time.Minute, Description: "200 m · 6:00 limit (200 yd · 5:20)"},
	"Splash9:11": {SkillID: "Splash9:11", Limit: 9 * time.Minute, Description: "300 m · 9:00 limit (300 yd · 8:00)"},
}

func benchmarkFor(skillID string) (timedBenchmark, bool) {
	benchmark, ok := timedBenchmarks[skillID]
	return benchmark, ok
}

type timedParticipantState struct {
	participant timedParticipant
	completedAt time.Duration
	completed   bool
}

type timedParticipant struct {
	AttemptID, PersonID, Name   string
	LapsRequired, LapsCompleted int
}

type timedHeatState struct {
	SkillID, Version, Lesson string
	LapsRequired             int
	Participants             []timedParticipantState
	LapsInput                string
	StartedAt                time.Time
	Elapsed                  time.Duration
	Running                  bool
	stop                     chan struct{}
}

func formatHeatElapsed(elapsed time.Duration) string {
	if elapsed < 0 {
		elapsed = 0
	}
	minutes := int(elapsed / time.Minute)
	seconds := int(elapsed/time.Second) % 60
	tenths := int(elapsed/time.Millisecond) / 100
	return fmt.Sprintf("%02d:%02d.%d", minutes, seconds, tenths)
}
