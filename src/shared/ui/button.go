package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"rectablet/src/shared/components/actionbutton"
	"rectablet/src/shared/components/design"
	"sync"
	"time"
)

// Compatibility names for workflow status colors; visual tokens come from COB.
var (
	Navy                     = design.Secondary
	Blue                     = design.Primary
	Rail                     = design.Background
	Muted                    = design.Alpha(design.Secondary, 179)
	Green                    = color.NRGBA{R: 23, G: 103, B: 93, A: 255}
	Red                      = design.Danger
	Amber                    = design.Header
	White                    = design.White
	Background               = design.Background
	successCheckIcon         = fyne.NewStaticResource("success-check.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M4 12l5 5L20 6" fill="none" stroke="#17675d" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
	selectedSuccessCheckIcon = fyne.NewStaticResource("selected-success-check.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M4 12l5 5L20 6" fill="none" stroke="#fff" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/></svg>`))
)

type TouchButton struct {
	widget.BaseWidget
	Text                  string
	Fill, Ink             color.Color
	FontSize, ScaleFactor float32
	MinimumFont           float32
	// Compact reduces vertical padding; negative MaxLabelLines allows unlimited wrapping.
	Compact                               bool
	SingleLine                            bool
	MaxLabelLines                         int
	HoldDuration                          time.Duration
	ProgressSegments                      int
	SuccessChecks                         int
	ProgressValue, ProgressTotal          int
	Disabled, Selected, Outlined, focused bool
	hovered                               bool
	// Selectable uses the library card treatment instead of an action fill.
	Selectable    bool
	OnTap, OnHold func()
	CanActivate   func() bool
	Icon          fyne.Resource
	IconOnly      bool
	timer         *time.Timer
	mu            sync.Mutex
	held, pressed bool
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
func (b *TouchButton) MouseIn(*desktop.MouseEvent)    { b.hovered = true; b.Refresh() }
func (b *TouchButton) MouseMoved(*desktop.MouseEvent) {}
func (b *TouchButton) MouseOut()                      { b.hovered = false; b.Release(); b.Refresh() }
func (b *TouchButton) TouchDown(*mobile.TouchEvent)   { b.down() }
func (b *TouchButton) TouchUp(*mobile.TouchEvent)     { b.Release() }
func (b *TouchButton) TouchCancel(*mobile.TouchEvent) { b.Release() }
func (b *TouchButton) FocusGained()                   { b.focused = true; b.Refresh() }
func (b *TouchButton) FocusLost()                     { b.focused = false; b.Release(); b.Refresh() }
func (b *TouchButton) TypedRune(rune)                 {}
func (b *TouchButton) TypedKey(e *fyne.KeyEvent) {
	if design.HandleEscape(b, e) {
		return
	}
	if e.Name == fyne.KeyReturn || e.Name == fyne.KeySpace {
		b.Tapped(nil)
	}
}
func (b *TouchButton) CreateRenderer() fyne.WidgetRenderer {
	r := &buttonRenderer{b: b, bg: canvas.NewRectangle(b.Fill), label: NewText(b.Text, b.FontSize, b.Ink)}
	r.label.Bold = true
	r.label.Center = true
	r.objects = []fyne.CanvasObject{r.bg}
	for i := 0; i < 3; i++ {
		segment := canvas.NewRectangle(Green)
		segment.Hide()
		r.progressSegments[i] = segment
		r.objects = append(r.objects, segment)
	}
	for i := range r.successChecks {
		check := canvas.NewImageFromResource(successCheckIcon)
		check.FillMode = canvas.ImageFillContain
		check.Hide()
		r.successChecks[i] = check
		r.objects = append(r.objects, check)
	}
	r.objects = append(r.objects, r.label)
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
	label            *TextBlock
	icon             *canvas.Image
	progressSegments [3]*canvas.Rectangle
	successChecks    [3]*canvas.Image
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
	for _, check := range r.successChecks {
		check.Hide()
	}
	r.bg.Resize(size)
	ink := actionbutton.PaintSurface(r.bg, actionbutton.SurfaceStyle{
		Fill: b.Fill, Ink: b.Ink, Selected: b.Selected, Outlined: b.Outlined,
		Focused: b.focused, Hovered: b.hovered, Disabled: b.Disabled, Scale: s, Selectable: b.Selectable,
	})
	r.label.Center = !b.Selectable
	r.label.Color = ink
	r.label.FontSize = b.FontSize
	r.label.MinimumFont = b.FontSize
	if b.MinimumFont > 0 {
		r.label.MinimumFont = b.MinimumFont
	}
	r.label.Text = b.Text
	r.label.ScaleFactor = s
	x, pad := ButtonPadding*s, 4*s
	if b.Compact {
		pad = 2 * s
	}
	available := max(1, size.Width-2*x)
	checkCount := min(3, max(0, b.SuccessChecks))
	if checkCount > 0 {
		checkSize := min(22*s, max(1, size.Height-4*s))
		gap := 3 * s
		checksWidth := float32(checkCount)*checkSize + float32(checkCount-1)*gap
		available = max(1, available-checksWidth-ButtonPadding*s)
		for i := 0; i < checkCount; i++ {
			check := r.successChecks[i]
			check.Resource = successCheckIcon
			if b.Selected && !b.Disabled {
				check.Resource = selectedSuccessCheckIcon
			}
			check.Resize(fyne.NewSize(checkSize, checkSize))
			check.Move(fyne.NewPos(size.Width-x-checksWidth+float32(i)*(checkSize+gap), (size.Height-checkSize)/2))
			check.Show()
			check.Refresh()
		}
	}
	r.label.MaxLines = 2
	if b.MaxLabelLines != 0 {
		r.label.MaxLines = b.MaxLabelLines
	}
	r.label.Wrap = !b.SingleLine
	if b.SingleLine {
		r.label.MaxLines = 1
	}
	if r.icon != nil {
		if b.IconOnly {
			h := min(32*s, min(size.Width-2*ButtonPadding*s, size.Height-4*s))
			h = max(1, h)
			r.icon.Move(fyne.NewPos((size.Width-h)/2, (size.Height-h)/2))
			r.icon.Resize(fyne.NewSize(h, h))
			r.label.Hide()
			r.bg.Refresh()
			return
		}
		h := min(56*s, size.Height/3)
		r.icon.Move(fyne.NewPos((size.Width-h)/2, (size.Height-h)/2-16*s))
		r.icon.Resize(fyne.NewSize(h, h))
		r.label.Move(fyne.NewPos(x, (size.Height+h)/2))
		r.label.Resize(fyne.NewSize(available, 30*s))
		r.label.MaxLines = 1
	} else {
		labelHeight := size.Height - 2*pad
		labelTop := pad
		r.label.Move(fyne.NewPos(x, labelTop))
		r.label.Resize(fyne.NewSize(available, max(1, labelHeight)))
		segmentCount := min(3, max(0, b.ProgressSegments))
		segmentInset := 2 * s
		segmentHeight := max(1, size.Height-2*segmentInset)
		segmentY := segmentInset
		if b.Selectable {
			segmentHeight = min(6*s, segmentHeight)
			segmentY = size.Height - segmentInset - segmentHeight
		}
		if b.ProgressTotal > 0 {
			value := min(b.ProgressTotal, max(0, b.ProgressValue))
			width := size.Width * float32(value) / float32(b.ProgressTotal)
			if width > 0 {
				segment := r.progressSegments[0]
				segment.FillColor = Green
				segment.CornerRadius = min(design.ControlRadius*s, segmentHeight/2)
				segment.Move(fyne.NewPos(0, segmentY))
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
				segment.CornerRadius = min(design.ControlRadius*s, segmentHeight/2)
				segment.Move(fyne.NewPos(startX+float32(i)*(segmentWidth+segmentGap), segmentY))
				segment.Resize(fyne.NewSize(segmentWidth, segmentHeight))
				segment.Show()
			}
		}
	}
	r.bg.Refresh()
	r.label.Refresh()
}
func (r *buttonRenderer) Refresh() { r.Layout(r.b.Size()); canvas.Refresh(r.b) }
