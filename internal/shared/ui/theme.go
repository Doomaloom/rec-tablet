package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

// Keep UI dialogs on Fyne's standard theme while custom lesson controls retain
// their own high-contrast colors and physical sizing.
type AppTheme struct{}

func (AppTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNamePrimary {
		return Blue
	}
	return theme.LightTheme().Color(n, theme.VariantLight)
}
func (AppTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (AppTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (AppTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }
