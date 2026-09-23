package shell

// ComposeScreen is the single assembly point for the shared tablet chrome and
// the feature-owned body. Individual component factories live in components.go.
func ComposeScreen(view *View, state State) ScreenComposition {
	builder := newComponentBuilder(view, state)

	newBackgroundComponent(builder)
	if len(state.LeftRailActions) > 0 {
		newActionRailComponents(builder)
	} else {
		newHomeRailComponent(builder)
		newBackButtonComponent(builder)
	}
	newPagingRailComponents(builder)
	newHeaderComponents(builder)
	newPrimaryInfoComponent(builder)
	newBodyComponent(builder)
	newFooterComponents(builder)

	return ScreenComposition{
		Components: builder.components,
		Buttons:    builder.buttons,
		Criterion:  view.Criterion,
	}
}
