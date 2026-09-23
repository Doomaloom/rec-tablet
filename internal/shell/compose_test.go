package shell

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"rectablet/internal/shared/ui"
	"testing"
)

type compositionPresenter struct {
	state   State
	changed func()
}

func (presenter *compositionPresenter) Snapshot() State           { return presenter.state }
func (presenter *compositionPresenter) Subscribe(callback func()) { presenter.changed = callback }
func (presenter *compositionPresenter) Home() error               { return nil }
func (presenter *compositionPresenter) Back() error               { return nil }
func (presenter *compositionPresenter) Page(int) error            { return nil }
func (presenter *compositionPresenter) Header() error             { return nil }
func (presenter *compositionPresenter) Submit() error             { return nil }
func (presenter *compositionPresenter) Dismiss()                  {}
func (presenter *compositionPresenter) OpenSetup() error          { return nil }

func TestComposeScreenUsesNamedCommonComponents(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	presenter := &compositionPresenter{state: State{Screen: "classes", Title: "My Classes", PrimaryInfo: "Instructor", PageCount: 2}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)

	wanted := []string{"background", "home-rail", "previous-page", "next-page", "screen-title", "primary-info"}
	seen := map[string]bool{}
	for _, component := range composition.Components {
		if component.Name == "" || component.Object == nil || component.Position == nil {
			t.Fatalf("invalid component: %+v", component)
		}
		if seen[component.Name] {
			t.Fatalf("duplicate component name: %s", component.Name)
		}
		seen[component.Name] = true
	}
	for _, name := range wanted {
		if !seen[name] {
			t.Fatalf("missing named component: %s", name)
		}
	}
	if len(composition.Buttons) != 3 {
		t.Fatalf("got %d registered buttons", len(composition.Buttons))
	}
}

