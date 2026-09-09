// Package app wires the desktop shell to the currently enabled workflow.
package app

import (
	"fmt"
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"rectablet/internal/platform/instance"
	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming"
	store "rectablet/internal/workflows/swimming/repository"
	"time"
)

type Options struct {
	Database              string
	Fullscreen, SmokeTest bool
	Screenshot            string
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
	s, e := store.Open(path)
	if e != nil {
		return e
	}
	defer s.Close()
	a := fyneapp.NewWithID("org.rectablet.app")
	a.Settings().SetTheme(ui.AppTheme{})
	w := a.NewWindow("Rec Tablet")
	c := swimming.New(s)
	v := shell.New(c, w)
	defer v.Close()
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
