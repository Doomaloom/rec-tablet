package app

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"

	"rectablet/next/app/internal/geometry"
	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

type compositionPresenter struct {
	state   contracts.State
	changed func()
}

func (presenter *compositionPresenter) Snapshot() contracts.State { return presenter.state }
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
	presenter := &compositionPresenter{state: contracts.State{Screen: "classes", Title: "My Classes", PrimaryInfo: "Instructor", PageCount: 2}}
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

func TestPagePositionIndicatorTracksTheCurrentPage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("page position")
	defer window.Close()
	presenter := &compositionPresenter{state: contracts.State{Page: 1, PageCount: 4}}
	view := New(presenter, window)
	defer view.Close()

	composition := ComposeScreen(view, presenter.state)
	track := componentNamed(composition.Components, "page-position-track")
	thumb := componentNamed(composition.Components, "page-position-thumb")
	tx, ty, tw, th := track.Position(800, 480)
	if tx != 724 || ty != 52 || tw != 12 || th != 424 {
		t.Fatalf("unexpected page track bounds: %v", []float32{tx, ty, tw, th})
	}
	thumbX, thumbY, thumbWidth, thumbHeight := thumb.Position(800, 480)
	if thumbX != tx || thumbY != 158 || thumbWidth != tw || thumbHeight != 106 {
		t.Fatalf("unexpected page 2 of 4 thumb bounds: %v", []float32{thumbX, thumbY, thumbWidth, thumbHeight})
	}
	if hasComponent(composition.Components, "page-position-label") || composition.Buttons["page-position"] != nil {
		t.Fatal("the page indicator should not add a label or interactive control")
	}
	trackRect, ok := track.Object.(*canvas.Rectangle)
	if !ok || trackRect.FillColor != ui.Rail || trackRect.StrokeColor != ui.Muted {
		t.Fatal("page track should have a clearly visible neutral fill and outline")
	}
	thumbRect, ok := thumb.Object.(*canvas.Rectangle)
	if !ok || thumbRect.FillColor != ui.Blue || thumbRect.StrokeColor != ui.Navy {
		t.Fatal("current page thumb should use the high-contrast blue treatment")
	}
	if x, _, width, _ := geometry.BodySlot(presenter.state)(800, 480); x+width != 716 {
		t.Fatalf("paged body should reserve the indicator lane; bounds end at %v", x+width)
	}

	for _, size := range []fyne.Size{fyne.NewSize(800, 480), fyne.NewSize(1200, 720), fyne.NewSize(1280, 800)} {
		scale := min(size.Width/800, size.Height/480)
		logicalWidth, logicalHeight := size.Width/scale, size.Height/scale
		bodyX, bodyY, bodyWidth, bodyHeight := geometry.BodySlot(presenter.state)(logicalWidth, logicalHeight)
		trackX, trackY, trackWidth, trackHeight := geometry.PagePositionTrackSlot(presenter.state)(logicalWidth, logicalHeight)
		railX, _, _, _ := geometry.PreviousPageSlot(logicalWidth, logicalHeight)
		if trackX-(bodyX+bodyWidth) != geometry.PagePositionGutter || railX-(trackX+trackWidth) != geometry.PagePositionGutter || trackY != bodyY || trackHeight != bodyHeight {
			t.Fatalf("indicator lane does not fit %v: body %v, track %v, rail x %v", size,
				[]float32{bodyX, bodyY, bodyWidth, bodyHeight}, []float32{trackX, trackY, trackWidth, trackHeight}, railX)
		}
		if (trackX+trackWidth)*scale > size.Width || (trackY+trackHeight)*scale > size.Height {
			t.Fatalf("indicator extends outside %v", size)
		}
	}
}

func TestPagePositionIndicatorHandlesEndpointsAndSinglePage(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		page  int
		wantY float32
	}{
		{name: "first", page: 0, wantY: 52},
		{name: "second", page: 1, wantY: 158},
		{name: "last", page: 3, wantY: 370},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, y, _, _ := geometry.PagePositionThumbSlot(contracts.State{Page: testCase.page, PageCount: 4})(800, 480)
			if y != testCase.wantY {
				t.Fatalf("thumb y = %v, want %v", y, testCase.wantY)
			}
		})
	}

	single := contracts.State{Page: 0, PageCount: 1}
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("single page")
	defer window.Close()
	presenter := &compositionPresenter{state: single}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, single)
	if !hasComponent(composition.Components, "previous-page") {
		t.Fatal("single-page screens should retain their existing paging controls")
	}
	if hasComponent(composition.Components, "page-position-track") || hasComponent(composition.Components, "page-position-thumb") {
		t.Fatal("single-page screens should not show a position indicator")
	}
	if x, _, width, _ := geometry.BodySlot(single)(800, 480); x+width != 740 {
		t.Fatalf("single-page body should not reserve an indicator lane; bounds end at %v", x+width)
	}
}

