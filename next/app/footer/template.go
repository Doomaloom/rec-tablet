package footer

import (
	"rectablet/next/app/internal/geometry"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

func renderFooterComponents(builder *contracts.RenderContext) {
	state := builder.State
	if state.HideFooter || state.FooterAction == "" {
		return
	}
	if state.FooterAction != "" {
		submitButton := builder.Button("submit", state.FooterAction, func() { _ = builder.Controller.Submit() })
		submitButton.Disabled = !state.CanSubmit
		submitButton.Fill = ui.Navy
		submitButton.FontSize = 20
		builder.Add("submit-footer", submitButton, geometry.SubmitFooterSlot)
	}
}
