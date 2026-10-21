package checklist

import (
	"image/color"

	"rectablet/next/features/assessment/internal/common"
)

// Controller binds criterion selections; the parent chooses grid or compact rendering.
type Controller struct {
	Actions common.Actions
	// Columns overrides the automatic large-checklist column count when positive.
	Columns   int
	Fill, Ink color.Color
}
