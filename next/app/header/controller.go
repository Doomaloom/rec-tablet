package header

import (
	"rectablet/next/contracts"
)

// Controller renders the current input and binds its actions.
type Controller struct{}

func (Controller) Compose(c *contracts.RenderContext) {
	if c.State.HideHeader {
		return
	}
	renderHeaderComponents(c)
	renderPrimaryInfoComponent(c)
}
