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
	presenter := &compositionPresenter{state: State{Screen: "classes", Title: "My Classes", Subtitle: "Instructor", PageCount: 2, TitleDetails: true}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)

	wanted := []string{"background", "home-rail", "previous-page", "next-page", "screen-title", "screen-context", "progress-footer", "device-status"}
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
	if len(composition.Buttons) != 5 {
		t.Fatalf("got %d registered buttons", len(composition.Buttons))
	}
}

func TestComposeScreenUsesCompactSplitHeader(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("split header")
	defer window.Close()
	presenter := &compositionPresenter{state: State{
		Title: "Full student and skill", HideContext: true,
		SplitHeader: &SplitHeader{Left: "8. Front crawl 50m", Right: "Alexandria M."},
	}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	if hasComponent(composition.Components, "screen-title") || hasComponent(composition.Components, "screen-context") {
		t.Fatal("split header should replace the normal title/context")
	}
	for i, id := range []string{"header-half-0", "header-half-1"} {
		x, y, width, height := componentNamed(composition.Components, id).Position(800, 480)
		if x != 100+float32(i)*304 || y != 14 || width != 296 || height != 32 {
			t.Fatalf("incorrect split header slot: %s at %v", id, []float32{x, y, width, height})
		}
		button := composition.Buttons[id]
		if !button.SingleLine || !button.Compact || button.FontSize != 18 {
			t.Fatal("split header should use compact, single-line 18pt text")
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
		surface.Add(text, func(float32, float32) (float32, float32, float32, float32) { return 100, 118, 500, 100 })
	}
	composition := ComposeScreen(view, presenter.state)
	seen := map[string]bool{}
	for _, component := range composition.Components {
		seen[component.Name] = true
	}
	if seen["notice-background"] || seen["notice-message"] || seen["notice-dismiss"] {
		t.Fatal("transient notice popup components were composed")
	}
	if !seen["feature-body-element-6"] {
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
	if x != 12 || y != 14 || width != 76 || height != 452 {
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
	if homeX != 12 || homeY != 246 || homeWidth != 76 || homeHeight != 220 {
		t.Fatalf("split Home rail slot = %v, %v, %v, %v", homeX, homeY, homeWidth, homeHeight)
	}
	if backX != 12 || backY != 14 || backWidth != 76 || backHeight != 220 {
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
	if hasComponent(composition.Components, "screen-title") {
		t.Fatal("mode switch should replace the swimmer title")
	}
	if composition.Buttons["header-switch-0"].Fill != ui.Blue || composition.Buttons["header-switch-1"].Fill != ui.White {
		t.Fatal("selected mode switch option styling is incorrect")
	}
	lastRight := float32(100)
	for _, id := range []string{"header-switch-0", "header-switch-1", "header-switch-2"} {
		x, _, width, _ := componentNamed(composition.Components, id).Position(800, 480)
		if x < lastRight || x+width > 700 {
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
		if x != 12 || width != 76 || y <= bottom || y+height > 466 {
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
	if hasComponent(composition.Components, "progress-footer") || hasComponent(composition.Components, "device-status") {
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
