package shell

import (
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

func TestComposeScreenPreservesFeatureBodyAndNoticeComponents(t *testing.T) {
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
	if !seen["notice-background"] || !seen["notice-message"] || !seen["notice-dismiss"] {
		t.Fatal("notice components were not composed")
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
	if homeX != 12 || homeY != 14 || homeWidth != 76 || homeHeight != 220 {
		t.Fatalf("split Home rail slot = %v, %v, %v, %v", homeX, homeY, homeWidth, homeHeight)
	}
	if backX != 12 || backY != 246 || backWidth != 76 || backHeight != 220 {
		t.Fatalf("split Back rail slot = %v, %v, %v, %v", backX, backY, backWidth, backHeight)
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
