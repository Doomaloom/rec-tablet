package runtime

import (
	"image"
	"image/color"
	"rectablet/next/features/login"
	"testing"
)

func TestBlankCaptureIsReported(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	if validateCapture(img) == nil {
		t.Fatal("blank screenshot reported as successful")
	}
	img.Set(2, 2, color.White)
	if err := validateCapture(img); err != nil {
		t.Fatal(err)
	}
}

func TestLogoutResultKeepsDigitsTypedWhileUploading(t *testing.T) {
	var submitted string
	screen := login.New(login.Options{Login: func(code string, done func(error)) { submitted = code; done(nil) }})
	for _, d := range []string{"1", "2", "3"} {
		screen.Digit(d)
	}
	showLoginNotice(screen, nil, "Results uploaded")
	for _, d := range []string{"4", "5", "6"} {
		screen.Digit(d)
	}
	_ = screen.Submit()
	if submitted != "123456" {
		t.Fatalf("typed digits were lost: %q", submitted)
	}
}
