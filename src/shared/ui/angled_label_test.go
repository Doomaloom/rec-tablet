package ui

import (
	"testing"
)

func TestAngledLabelRendersClockwiseSlope(t *testing.T) {
	label := NewAngledLabel("1. Streamline", 350, 15, 12, White, Navy, Navy)
	image := label.raster.Generator(int(label.MinSize().Width), int(label.MinSize().Height))
	minimumX, maximumX, maximumY := image.Bounds().Dx(), -1, -1
	for y := 0; y < image.Bounds().Dy(); y++ {
		for x := 0; x < image.Bounds().Dx(); x++ {
			_, _, _, alpha := image.At(x, y).RGBA()
			if alpha > 0 {
				minimumX = min(minimumX, x)
				maximumX = max(maximumX, x)
				maximumY = max(maximumY, y)
			}
		}
	}
	if maximumX < minimumX {
		t.Fatal("angled label rendered no visible pixels")
	}
	third := max(1, (maximumX-minimumX+1)/3)
	leftTop, rightTop := image.Bounds().Dy(), image.Bounds().Dy()
	for y := 0; y < image.Bounds().Dy(); y++ {
		for x := minimumX; x < minimumX+third; x++ {
			_, _, _, alpha := image.At(x, y).RGBA()
			if alpha > 0 && y < leftTop {
				leftTop = y
			}
		}
		for x := maximumX - third + 1; x <= maximumX; x++ {
			_, _, _, alpha := image.At(x, y).RGBA()
			if alpha > 0 && y < rightTop {
				rightTop = y
			}
		}
	}
	if rightTop <= leftTop {
		t.Fatalf("label did not slope down to the right: left top %d, right top %d", leftTop, rightTop)
	}
	if maximumX < image.Bounds().Dx()-3 || maximumY < image.Bounds().Dy()-3 {
		t.Fatalf("label did not anchor to its grid corner: right %d, bottom %d", maximumX, maximumY)
	}
}

func TestAngledLabelsUseVariableHeightForWrappedText(t *testing.T) {
	short := NewAngledLabel("1 Entry", 350, 15, 12, White, Navy, Navy)
	wrapped := NewAngledLabel("2 Safe movement and control during entry and exit", 350, 15, 12, White, Navy, Navy)
	if wrapped.MinSize().Height <= short.MinSize().Height {
		t.Fatalf("wrapped header did not grow in height: short=%v wrapped=%v", short.MinSize().Height, wrapped.MinSize().Height)
	}
	face := newAngledLabelFace(15)
	defer face.Close()
	lines := wrapAngledText("2 Safe movement and control during entry and exit", face, 330)
	if len(lines) == 0 || short.CardHeight >= wrapped.CardHeight {
		t.Fatalf("wrapped header did not reserve extra height: lines=%v short=%f wrapped=%f", lines, short.CardHeight, wrapped.CardHeight)
	}
}
