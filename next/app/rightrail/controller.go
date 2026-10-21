package rightrail

import (
	"rectablet/next/contracts"
)

// Controller renders the current input and binds its actions.
type Controller struct{}

func (Controller) Compose(c *contracts.RenderContext) {
	if c.State.RightRailActions != nil {
		renderActionRailComponents(c)
		return
	}
	renderPagingRailComponents(c)
	renderPagePositionIndicatorComponents(c)
}