func TestPagePositionThumbHeightIsOnePageOfTheTrack(t *testing.T) {
	_, _, _, trackHeight := geometry.PagePositionTrackSlot(contracts.State{Page: 0, PageCount: 2})(800, 480)
	for _, pageCount := range []int{2, 4} {
		for _, page := range []int{0, pageCount - 1} {
			_, y, _, thumbHeight := geometry.PagePositionThumbSlot(contracts.State{Page: page, PageCount: pageCount})(800, 480)
			if thumbHeight*float32(pageCount) != trackHeight {
				t.Fatalf("%d page thumb height %v is not one page of track height %v", pageCount, thumbHeight, trackHeight)
			}
			wantY := float32(52)
			if page == pageCount-1 {
				wantY += trackHeight - thumbHeight
			}
			if y != wantY {
				t.Fatalf("page %d of %d thumb y = %v, want %v", page+1, pageCount, y, wantY)
			}
		}
	}
}

func TestComposeScreenUsesProminentSplitHeader(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("split header")
	defer window.Close()
	presenter := &compositionPresenter{state: contracts.State{
		Title:       "Full student and skill",
		SplitHeader: &contracts.SplitHeader{Left: "8. Front crawl 50m", Right: "Alexandria M."},
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
		if x != 60+float32(i)*342 || y != 4 || width != 338 || height != contracts.SplitHeaderHeight {
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
	presenter := &compositionPresenter{state: contracts.State{Screen: "assessment", Title: "Swimmer", Notice: "Saved", Error: true}}
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

	withoutBack := &compositionPresenter{state: contracts.State{Title: "My Classes", PageCount: 1}}
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

	withBack := &compositionPresenter{state: contracts.State{Title: "Class menu", CanBack: true, PageCount: 1}}
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
	presenter := &compositionPresenter{state: contracts.State{
		Title: "Choose a Swimmer",
		HeaderSwitch: []contracts.HeaderSwitchOption{
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
	presenter := &compositionPresenter{state: contracts.State{CanBack: true, LeftRailActions: []contracts.RailAction{
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
	presenter := &compositionPresenter{state: contracts.State{
		Title:          "Multi Mark",
		CanBack:        true,
		HideFooter:     true,
		HomeRailAction: &contracts.RailAction{Label: "Save", Icon: icon},
		BackRailAction: &contracts.RailAction{Label: "Discard", Icon: icon},
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

func componentNamed(components []contracts.LayoutComponent, name string) contracts.LayoutComponent {
	for _, component := range components {
		if component.Name == name {
			return component
		}
	}
	panic("missing component: " + name)
}

func hasComponent(components []contracts.LayoutComponent, name string) bool {
	for _, component := range components {
		if component.Name == name {
			return true
		}
	}
	return false
}

func TestClockTitleUpdatePreservesPressedButtons(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := app.NewWindow("clock")
	var update func(string)
	presenter := &compositionPresenter{state: contracts.State{Screen: "timed_heat", Title: "00:00.0", CanBack: true, HeaderActions: []contracts.HeaderAction{{Label: "Pause", Run: func() error { return nil }}}, BindTitle: func(fn func(string)) { update = fn }}}
	view := New(presenter, window)
	defer view.Close()
	composition := ComposeScreen(view, presenter.state)
	var title *ui.TextBlock
	for _, component := range composition.Components {
		if component.Name == "screen-title" {
			title = component.Object.(*ui.TextBlock)
		}
	}
	if title == nil || update == nil {
		t.Fatal("clock title not bound")
	}
	for id, button := range view.Buttons {
		if button.Disabled {
			continue
		}
		tapped := false
		button.CanActivate = func() bool { return true }
		button.OnTap = func() { tapped = true }
		button.TouchDown(nil)
		update("00:01.2")
		button.TouchUp(nil)
		button.Tapped(nil)
		if !tapped || view.Buttons[id] != button {
			t.Fatalf("clock interrupted %s", id)
		}
	}
	if title.Text != "00:01.2" {
		t.Fatal("clock label did not change")
	}
}
