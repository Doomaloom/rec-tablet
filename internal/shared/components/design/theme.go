// Package design defines the COB Aquatics palette and locally scoped Fyne theme.
package design

import (
	"embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

//go:embed fonts/*
var fonts embed.FS
var (
	Primary     = color.NRGBA{0, 122, 204, 255}
	Secondary   = color.NRGBA{0, 95, 153, 255}
	White       = color.NRGBA{255, 255, 255, 255}
	Text        = color.NRGBA{51, 51, 51, 255}
	Background  = color.NRGBA{230, 247, 255, 255}
	Header      = color.NRGBA{196, 83, 0, 255}
	Hover       = color.NRGBA{102, 217, 239, 255}
	Danger      = color.NRGBA{214, 69, 69, 255}
	DangerHover = color.NRGBA{192, 54, 54, 255}
	regular     = resource("Rubik-Regular.ttf")
	semibold    = resource("Rubik-SemiBold.ttf")
	italic      = resource("Rubik-RegularItalic.ttf")
	boldItalic  = resource("Rubik-SemiBoldItalic.ttf")
)

const (
	ControlRadius float32 = 16
	CardRadius    float32 = 24
	Target        float32 = 44
	Gap           float32 = 8
)

func resource(name string) fyne.Resource {
	data, err := fonts.ReadFile("fonts/" + name)
	if err != nil {
		panic(err)
	}
	return fyne.NewStaticResource(name, data)
}
func Alpha(c color.Color, a uint8) color.Color {
	value := color.NRGBAModel.Convert(c).(color.NRGBA)
	value.A = a
	return value
}

// Opacity fades an existing color without making transparent surfaces opaque.
func Opacity(c color.Color, opacity uint8) color.Color {
	value := color.NRGBAModel.Convert(c).(color.NRGBA)
	value.A = uint8(uint16(value.A) * uint16(opacity) / 255)
	return value
}

// Theme can override foreground, field fill, and text size for one subtree.
type Theme struct {
	Foreground, Input color.Color
	TextSize          float32
}

func (t Theme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNamePrimary, theme.ColorNameSelection:
		return Primary
	case theme.ColorNameBackground:
		return Background
	case theme.ColorNameForeground:
		if t.Foreground != nil {
			return t.Foreground
		}
		return Secondary
	case theme.ColorNameInputBackground:
		if t.Input != nil {
			return t.Input
		}
		return Background
	case theme.ColorNameInputBorder:
		return Secondary
	case theme.ColorNameButton:
		return Background
	case theme.ColorNameDisabledButton:
		return Alpha(Secondary, 25)
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return Alpha(Secondary, 153)
	case theme.ColorNameFocus:
		return Alpha(Primary, 75)
	case theme.ColorNameHover:
		return Alpha(Hover, 90)
	case theme.ColorNameError:
		return Danger
	case theme.ColorNameOverlayBackground:
		return White
	}
	return theme.LightTheme().Color(n, theme.VariantLight)
}
func (Theme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTheme().Font(s)
	}
	if s.Bold && s.Italic {
		return boldItalic
	}
	if s.Bold {
		return semibold
	}
	if s.Italic {
		return italic
	}
	return regular
}
func (Theme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (t Theme) Size(n fyne.ThemeSizeName) float32 {
	switch n {
	case theme.SizeNameText:
		if t.TextSize > 0 {
			return t.TextSize
		}
		return 14
	case theme.SizeNamePadding:
		return Gap
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameInputRadius, theme.SizeNameButtonRadius:
		return ControlRadius
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameCardRadius, theme.SizeNamePopupRadius, theme.SizeNameDialogRadius:
		return CardRadius
	}
	return theme.DefaultTheme().Size(n)
}
func Scope(obj fyne.CanvasObject) *container.ThemeOverride {
	return container.NewThemeOverride(obj, Theme{})
}
