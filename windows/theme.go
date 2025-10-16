package windows

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme defines a modern theme for the Delta Sharing Browser
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf5, A: 0xff} // Light gray background
		case theme.ColorNameButton:
			return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff} // Material Blue
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff} // Material Blue
		case theme.ColorNameHover:
			return color.NRGBA{R: 0x64, G: 0xb5, B: 0xf6, A: 0xff} // Lighter blue
		case theme.ColorNameFocus:
			return color.NRGBA{R: 0x19, G: 0x76, B: 0xd2, A: 0xff} // Darker blue
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xff} // Dark gray text
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // White input
		case theme.ColorNameSelection:
			return color.NRGBA{R: 0xbb, G: 0xde, B: 0xfb, A: 0xff} // Light blue selection
		case theme.ColorNameForegroundOnPrimary:
			return color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff} // Dark gray for toolbar icons
		}
	} else {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff} // Dark background
		case theme.ColorNameButton:
			return color.NRGBA{R: 0x42, G: 0xa5, B: 0xf5, A: 0xff} // Lighter blue for dark mode
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x42, G: 0xa5, B: 0xf5, A: 0xff}
		case theme.ColorNameHover:
			return color.NRGBA{R: 0x64, G: 0xb5, B: 0xf6, A: 0xff}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 0x90, G: 0xca, B: 0xf9, A: 0xff}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff} // Light text
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0x2d, G: 0x2d, B: 0x2d, A: 0xff}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 0x1e, G: 0x88, B: 0xe5, A: 0xff}
		case theme.ColorNameForegroundOnPrimary:
			return color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff} // Dark gray for toolbar icons
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameSeparatorThickness:
		return 1
	}
	return theme.DefaultTheme().Size(name)
}
