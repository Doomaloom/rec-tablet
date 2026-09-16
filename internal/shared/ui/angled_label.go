package ui

import (
	"image"
	"image/color"
	"math"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	angledLabelFontOnce sync.Once
	angledLabelFont     *opentype.Font
)

// AngledLabel renders a variable-height subskill card. The optional tap action
// can expose the full text of a long criterion.
type AngledLabel struct {
	widget.BaseWidget
	OnTap      func()
	CardHeight float64
	raster     *canvas.Raster
}

func NewAngledLabel(text string, length, textSize, angle float64, fill, ink, border color.Color) *AngledLabel {
	var cardHeight float64
	face := newAngledLabelFace(textSize)
	if face != nil {
		if len(text) > 35 {
			cardHeight = float64(2*face.Metrics().Height.Ceil() + 2*angledLabelPadding(textSize))
		} else {
			cardHeight = float64(face.Metrics().Height.Ceil() + angledLabelPadding(textSize))
		}
		_ = face.Close()
	}
	radians := angle * math.Pi / 180
	sine, cosine := math.Abs(math.Sin(radians)), math.Abs(math.Cos(radians))
	slotWidth := math.Ceil(length*cosine+cardHeight*sine) + 2
	slotHeight := math.Ceil(length*sine+cardHeight*cosine) + 2
	label := &AngledLabel{CardHeight: cardHeight}
	label.raster = canvas.NewRaster(func(width, height int) image.Image {
		return renderAngledLabel(text, length, cardHeight, textSize, radians, slotWidth, slotHeight, width, height, fill, ink, border)
	})
	label.raster.SetMinSize(fyne.NewSize(float32(slotWidth), float32(slotHeight)))
	label.ExtendBaseWidget(label)
	return label
}

func (label *AngledLabel) MinSize() fyne.Size { return label.raster.MinSize() }
func (label *AngledLabel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(label.raster)
}
func (label *AngledLabel) Tapped(*fyne.PointEvent) {
	if label.OnTap != nil {
		label.OnTap()
	}
}

