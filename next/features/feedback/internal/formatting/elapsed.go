package formatting

import (
	"fmt"
)

func TimedElapsed(milliseconds int64) string {
	if milliseconds < 0 {
		milliseconds = 0
	}
	minutes := milliseconds / 60000
	seconds := (milliseconds / 1000) % 60
	tenths := (milliseconds % 1000) / 100
	return fmt.Sprintf("%02d:%02d.%d", minutes, seconds, tenths)
}
