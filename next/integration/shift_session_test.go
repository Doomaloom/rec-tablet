package integration_test

import (
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"rectablet/next/app"
	"rectablet/next/features/login"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/ui"
)

func sessionShot(t *testing.T, window fyne.Window, name string) {
	t.Helper()
	if dir := os.Getenv("REC_TABLET_SCREENSHOTS"); dir != "" {
		check(t, os.MkdirAll(dir, 0755))
		file := must(os.Create(filepath.Join(dir, name+".png")))
		check(t, png.Encode(file, window.Canvas().Capture()))
		check(t, file.Close())
	}
}
func tapSession(t *testing.T, view *app.View, id string) {
	t.Helper()
	button := view.Buttons[id]
	if button == nil {
		t.Fatalf("missing %s", id)
	}
	button.CanActivate = func() bool { return true }
	test.Tap(button)
}

func TestShiftLoginKeypadAndSettings(t *testing.T) {
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		t.Run(string(mode), func(t *testing.T) {
			a := test.NewApp()
			a.Settings().SetTheme(ui.AppTheme{})
			w := a.NewWindow("Shift login")
			var received string
			var finish func(error)
			calls := 0
			c := login.New(login.Options{DisplayMode: mode, Server: func() string { return "https://share.example" }, SaveServer: func(string) error { return nil }, Login: func(code string, done func(error)) { calls++; received, finish = code, done }})
			view := app.New(c, w)
			w.Show()
			t.Cleanup(func() { view.Close(); w.Close(); a.Quit() })
			sessionShot(t, w, "login-"+string(mode))
			if !view.Buttons["login:submit"].Disabled {
				t.Fatal("incomplete code can submit")
			}
			for id, button := range view.Buttons {
				if id == "title" {
					continue
				}
				position := button.Position()
				if button.Size().Height < 44 || position.Y < 0 || position.Y+button.Size().Height > w.Canvas().Size().Height {
					t.Fatalf("unusable login control %s: %v %v", id, position, button.Size())
				}
			}
			tapSession(t, view, "login:9")
			tapSession(t, view, "login:Delete")
			tapSession(t, view, "login:8")
			tapSession(t, view, "login:Clear")
			for _, digit := range []string{"0", "0", "1", "2", "3", "4"} {
				tapSession(t, view, "login:"+digit)
			}
			tapSession(t, view, "login:submit")
			tapSession(t, view, "login:submit")
			if received != "001234" || calls != 1 {
				t.Fatalf("login received %q (%d calls)", received, calls)
			}
			sessionShot(t, w, "login-loading-"+string(mode))
			finish(errors.New("Session code is invalid or sharing has ended. Ask for a new code."))
			sessionShot(t, w, "login-error-"+string(mode))
			visible := false
			for _, object := range view.Root.Objects {
				if text, ok := object.(*ui.TextBlock); ok && text.Text == c.Snapshot().Notice {
					visible = true
					if text.Truncated {
						t.Fatal("login error truncated")
					}
				}
			}
			if !visible || view.Buttons["login:submit"].Disabled {
				t.Fatal("failure not visible or retry disabled")
			}
			// The Clear-button hold reaches device settings without
			// exposing classes or saved-instructor switching.
			view.Buttons["login:Clear"].OnHold()
			if c.Snapshot().Screen != "login_setup" || view.Buttons["row:server"] == nil || view.Buttons["row:display"] == nil || view.Buttons["row:export"] != nil {
				t.Fatal("unexpected logged-out settings")
			}
			tapSession(t, view, "row:server")
			w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
			if w.Canvas().Overlays().Top() != nil {
				t.Fatal("server modal not dismissed")
			}
			_ = c.Back()
			if c.Snapshot().Screen != "login" {
				t.Fatal("back escaped login")
			}
			oldKey := w.Canvas().OnTypedKey()
			staleButton := view.Buttons["login:1"]
			view.Close()
			oldKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
			test.Tap(staleButton)
			if calls != 1 {
				t.Fatal("closed view remained active")
			}
		})
	}
}

func TestShiftLogoutConfirmationAndDetachedNavigation(t *testing.T) {
	for _, mode := range []ui.DisplayMode{ui.DisplayCompact, ui.DisplayLarge} {
		t.Run(string(mode), func(t *testing.T) {
			f := setupUI(t)
			f.c.ConfigureDisplay(mode, nil)
			loggedOut := 0
			f.c.ConfigureSession(func() { loggedOut++; f.view.Close() }, "https://share.example", nil)
			sessionShot(t, f.w, "shift-home-"+string(mode))
			if f.view.Buttons["home"].Text != "Log out" {
				t.Fatal("logout missing on My Classes")
			}
			oldKey := f.w.Canvas().OnTypedKey()
			for _, dismiss := range []string{"Cancel", "Escape"} {
				f.tap(t, "home")
				overlay := f.w.Canvas().Overlays().Top()
				if overlay == nil || loggedOut != 0 {
					t.Fatal("logout skipped confirmation")
				}
				sessionShot(t, f.w, "logout-confirm-"+string(mode))
				if dismiss == "Escape" {
					f.w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
				} else {
					label := dismiss
					button := dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == label })
					test.Tap(button)
				}
				if loggedOut != 0 || f.w.Canvas().Overlays().Top() != nil || f.c.Snapshot().Screen != "classes" {
					t.Fatal("dismissal logged out or left confirmation")
				}
			}
			// Other screens retain their normal Home action.
			f.row(t, 0)
			sessionShot(t, f.w, "shift-navigation-"+string(mode))
			if f.view.Buttons["home"].Text != "Home" {
				t.Fatal("logout replaced navigation outside home")
			}
			f.tap(t, "home")
			f.tap(t, "home")
			overlay := f.w.Canvas().Overlays().Top()
			accept := dialogWidget(t, overlay, func(b *actionbutton.Button) bool { return b.AccessibilityLabel() == "Log out" })
			test.Tap(accept)
			test.Tap(accept)
			if loggedOut != 1 || f.w.Canvas().Overlays().Top() != nil {
				t.Fatal("confirmation did not end session exactly once")
			}
			check(t, f.c.Maintenance())
			if f.c.Snapshot().Screen != "classes" {
				t.Fatal("detached workflow navigated")
			}
			// Replace the workflow, then exercise a retained handler from the old view.
			gate := login.New(login.Options{DisplayMode: mode})
			v := app.New(gate, f.w)
			defer v.Close()
			oldKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
			f.w.Canvas().OnTypedKey()(&fyne.KeyEvent{Name: fyne.KeyEscape})
			if gate.Snapshot().Screen != "login" {
				t.Fatal("old navigation escaped the new gate")
			}
			v.Buttons["login:Clear"].OnHold()
			if gate.Snapshot().Screen != "login_setup" {
				t.Fatal("replacement settings action not active")
			}
		})
	}
}
