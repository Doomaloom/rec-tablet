package background

import (
	"rectablet/next/contracts"
)

// Controller renders the current input and binds its actions.
type Controller struct{}

func (Controller) Compose(c *contracts.RenderContext) { renderBackgroundComponent(c) }
