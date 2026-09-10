package shell

// ComposeScreen is the single assembly point for the shared tablet chrome and
// the feature-owned body. Individual component factories live in components.go.
func ComposeScreen(view *View, state State) ScreenComposition {
	builder := newComponentBuilder(view, state)

	newBackgroundComponent(builder)
	newHomeRailComponent(builder)
	newPagingRailComponents(builder)
	newBackButtonComponent(builder)
	newHeaderComponents(builder)
	newContextComponent(builder)
	newBodyComponent(builder)
	newFooterComponents(builder)

	return ScreenComposition{
		Components: builder.components,
		Buttons:    builder.buttons,
		Criterion:  view.Criterion,
	}
}
