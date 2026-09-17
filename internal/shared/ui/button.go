package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"sync"
	"time"
)

var (
	Navy       = color.NRGBA{R: 18, G: 58, B: 115, A: 255}
	Blue       = color.NRGBA{R: 49, G: 93, B: 175, A: 255}
	Rail       = color.NRGBA{R: 225, G: 233, B: 244, A: 255}
	Muted      = color.NRGBA{R: 82, G: 104, B: 130, A: 255}
	Green      = color.NRGBA{R: 23, G: 103, B: 93, A: 255}
	Red        = color.NRGBA{R: 164, G: 54, B: 51, A: 255}
	Amber      = color.NRGBA{R: 129, G: 83, B: 38, A: 255}
	White      = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	Background = color.NRGBA{R: 243, G: 246, B: 251, A: 255}
)

type TouchButton struct {
	widget.BaseWidget
	Text, Detail          string
	Fill, Ink             color.Color
	FontSize, ScaleFactor float32
	MinimumFont           float32
	// Compact reduces vertical padding; negative MaxLabelLines allows unlimited wrapping.
	Compact                               bool
	MaxLabelLines                         int
	HoldDuration                          time.Duration
	ProgressSegments                      int
	ProgressValue, ProgressTotal          int
	Disabled, Selected, Outlined, focused bool
	OnTap, OnHold                         func()
	CanActivate                           func() bool
	Icon                                  fyne.Resource
	IconOnly                              bool
	timer                                 *time.Timer
	mu                                    sync.Mutex
	held, pressed                         bool
}

func NewButton(text string, fn func()) *TouchButton {
	b := &TouchButton{Text: text, Fill: Blue, Ink: White, FontSize: 22, ScaleFactor: 1, HoldDuration: 3 * time.Second, OnTap: fn}
	b.ExtendBaseWidget(b)
	return b
}
func (b *TouchButton) allowed() bool { return !b.Disabled && (b.CanActivate == nil || b.CanActivate()) }
func (b *TouchButton) SetScale(s float32) {
	if b.ScaleFactor != s {
		b.ScaleFactor = s
		b.Refresh()
	}
}
func (b *TouchButton) MinSize() fyne.Size { return fyne.NewSize(1, 1) }
func (b *TouchButton) Tapped(*fyne.PointEvent) {
	b.mu.Lock()
	held := b.held
	b.held = false
	b.mu.Unlock()
	if !held && b.allowed() && b.OnTap != nil {
		b.OnTap()
	}
}
func (b *TouchButton) down() {
	if !b.allowed() {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pressed = true
	b.held = false
	if b.timer != nil {
		b.timer.Stop()
	}
	if b.OnHold != nil {
		holdDuration := b.HoldDuration
		if holdDuration <= 0 {
			holdDuration = 3 * time.Second
		}
		b.timer = time.AfterFunc(holdDuration, func() {
			fyne.Do(func() {
				b.mu.Lock()
				fire := b.pressed
				b.held = fire
				b.mu.Unlock()
				if fire && b.allowed() {
					b.OnHold()
				}
			})
		})
	}
}
func (b *TouchButton) Release() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pressed = false
	if b.timer != nil {
		b.timer.Stop()
	}
}
func (b *TouchButton) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		b.down()
	}
}
func (b *TouchButton) MouseUp(*desktop.MouseEvent)    { b.Release() }
func (b *TouchButton) MouseIn(*desktop.MouseEvent)    {}
func (b *TouchButton) MouseMoved(*desktop.MouseEvent) {}
func (b *TouchButton) MouseOut()                      { b.Release() }
func (b *TouchButton) TouchDown(*mobile.TouchEvent)   { b.down() }
func (b *TouchButton) TouchUp(*mobile.TouchEvent)     { b.Release() }
func (b *TouchButton) TouchCancel(*mobile.TouchEvent) { b.Release() }
func (b *TouchButton) FocusGained()                   { b.focused = true; b.Refresh() }
func (b *TouchButton) FocusLost()                     { b.focused = false; b.Release(); b.Refresh() }
func (b *TouchButton) TypedRune(rune)                 {}
func (b *TouchButton) TypedKey(e *fyne.KeyEvent) {
	if e.Name == fyne.KeyReturn || e.Name == fyne.KeySpace {
		b.Tapped(nil)
	}
}
func (b *TouchButton) CreateRenderer() fyne.WidgetRenderer {
	r := &buttonRenderer{b: b, bg: canvas.NewRectangle(b.Fill), label: NewText(b.Text, b.FontSize, b.Ink), detail: NewText(b.Detail, 15, b.Ink)}
	r.label.Bold = true
	r.label.Center = true
	r.detail.Center = true
	r.detail.Wrap = false
	r.detail.MaxLines = 1
	r.objects = []fyne.CanvasObject{r.bg}
	for i := 0; i < 3; i++ {
		segment := canvas.NewRectangle(Green)
		segment.Hide()
		r.progressSegments[i] = segment
		r.objects = append(r.objects, segment)
	}
	r.objects = append(r.objects, r.label, r.detail)
	if b.Icon != nil {
		r.icon = canvas.NewImageFromResource(b.Icon)
		r.icon.FillMode = canvas.ImageFillContain
		r.objects = append(r.objects, r.icon)
	}
	return r
}