func newAngledLabelFace(size float64) font.Face {
	angledLabelFontOnce.Do(func() {
		resource := theme.DefaultTextBoldFont()
		if resource != nil {
			angledLabelFont, _ = opentype.Parse(resource.Content())
		}
	})
	if angledLabelFont == nil {
		return nil
	}
	face, _ := opentype.NewFace(angledLabelFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	return face
}

func angledLabelPadding(textSize float64) int {
	return max(2, int(math.Round(textSize*0.2)))
}

func renderAngledLabel(text string, length, cardHeight, textSize, radians, slotWidth, slotHeight float64, width, height int, fill, ink, border color.Color) image.Image {
	output := image.NewNRGBA(image.Rect(0, 0, max(1, width), max(1, height)))
	if width < 2 || height < 2 {
		return output
	}

	// Render at twice the output resolution so diagonal type and borders remain
	// smooth at the tablet's native 800x480 size.
	const sampling = 2
	scale := math.Min(float64(width)/slotWidth, float64(height)/slotHeight) * sampling
	face := newAngledLabelFace(textSize * scale)
	if face == nil {
		return output
	}
	defer face.Close()
	labelWidth := max(1, int(math.Round(length*scale)))
	labelHeight := max(1, int(math.Round(cardHeight*scale)))
	paddingX := max(1, int(math.Round(10*scale)))
	availableWidth := max(1, labelWidth-2*paddingX)
	lines := wrapAngledText(text, face, availableWidth)
	if len(lines) > 2 {
		lines = lines[:2]
		lines[1] = fitAngledEllipsis(lines[1], face, availableWidth)
	}

	label := image.NewNRGBA(image.Rect(0, 0, labelWidth, labelHeight))
	drawRoundedLabel(label, fill, border, scale)
	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil()
	baseline := (labelHeight-len(lines)*lineHeight)/2 + metrics.Ascent.Ceil()
	for _, line := range lines {
		drawer := font.Drawer{
			Dst:  label,
			Src:  image.NewUniform(ink),
			Face: face,
			Dot:  fixed.P(paddingX, baseline),
		}
		drawer.DrawString(line)
		baseline += lineHeight
	}

	rotated := image.NewNRGBA(image.Rect(0, 0, width*sampling, height*sampling))
	rotateInto(rotated, label, radians)
	draw.CatmullRom.Scale(output, output.Bounds(), rotated, rotated.Bounds(), draw.Src, nil)
	return output
}

func wrapAngledText(text string, face font.Face, availableWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	line := ""
	for _, word := range words {
		if font.MeasureString(face, word).Ceil() > availableWidth {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			chunks := splitAngledWord(word, face, availableWidth)
			lines = append(lines, chunks[:len(chunks)-1]...)
			line = chunks[len(chunks)-1]
			continue
		}
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if line == "" || font.MeasureString(face, candidate).Ceil() <= availableWidth {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
	}
	return append(lines, line)
}

func splitAngledWord(word string, face font.Face, availableWidth int) []string {
	chunks := []string{}
	chunk := ""
	for _, character := range word {
		candidate := chunk + string(character)
		if chunk != "" && font.MeasureString(face, candidate).Ceil() > availableWidth {
			chunks = append(chunks, chunk)
			chunk = ""
		}
		chunk += string(character)
	}
	if chunk != "" || len(chunks) == 0 {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func fitAngledEllipsis(text string, face font.Face, availableWidth int) string {
	runes := []rune(strings.TrimSpace(text))
	for len(runes) > 0 && font.MeasureString(face, string(runes)+"…").Ceil() > availableWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func drawRoundedLabel(target *image.NRGBA, fill, border color.Color, scale float64) {
	width, height := target.Bounds().Dx(), target.Bounds().Dy()
	radius := max(1, int(math.Round(6*scale)))
	borderWidth := max(1, int(math.Round(1.5*scale)))
	fillColor := color.NRGBAModel.Convert(fill).(color.NRGBA)
	borderColor := color.NRGBAModel.Convert(border).(color.NRGBA)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !insideRoundedRectangle(x, y, width, height, radius) {
				continue
			}
			pixel := borderColor
			if insideRoundedRectangle(x-borderWidth, y-borderWidth, width-2*borderWidth, height-2*borderWidth, max(1, radius-borderWidth)) {
				pixel = fillColor
			}
			target.SetNRGBA(x, y, pixel)
		}
	}
}

func insideRoundedRectangle(x, y, width, height, radius int) bool {
	if width <= 0 || height <= 0 || x < 0 || y < 0 || x >= width || y >= height {
		return false
	}
	if x >= radius && x < width-radius || y >= radius && y < height-radius {
		return true
	}
	cx := radius
	if x >= width-radius {
		cx = width - radius - 1
	}
	cy := radius
	if y >= height-radius {
		cy = height - radius - 1
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= radius*radius
}

func rotateInto(destination *image.NRGBA, source *image.NRGBA, radians float64) {
	sine, cosine := math.Sin(radians), math.Cos(radians)
	sourceCenterX := float64(source.Bounds().Dx()-1) / 2
	sourceCenterY := float64(source.Bounds().Dy()-1) / 2
	rotatedWidth := math.Abs(float64(source.Bounds().Dx())*cosine) + math.Abs(float64(source.Bounds().Dy())*sine)
	rotatedHeight := math.Abs(float64(source.Bounds().Dx())*sine) + math.Abs(float64(source.Bounds().Dy())*cosine)
	// Anchor the rotated card to the lower-right corner of its slot. Matrix
	// headers use that corner to meet the top edge of their result column.
	destinationCenterX := float64(destination.Bounds().Dx()) - rotatedWidth/2
	destinationCenterY := float64(destination.Bounds().Dy()) - rotatedHeight/2
	for y := 0; y < destination.Bounds().Dy(); y++ {
		for x := 0; x < destination.Bounds().Dx(); x++ {
			dx := float64(x) - destinationCenterX
			dy := float64(y) - destinationCenterY
			sourceX := cosine*dx + sine*dy + sourceCenterX
			sourceY := -sine*dx + cosine*dy + sourceCenterY
			sx, sy := int(math.Round(sourceX)), int(math.Round(sourceY))
			if sx < 0 || sy < 0 || sx >= source.Bounds().Dx() || sy >= source.Bounds().Dy() {
				continue
			}
			destination.SetNRGBA(x, y, source.NRGBAAt(sx, sy))
		}
	}
}
