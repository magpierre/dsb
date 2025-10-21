// Copyright 2025 Magnus Pierre
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package windows

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeType represents the type of theme
type ThemeType string

const (
	ThemeTypeCustom       ThemeType = "custom"
	ThemeTypeShadcnSlate  ThemeType = "shadcn-slate"
	ThemeTypeShadcnStone  ThemeType = "shadcn-stone"
	ThemeTypeDefault      ThemeType = "default"
)

// CustomTheme defines a modern theme for the Delta Sharing Browser
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

// ShadcnSlateTheme defines a theme inspired by shadcn/ui design system
// Based on the neutral/slate color palette with clean, modern aesthetics
type ShadcnSlateTheme struct{}

var _ fyne.Theme = (*ShadcnSlateTheme)(nil)

// ShadcnStoneTheme defines a theme inspired by shadcn/ui design system
// Based on the stone color palette with warm, earthy tones
type ShadcnStoneTheme struct{}

var _ fyne.Theme = (*ShadcnStoneTheme)(nil)

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

// ShadcnSlateTheme implementation
// Color scheme based on shadcn/ui's neutral/slate palette
func (s ShadcnSlateTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		switch name {
		case theme.ColorNameBackground:
			// hsl(0 0% 100%) - Pure white
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameForeground:
			// hsl(222.2 84% 4.9%) - Very dark slate for text
			return color.NRGBA{R: 0x02, G: 0x08, B: 0x17, A: 0xff}
		case theme.ColorNameDisabledButton:
			// Muted background
			return color.NRGBA{R: 0xf1, G: 0xf5, B: 0xf9, A: 0xff}
		case theme.ColorNameButton:
			// Light muted background for secondary buttons
			return color.NRGBA{R: 0xf1, G: 0xf5, B: 0xf9, A: 0xff}
		case theme.ColorNamePrimary:
			// Primary - dark slate for selected/primary elements
			return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff}
		case theme.ColorNameHover:
			// Slightly lighter slate for hover
			return color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff}
		case theme.ColorNameFocus:
			// Darker slate for focus
			return color.NRGBA{R: 0x0a, G: 0x0f, B: 0x1e, A: 0xff}
		case theme.ColorNamePressed:
			// Darker for pressed state
			return color.NRGBA{R: 0x0f, G: 0x17, B: 0x2a, A: 0xff}
		case theme.ColorNameInputBackground:
			// Pure white for inputs
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameSelection:
			// hsl(215 25% 85%) - Light slate selection
			return color.NRGBA{R: 0xcb, G: 0xd5, B: 0xe1, A: 0xff}
		case theme.ColorNameDisabled:
			// Muted foreground - mid slate
			return color.NRGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff}
		case theme.ColorNamePlaceHolder:
			// Muted foreground
			return color.NRGBA{R: 0x64, G: 0x74, B: 0x8b, A: 0xff}
		case theme.ColorNameScrollBar:
			// Border color
			return color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff}
		case theme.ColorNameShadow:
			// Subtle shadow
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0a}
		case theme.ColorNameInputBorder:
			// hsl(214.3 31.8% 91.4%) - Light border
			return color.NRGBA{R: 0xe2, G: 0xe8, B: 0xf0, A: 0xff}
		case theme.ColorNameForegroundOnPrimary:
			// White text on dark primary
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		}
	} else { // Dark mode
		switch name {
		case theme.ColorNameBackground:
			// hsl(222.2 84% 4.9%) - Very dark slate
			return color.NRGBA{R: 0x02, G: 0x08, B: 0x17, A: 0xff}
		case theme.ColorNameForeground:
			// hsl(210 40% 98%) - Almost white
			return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
		case theme.ColorNameDisabledButton:
			// Muted dark
			return color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
		case theme.ColorNameButton:
			// Light foreground on dark
			return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
		case theme.ColorNamePrimary:
			// Light primary in dark mode
			return color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
		case theme.ColorNameHover:
			// hsl(217.2 32.6% 25%) - Lighter dark slate
			return color.NRGBA{R: 0x2d, G: 0x37, B: 0x48, A: 0xff}
		case theme.ColorNameFocus:
			// Brighter for focus
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameInputBackground:
			// hsl(217.2 32.6% 17.5%) - Dark slate input
			return color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
		case theme.ColorNameSelection:
			// hsl(217.2 32.6% 30%) - Mid slate selection
			return color.NRGBA{R: 0x33, G: 0x41, B: 0x55, A: 0xff}
		case theme.ColorNameDisabled:
			// hsl(215 20.2% 65.1%) - Light slate for disabled
			return color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}
		case theme.ColorNamePlaceHolder:
			// Muted foreground
			return color.NRGBA{R: 0x94, G: 0xa3, B: 0xb8, A: 0xff}
		case theme.ColorNameScrollBar:
			// Border - dark slate
			return color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
		case theme.ColorNameShadow:
			// Darker shadow
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40}
		case theme.ColorNameInputBorder:
			// hsl(217.2 32.6% 17.5%) - Dark border
			return color.NRGBA{R: 0x1e, G: 0x29, B: 0x3b, A: 0xff}
		case theme.ColorNameForegroundOnPrimary:
			// Dark text on light primary
			return color.NRGBA{R: 0x02, G: 0x08, B: 0x17, A: 0xff}
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (s ShadcnSlateTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (s ShadcnSlateTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (s ShadcnSlateTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 12 // Slightly more padding for cleaner look
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 10 // Thinner scrollbars
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInnerPadding:
		return 8
	}
	return theme.DefaultTheme().Size(name)
}

// ShadcnStoneTheme implementation
// Color scheme based on shadcn/ui's stone palette with warm, earthy tones
func (st ShadcnStoneTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		switch name {
		case theme.ColorNameBackground:
			// hsl(60 9.1% 97.8%) - stone-50: Very light warm gray
			return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xf9, A: 0xff}
		case theme.ColorNameForeground:
			// hsl(20 14.3% 4.1%) - stone-950: Very dark for text
			return color.NRGBA{R: 0x0c, G: 0x0a, B: 0x09, A: 0xff}
		case theme.ColorNameDisabledButton:
			// stone-100 - Muted background
			return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf4, A: 0xff}
		case theme.ColorNameButton:
			// stone-100 - Light background for secondary buttons
			return color.NRGBA{R: 0xf5, G: 0xf5, B: 0xf4, A: 0xff}
		case theme.ColorNamePrimary:
			// hsl(24 9.8% 10%) - stone-900: Dark for selected/primary elements
			return color.NRGBA{R: 0x1c, G: 0x19, B: 0x17, A: 0xff}
		case theme.ColorNameHover:
			// stone-200 - Light hover state
			return color.NRGBA{R: 0xe9, G: 0xe5, B: 0xe3, A: 0xff}
		case theme.ColorNameFocus:
			// stone-700 - Darker for focus
			return color.NRGBA{R: 0x44, G: 0x40, B: 0x3c, A: 0xff}
		case theme.ColorNamePressed:
			// stone-900 - Darker for pressed state
			return color.NRGBA{R: 0x1c, G: 0x19, B: 0x17, A: 0xff}
		case theme.ColorNameInputBackground:
			// Pure white for inputs
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameSelection:
			// hsl(24 5.7% 82.9%) - stone-300: Light selection
			return color.NRGBA{R: 0xd9, G: 0xd3, B: 0xce, A: 0xff}
		case theme.ColorNameDisabled:
			// stone-400 - Muted foreground
			return color.NRGBA{R: 0xa9, G: 0xa2, B: 0x9d, A: 0xff}
		case theme.ColorNamePlaceHolder:
			// stone-400 - Muted foreground
			return color.NRGBA{R: 0xa9, G: 0xa2, B: 0x9d, A: 0xff}
		case theme.ColorNameScrollBar:
			// stone-200 - Border color
			return color.NRGBA{R: 0xe9, G: 0xe5, B: 0xe3, A: 0xff}
		case theme.ColorNameShadow:
			// Subtle shadow
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0a}
		case theme.ColorNameInputBorder:
			// stone-200 - Light border
			return color.NRGBA{R: 0xe9, G: 0xe5, B: 0xe3, A: 0xff}
		case theme.ColorNameForegroundOnPrimary:
			// White text on dark primary
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		}
	} else { // Dark mode
		switch name {
		case theme.ColorNameBackground:
			// hsl(20 14.3% 4.1%) - stone-950: Very dark background
			return color.NRGBA{R: 0x0c, G: 0x0a, B: 0x09, A: 0xff}
		case theme.ColorNameForeground:
			// hsl(60 9.1% 97.8%) - stone-50: Almost white
			return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xf9, A: 0xff}
		case theme.ColorNameDisabledButton:
			// stone-900 - Muted dark
			return color.NRGBA{R: 0x1c, G: 0x19, B: 0x17, A: 0xff}
		case theme.ColorNameButton:
			// stone-50 - Light foreground on dark
			return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xf9, A: 0xff}
		case theme.ColorNamePrimary:
			// stone-50 - Light primary in dark mode
			return color.NRGBA{R: 0xfa, G: 0xfa, B: 0xf9, A: 0xff}
		case theme.ColorNameHover:
			// stone-700 - Lighter dark tone
			return color.NRGBA{R: 0x44, G: 0x40, B: 0x3c, A: 0xff}
		case theme.ColorNameFocus:
			// Brighter for focus
			return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
		case theme.ColorNameInputBackground:
			// hsl(24 9.8% 10%) - stone-900: Dark input
			return color.NRGBA{R: 0x1c, G: 0x19, B: 0x17, A: 0xff}
		case theme.ColorNameSelection:
			// hsl(12 6.5% 15.1%) - stone-800: Mid-dark selection
			return color.NRGBA{R: 0x29, G: 0x25, B: 0x24, A: 0xff}
		case theme.ColorNameDisabled:
			// stone-500 - Muted mid-tone
			return color.NRGBA{R: 0x78, G: 0x71, B: 0x6c, A: 0xff}
		case theme.ColorNamePlaceHolder:
			// stone-500 - Muted foreground
			return color.NRGBA{R: 0x78, G: 0x71, B: 0x6c, A: 0xff}
		case theme.ColorNameScrollBar:
			// stone-800 - Border
			return color.NRGBA{R: 0x29, G: 0x25, B: 0x24, A: 0xff}
		case theme.ColorNameShadow:
			// Darker shadow
			return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40}
		case theme.ColorNameInputBorder:
			// stone-800 - Dark border
			return color.NRGBA{R: 0x29, G: 0x25, B: 0x24, A: 0xff}
		case theme.ColorNameForegroundOnPrimary:
			// Dark text on light primary
			return color.NRGBA{R: 0x0c, G: 0x0a, B: 0x09, A: 0xff}
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (st ShadcnStoneTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (st ShadcnStoneTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (st ShadcnStoneTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 12 // Slightly more padding for cleaner look
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 10 // Thinner scrollbars
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInnerPadding:
		return 8
	}
	return theme.DefaultTheme().Size(name)
}

// ThemeManager handles theme preferences and switching
type ThemeManager struct {
	app         fyne.App
	currentType ThemeType
}

// NewThemeManager creates a new theme manager
func NewThemeManager(app fyne.App) *ThemeManager {
	tm := &ThemeManager{
		app:         app,
		currentType: ThemeTypeCustom,
	}

	// Load saved theme preference
	savedTheme := app.Preferences().StringWithFallback("theme", string(ThemeTypeCustom))
	tm.currentType = ThemeType(savedTheme)

	return tm
}

// GetCurrentTheme returns the current theme instance
func (tm *ThemeManager) GetCurrentTheme() fyne.Theme {
	switch tm.currentType {
	case ThemeTypeShadcnSlate:
		return &ShadcnSlateTheme{}
	case ThemeTypeShadcnStone:
		return &ShadcnStoneTheme{}
	case ThemeTypeDefault:
		return theme.DefaultTheme()
	default:
		return &CustomTheme{}
	}
}

// SetTheme changes the current theme and saves the preference
func (tm *ThemeManager) SetTheme(themeType ThemeType) {
	tm.currentType = themeType
	tm.app.Preferences().SetString("theme", string(themeType))
	tm.app.Settings().SetTheme(tm.GetCurrentTheme())
}

// GetCurrentType returns the current theme type
func (tm *ThemeManager) GetCurrentType() ThemeType {
	return tm.currentType
}

// GetThemeName returns a user-friendly name for the theme type
func GetThemeName(themeType ThemeType) string {
	switch themeType {
	case ThemeTypeShadcnSlate:
		return "shadcn - neutral"
	case ThemeTypeShadcnStone:
		return "shadcn - stone"
	case ThemeTypeDefault:
		return "Fyne Default"
	default:
		return "Original Custom"
	}
}