type buttonRenderer struct {
	b                *TouchButton
	bg               *canvas.Rectangle
	label, detail    *TextBlock
	icon             *canvas.Image
	progressSegments [3]*canvas.Rectangle
	objects          []fyne.CanvasObject
}

func (r *buttonRenderer) MinSize() fyne.Size           { return fyne.NewSize(1, 1) }
func (r *buttonRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *buttonRenderer) Destroy()                     { r.b.Release() }
func (r *buttonRenderer) Layout(size fyne.Size) {
	b := r.b
	s := b.ScaleFactor
	for _, segment := range r.progressSegments {
		segment.Hide()
	}
	r.bg.Resize(size)
	r.bg.CornerRadius = 18 * s
	r.bg.FillColor = b.Fill
	r.bg.StrokeWidth = 0
	if b.Selected || b.Outlined || b.focused {
		r.bg.StrokeWidth = 3 * s
		r.bg.StrokeColor = Navy
	}
	ink := b.Ink
	if b.Disabled {
		r.bg.FillColor = Rail
		ink = Muted
	}
	r.label.Color = ink
	r.detail.Color = ink
	r.label.FontSize = b.FontSize
	r.label.MinimumFont = b.FontSize
	if b.MinimumFont > 0 {
		r.label.MinimumFont = b.MinimumFont
	}
	r.label.Text = b.Text
	r.detail.Text = b.Detail
	r.label.ScaleFactor = s
	r.detail.ScaleFactor = s
	x, pad := 12*s, 12*s
	if b.Compact {
		pad = 2 * s
	}
	available := max(1, size.Width-2*x)
	r.label.MaxLines = 2
	if b.MaxLabelLines != 0 {
		r.label.MaxLines = b.MaxLabelLines
	}
	if r.icon != nil {
		if b.IconOnly {
			h := min(96*s, size.Height/2)
			r.icon.Move(fyne.NewPos((size.Width-48*s)/2, (size.Height-h)/2))
			r.icon.Resize(fyne.NewSize(48*s, h))
			r.label.Hide()
			r.detail.Hide()
			r.bg.Refresh()
			return
		}
		h := min(56*s, size.Height/3)
		r.icon.Move(fyne.NewPos((size.Width-h)/2, (size.Height-h)/2-16*s))
		r.icon.Resize(fyne.NewSize(h, h))
		r.label.Move(fyne.NewPos(x, (size.Height+h)/2))
		r.label.Resize(fyne.NewSize(available, 30*s))
		r.label.MaxLines = 1
		r.detail.Hide()
	} else {
		labelHeight := size.Height - 2*pad
		if b.Detail != "" {
			labelHeight -= 23 * s
			r.detail.Show()
			r.detail.Move(fyne.NewPos(x, size.Height-pad-23*s))
			r.detail.Resize(fyne.NewSize(available, 23*s))
		} else {
			r.detail.Hide()
		}
		r.label.Move(fyne.NewPos(x, pad))
		r.label.Resize(fyne.NewSize(available, max(1, labelHeight)))
		segmentCount := min(3, max(0, b.ProgressSegments))
		segmentInset := 2 * s
		segmentHeight := max(1, size.Height-2*segmentInset)
		if b.ProgressTotal > 0 {
			value := min(b.ProgressTotal, max(0, b.ProgressValue))
			width := size.Width * float32(value) / float32(b.ProgressTotal)
			if width > 0 {
				segment := r.progressSegments[0]
				segment.FillColor = Green
				segment.CornerRadius = 16 * s
				segment.Move(fyne.NewPos(0, segmentInset))
				segment.Resize(fyne.NewSize(width, segmentHeight))
				segment.Show()
			}
		} else {
			segmentGap := 2 * s
			segmentWidth := max(1, (size.Width-2*segmentInset-2*segmentGap)/3)
			totalWidth := 3*segmentWidth + 2*segmentGap
			startX := (size.Width - totalWidth) / 2
			for i := 0; i < segmentCount; i++ {
				segment := r.progressSegments[i]
				segment.FillColor = Green
				segment.CornerRadius = 16 * s
				segment.Move(fyne.NewPos(startX+float32(i)*(segmentWidth+segmentGap), segmentInset))
				segment.Resize(fyne.NewSize(segmentWidth, segmentHeight))
				segment.Show()
			}
		}
	}
	r.bg.Refresh()
	r.label.Refresh()
	r.detail.Refresh()
}
func (r *buttonRenderer) Refresh() { r.Layout(r.b.Size()); canvas.Refresh(r.b) }
