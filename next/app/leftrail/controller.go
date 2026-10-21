package leftrail

import (
	"rectablet/next/contracts"
)

// Controller renders the current input and binds its actions.
type Controller struct{}

func (Controller) Compose(c *contracts.RenderContext) {
	if c.State.LeftRailActions != nil {
		renderActionRailComponents(c)
	} else {
		renderHomeRailComponent(c)
		renderBackButtonComponent(c)
	}
}
