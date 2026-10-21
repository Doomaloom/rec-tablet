package runtime

import (
	"image"
	"image/color"
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