func TestComposeScreenUsesProminentSplitHeader(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("split header")
	defer window.Close()
	presenter := &compositionPresenter{state: State{
		Title:       "Full student and skill",
		SplitHeader: &SplitHeader{Left: "8. Front crawl 50m", Right: "Alexandria M."},
	}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	if hasComponent(composition.Components, "screen-title") || hasComponent(composition.Components, "primary-info") {
		t.Fatal("split header should replace the normal title and omit empty information")
	}
	for i, id := range []string{"header-half-0", "header-half-1"} {
		component := componentNamed(composition.Components, id)
		x, y, width, height := component.Position(800, 480)
		if x != 60+float32(i)*342 || y != 4 || width != 338 || height != SplitHeaderHeight {
			t.Fatalf("incorrect split header slot: %s at %v", id, []float32{x, y, width, height})
		}
		label, ok := component.Object.(*ui.TextBlock)
		if !ok || !label.Wrap || label.FontSize != 22 || label.MaxLines != -1 || label.MinimumFont != 16 || composition.Buttons[id] != nil {
			t.Fatal("split header should use prominent wrapping text")
		}
	}
}

func TestComposeScreenPreservesFeatureBodyWithoutPopupNotices(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	presenter := &compositionPresenter{state: State{Screen: "assessment", Title: "Swimmer", Notice: "Saved", Error: true}}
	view := New(presenter, window)
	defer view.Close()
	presenter.state.ComposeBody = func(surface *ui.Surface) {
		text := ui.NewText("Feature content", 20, ui.White)
		surface.AddNamed("fixture", text, surface.BodyBounds)
	}
	composition := ComposeScreen(view, presenter.state)
	seen := map[string]bool{}
	for _, component := range composition.Components {
		seen[component.Name] = true
	}
	if seen["notice-background"] || seen["notice-message"] || seen["notice-dismiss"] {
		t.Fatal("transient notice popup components were composed")
	}
	if !seen["feature-fixture"] {
		t.Fatalf("feature body was not named: %v", seen)
	}
}

func TestComposeScreenSharesLeftRailBetweenHomeAndBack(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	defer window.Close()

	withoutBack := &compositionPresenter{state: State{Title: "My Classes", PageCount: 1}}
	withoutBackView := New(withoutBack, window)
	withoutBackComposition := ComposeScreen(withoutBackView, withoutBack.state)
	withoutBackView.Close()

	fullHome := componentNamed(withoutBackComposition.Components, "home-rail")
	x, y, width, height := fullHome.Position(800, 480)
	if x != 4 || y != 4 || width != 52 || height != 472 {
		t.Fatalf("full Home rail slot = %v, %v, %v, %v", x, y, width, height)
	}
	if hasComponent(withoutBackComposition.Components, "back-button") {
		t.Fatal("Back should be hidden when the route cannot go back")
	}

	withBack := &compositionPresenter{state: State{Title: "Class menu", CanBack: true, PageCount: 1}}
	withBackView := New(withBack, window)
	withBackComposition := ComposeScreen(withBackView, withBack.state)
	withBackView.Close()

	homeX, homeY, homeWidth, homeHeight := componentNamed(withBackComposition.Components, "home-rail").Position(800, 480)
	backX, backY, backWidth, backHeight := componentNamed(withBackComposition.Components, "back-button").Position(800, 480)
	if homeX != 4 || homeY != 242 || homeWidth != 52 || homeHeight != 234 {
		t.Fatalf("split Home rail slot = %v, %v, %v, %v", homeX, homeY, homeWidth, homeHeight)
	}
	if backX != 4 || backY != 4 || backWidth != 52 || backHeight != 234 {
		t.Fatalf("split Back rail slot = %v, %v, %v, %v", backX, backY, backWidth, backHeight)
	}
}

func TestComposeScreenRendersHeaderModeSwitch(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	defer window.Close()
	presenter := &compositionPresenter{state: State{
		Title: "Choose a Swimmer",
		HeaderSwitch: []HeaderSwitchOption{
			{Label: "Single", Selected: true},
			{Label: "Multi"},
			{Label: "Class"},
		},
	}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	if !hasComponent(composition.Components, "header-switch-0") || !hasComponent(composition.Components, "header-switch-1") || !hasComponent(composition.Components, "header-switch-2") {
		t.Fatal("mode switch buttons were not composed")
	}
	if !hasComponent(composition.Components, "screen-title") {
		t.Fatal("mode switch should keep the screen title visible")
	}
	if composition.Buttons["header-switch-0"].Fill != ui.Blue || composition.Buttons["header-switch-1"].Fill != ui.White {
		t.Fatal("selected mode switch option styling is incorrect")
	}
	lastRight := ui.ContentInset
	for _, id := range []string{"header-switch-0", "header-switch-1", "header-switch-2"} {
		x, _, width, _ := componentNamed(composition.Components, id).Position(800, 480)
		if x < lastRight || x+width > 800-ui.ContentInset {
			t.Fatal("mode switch overlaps another option or rail")
		}
		lastRight = x + width
	}
}

func TestComposeScreenOrdersFourRailActions(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	defer window.Close()
	presenter := &compositionPresenter{state: State{CanBack: true, LeftRailActions: []RailAction{
		{ID: "save-student"}, {ID: "discard-student"}, {ID: "discard-all"}, {ID: "save-all"},
	}}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	if composition.Buttons["home"] != nil || composition.Buttons["back"] != nil {
		t.Fatal("ordered rail must replace Home/Back")
	}
	bottom := float32(0)
	for _, action := range presenter.state.LeftRailActions {
		x, y, width, height := componentNamed(composition.Components, action.ID).Position(800, 480)
		if x != 4 || width != 52 || y <= bottom || y+height > 476 {
			t.Fatalf("rail action outside its slot: %s", action.ID)
		}
		bottom = y + height
	}
}

func TestComposeScreenSupportsIconRailActionsAndHiddenFooter(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("composition")
	defer window.Close()
	icon := fyne.NewStaticResource("action.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><path d="M1 5l3 3 5-7" fill="none" stroke="#fff"/></svg>`))
	presenter := &compositionPresenter{state: State{
		Title:          "Multi Mark",
		CanBack:        true,
		HideFooter:     true,
		HomeRailAction: &RailAction{Label: "Save", Icon: icon},
		BackRailAction: &RailAction{Label: "Discard", Icon: icon},
	}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	if !composition.Buttons["home"].IconOnly || !composition.Buttons["back"].IconOnly {
		t.Fatal("icon rail actions should hide their fallback text")
	}
	if hasComponent(composition.Components, "primary-info") {
		t.Fatal("hidden matrix footer was still composed")
	}
}

func componentNamed(components []LayoutComponent, name string) LayoutComponent {
	for _, component := range components {
		if component.Name == name {
			return component
		}
	}
	panic("missing component: " + name)
}

func hasComponent(components []LayoutComponent, name string) bool {
	for _, component := range components {
		if component.Name == name {
			return true
		}
	}
	return false
}
