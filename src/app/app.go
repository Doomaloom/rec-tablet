// Package app wires the desktop shell to the currently enabled workflow.
package app

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"rectablet/src/platform/instance"
	"rectablet/src/shared/ui"
	"rectablet/src/shell"
	"rectablet/src/workflows/swimming"
	store "rectablet/src/workflows/swimming/repository"
	"rectablet/src/workflows/swimming/setup"
	"rectablet/src/workflows/swimming/transfer"
	"sync"
	"time"
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
	profiles, e := newProfileManager(path)
	if e != nil {
		return e
	}
	lock, e := instance.Acquire(path)
	if e != nil {
		return e
	}
	s, e := store.Open(path)
	if e != nil {
		lock.Close()
		return e
	}
	if e = profiles.registerStore(s); e != nil {
		s.Close()
		lock.Close()
		return e
	}
	if target := profiles.activePath(); target != path {
		_ = s.Close()
		_ = lock.Close()
		if _, e = os.Stat(target); e != nil {
			return fmt.Errorf("saved session database is missing: %w", e)
		}
		lock, e = instance.Acquire(target)
		if e != nil {
			return e
		}
		s, e = store.Open(target)
		if e != nil {
			lock.Close()
			return e
		}
	}
	defer func() { _ = s.Close(); _ = lock.Close() }()
	a := fyneapp.NewWithID("org.rectablet.app")
	a.Settings().SetTheme(ui.AppTheme{})
	w := a.NewWindow("Rec Tablet")
	var view *shell.View
	var mount func()
	var importMu sync.Mutex
	var importGeneration uint64
	var cancelImport context.CancelFunc
	cancelCurrentImport := func() {
		importMu.Lock()
		defer importMu.Unlock()
		if cancelImport != nil {
			cancelImport()
			cancelImport = nil
		}
	}
	defer cancelCurrentImport()
	switchProfile := func(key string) error {
		target, ok := profiles.pathFor(key)
		if !ok {
			return fmt.Errorf("saved session is unavailable")
		}
		if _, err := os.Stat(target); err != nil {
			return fmt.Errorf("saved session database is missing: %w", err)
		}
		if target == s.DatabasePath() {
			return profiles.activate(key)
		}
		newLock, err := instance.Acquire(target)
		if err != nil {
			return err
		}
		newStore, err := store.Open(target)
		if err != nil {
			newLock.Close()
			return err
		}
		if err = profiles.activate(key); err != nil {
			newStore.Close()
			newLock.Close()
			return err
		}
		if view != nil {
			view.Close()
		}
		_ = s.Close()
		_ = lock.Close()
		s, lock = newStore, newLock
		mount()
		return nil
	}
	joinShare := func(code string, done func(error)) {
		importMu.Lock()
		if cancelImport != nil {
			cancelImport()
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancelImport = cancel
		importGeneration++
		generation := importGeneration
		importMu.Unlock()
		finishImport := func() {
			importMu.Lock()
			if importGeneration == generation {
				cancelImport = nil
			}
			importMu.Unlock()
		}
		go func() {
			pkg, err := transfer.FetchSharedPackage(ctx, profiles.shareServer(options.ShareServer), code)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				fyne.Do(func() {
					finishImport()
					if ctx.Err() == nil {
						done(err)
					}
				})
				return
			}
			if ctx.Err() != nil {
				return
			}
			fyne.Do(func() {
				defer finishImport()
				if ctx.Err() != nil {
					return
				}
				dataset, metaErr := s.Meta("dataset_id")
				if metaErr != nil {
					done(metaErr)
					return
				}
				instructor, metaErr := s.Meta("instructor")
				if metaErr != nil {
					done(metaErr)
					return
				}
				key := profileKey(pkg.DatasetID, pkg.Instructor)
				profile := savedProfile{Key: key, DatasetID: pkg.DatasetID, Instructor: pkg.Instructor}
				if dataset == pkg.DatasetID && instructor == pkg.Instructor {
					if _, err = s.Import(pkg); err == nil {
						profile, err = profiles.registerPackage(pkg, s.DatabasePath())
					}
					if err == nil {
						err = profiles.activate(profile.Key)
					}
					if err != nil {
						done(err)
						return
					}
					if view != nil {
						view.Close()
					}
					mount()
					done(nil)
					return
				}
				profilePath := profiles.packagePath(pkg)
				if openErr := os.MkdirAll(filepath.Dir(profilePath), 0700); openErr != nil {
					done(openErr)
					return
				}
				newLock, openErr := instance.Acquire(profilePath)
				if openErr != nil {
					done(openErr)
					return
				}
				profile, err = profiles.importPackage(pkg)
				if err != nil {
					newLock.Close()
					done(err)
					return
				}
				newStore, openErr := store.Open(profile.Path)
				if openErr != nil {
					newLock.Close()
					done(openErr)
					return
				}
				if err = profiles.activate(profile.Key); err != nil {
					newStore.Close()
					newLock.Close()
					done(err)
					return
				}
				if view != nil {
					view.Close()
				}
				_ = s.Close()
				_ = lock.Close()
				s, lock = newStore, newLock
				mount()
				done(nil)
			})
		}()
	}
	mount = func() {
		controller := swimming.New(s)
		controller.ConfigureDisplay(ui.DisplayMode(a.Preferences().StringWithFallback("display-mode", string(ui.DisplayCompact))), func(mode ui.DisplayMode) {
			a.Preferences().SetString("display-mode", string(mode.Normalize()))
		})
		profileOptions := make([]setup.ProfileOption, 0)
		for _, item := range profiles.rows() {
			profileOptions = append(profileOptions, setup.ProfileOption{Key: item.Key, Label: item.Label})
		}
		if dates, err := s.Dates(); err == nil {
			controller.SetDate(nearestLessonDate(dates, time.Now().Format("2006-01-02")))
		}
		controller.ConfigureDeviceSessions(joinShare, profileOptions, switchProfile, profiles.shareServer(options.ShareServer), profiles.setShareServer, cancelCurrentImport)
		view = shell.New(controller, w)
	}
	mount()
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
