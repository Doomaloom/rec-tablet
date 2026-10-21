package checklist

import (
	"rectablet/next/features/assessment/internal/common"
)

// Controller binds criterion selections; the parent chooses grid or compact rendering.
type Controller struct{ Actions common.Actions }
