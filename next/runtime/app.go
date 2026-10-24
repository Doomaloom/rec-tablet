// Package runtime owns application startup, device sessions, and shutdown.
package runtime

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"

	swimming "rectablet/next"
	"rectablet/next/app"
	"rectablet/next/contracts"
	"rectablet/next/features/login"
	"rectablet/src/platform/instance"
	"rectablet/src/shared/ui"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/transfer"
)

type Options struct {
	Database              string
	ShareServer           string
	Fullscreen, SmokeTest bool
	Screenshot            string
}

func nearestLessonDate(dates []string, today string) string {
	if len(dates) == 0 {
		return today
	}
	best := dates[0]
	bestDistance := time.Duration(1<<63 - 1)
	current, _ := time.Parse("2006-01-02", today)
	for _, date := range dates {
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}
		distance := parsed.Sub(current).Abs()
		if distance < bestDistance {
			best, bestDistance = date, distance
		}
	}
	return best
}

func Run(options Options) error {
	path, e := filepath.Abs(options.Database)
	if e != nil {
		return e
	}
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	lock, e := instance.Acquire(path)
	if e != nil {
		return e
	}
	defer lock.Close()
	profiles, e := newProfileManager(path)
	if e != nil {
		return e
	}
	// Register an existing base dataset for compatibility, but never resume it
	// without a newly verified export code.
	base, e := store.Open(path)
	if e != nil {
		return e
	}
	e = profiles.registerStore(base)
	closeErr := base.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	a := fyneapp.NewWithID("org.rectablet.app")
	a.Settings().SetTheme(ui.AppTheme{})
	w := a.NewWindow("Rec Tablet")
	session := &deviceSession{profiles: profiles, server: options.ShareServer, fetch: transfer.FetchSharedPackage, dispatch: fyne.Do}
	defer session.close()
	var view *app.View
	var mountLogin func(error)
	var mountWorkflow func()
	mode := func() ui.DisplayMode {
		return ui.DisplayMode(a.Preferences().StringWithFallback("display-mode", string(ui.DisplayCompact))).Normalize()
	}
	saveMode := func(value ui.DisplayMode) { a.Preferences().SetString("display-mode", string(value.Normalize())) }
	mount := func(controller contracts.Presenter) {
		if view != nil {
			view.Close()
		}
		view = app.New(controller, w)
	}
	mountLogin = func(loginError error) {
		controller := login.New(login.Options{
			DisplayMode: mode(), SaveDisplayMode: saveMode,
			Server:     func() string { return profiles.shareServer(options.ShareServer) },
			SaveServer: profiles.setShareServer,
			Login: func(code string, done func(error)) {
				session.login(code, func(err error) {
					done(err)
					if err == nil {
						mountWorkflow()
					}
				})
			},
			Cancel: session.cancelLogin,
		})
		if loginError != nil {
			controller.Fail(loginError)
		}
		mount(controller)
	}
	mountWorkflow = func() {
		controller := swimming.New(session.store)
		controller.ConfigureDisplay(mode(), saveMode)
		if dates, err := session.store.Dates(); err == nil {
			controller.SetDate(nearestLessonDate(dates, time.Now().Format("2006-01-02")))
		}
		controller.ConfigureSession(func() {
			if view != nil {
				view.Close()
			}
			mountLogin(session.logout())
		}, profiles.shareServer(options.ShareServer), profiles.setShareServer)
		mount(controller)
	}
	mountLogin(nil)
	defer func() {
		if view != nil {
			view.Close()
		}
	}()
	w.SetFullScreen(options.Fullscreen)
	capture := func() error {
		if options.Screenshot == "" {
			return nil
		}
		img := w.Canvas().Capture()
		if e := validateCapture(img); e != nil {
			return e
		}
		if e := os.MkdirAll(filepath.Dir(options.Screenshot), 0755); e != nil {
			return e
		}
		f, e := os.Create(options.Screenshot)
		if e != nil {
			return e
		}
		e = png.Encode(f, img)
		closeErr := f.Close()
		if e != nil {
			return e
		}
		return closeErr
	}
	if headless {
		if !options.SmokeTest {
			return fmt.Errorf("this is a headless verification build; run with --smoke-test or build without -tags ci")
		}
		w.Show()
		e = capture()
		w.Close()
		a.Quit()
		return e
	}
	var smokeErr error
	if options.SmokeTest {
		go func() { time.Sleep(time.Second); fyne.Do(func() { smokeErr = capture(); w.Close(); a.Quit() }) }()
	}
	w.ShowAndRun()
	return smokeErr
}

// Some OpenGL drivers return a blank front buffer despite displaying the app.
// Do not report that as a successful screenshot or replace a previous image.
func validateCapture(img image.Image) error {
	b := img.Bounds()
	if !b.Empty() {
		r, g, blue, a := img.At(b.Min.X, b.Min.Y).RGBA()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r2, g2, b2, a2 := img.At(x, y).RGBA()
				if r != r2 || g != g2 || blue != b2 || a != a2 {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("graphics driver returned a blank screenshot; use make smoke for software capture or capture the desktop window")
}
