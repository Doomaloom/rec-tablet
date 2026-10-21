package timedheat

import (
	"fmt"
	"time"
)

type ParticipantState struct {
	Participant Participant
	CompletedAt time.Duration
	Completed   bool
}

type Participant struct {
	AttemptID, PersonID, Name   string
	LapsRequired, LapsCompleted int
}

type Heat struct {
	SkillID, Version, Lesson string
	LapsRequired             int
	Participants             []ParticipantState
	LapsInput                string
	StartedAt                time.Time
	Elapsed                  time.Duration
	Running                  bool
	Stop                     chan struct{}
}

func formatElapsed(elapsed time.Duration) string {
	if elapsed < 0 {
		elapsed = 0
	}
	minutes := int(elapsed / time.Minute)
	seconds := int(elapsed/time.Second) % 60
	tenths := int(elapsed/time.Millisecond) / 100
	return fmt.Sprintf("%02d:%02d.%d", minutes, seconds, tenths)
}
