package shared

import (
	"fyne.io/fyne/v2"
)

var (
	SaveActionIcon       = fyne.NewStaticResource("assessment-save.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M10 34l14 14 30-32" fill="none" stroke="#fff" stroke-width="8" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
	DiscardActionIcon    = fyne.NewStaticResource("assessment-discard.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M14 14l36 36M50 14L14 50" fill="none" stroke="#fff" stroke-width="8" stroke-linecap="round"/></svg>`))
	SaveAllActionIcon    = fyne.NewStaticResource("assessment-save-all.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M3 32l10 12 23-27M28 35l10 12 23-27" fill="none" stroke="#fff" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
	DiscardAllActionIcon = fyne.NewStaticResource("assessment-discard-all.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64"><path d="M5 21l20 22M25 21L5 43M39 21l20 22M59 21L39 43" fill="none" stroke="#fff" stroke-width="6" stroke-linecap="round"/></svg>`))
)
