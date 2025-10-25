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
	"fmt"
	"image/color"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SyntaxEditor is a custom widget that provides syntax-highlighted code editing
type SyntaxEditor struct {
	widget.BaseWidget
	textGrid        *widget.TextGrid
	enabled         bool
	onChange        func(string)
	maxLineNumWidth int        // Width needed for line numbers
	highlightedLine int        // Currently highlighted line (1-indexed, 0 = none)
	mu              sync.Mutex // Protects textGrid from concurrent access
}

// NewSyntaxEditor creates a new syntax editor widget
func NewSyntaxEditor() *SyntaxEditor {
	se := &SyntaxEditor{
		enabled:  true,
		textGrid: widget.NewTextGrid(),
	}

	// Enable built-in line numbers on the TextGrid
	se.textGrid.ShowLineNumbers = true

	// Note: Don't call SetText("") on an empty TextGrid with ShowLineNumbers = true
	// as it can cause an index out of bounds panic in Fyne v2.7.0

	se.ExtendBaseWidget(se)

	return se
}

// SetText sets the text content and applies syntax highlighting
func (se *SyntaxEditor) SetText(text string) {
	se.mu.Lock()
	defer se.mu.Unlock()

	if !se.enabled {
		se.textGrid.SetText(text)
		return
	}

	// Split text into lines
	lines := strings.Split(text, "\n")

	// Calculate the width needed for line numbers (minimum 2 for single-digit line numbers)
	se.maxLineNumWidth = len(fmt.Sprintf("%d", len(lines)))
	if se.maxLineNumWidth < 2 {
		se.maxLineNumWidth = 2
	}

	// Build all rows first
	rows := make([]widget.TextGridRow, len(lines))
	for lineNum, line := range lines {
		isHighlighted := (lineNum + 1) == se.highlightedLine
		rows[lineNum] = se.createStyledRow(lineNum+1, line, se.maxLineNumWidth, isHighlighted)
	}

	// Update TextGrid with all rows
	// Set the rows directly instead of calling SetText first to avoid cell duplication
	se.textGrid.Rows = rows

	se.textGrid.Refresh()

	// Call onChange callback
	if se.onChange != nil {
		se.onChange(text)
	}
}

// createStyledRow parses a line and creates a styled TextGrid row without line number (TextGrid handles that)
func (se *SyntaxEditor) createStyledRow(lineNum int, lineText string, maxLineNumWidth int, isHighlighted bool) widget.TextGridRow {
	// Parse the line to get styled cells
	cells := ParseGoLine(lineText)

	// Create TextGrid row with just the code cells (no manual line numbers)
	row := widget.TextGridRow{
		Cells: make([]widget.TextGridCell, len(cells)),
	}

	// Determine background color based on highlighting
	var bgColor color.Color
	if isHighlighted {
		bgColor = theme.SelectionColor()
	}

	// Add the styled code cells with highlighted background if needed
	for col, styledCell := range cells {
		// If highlighted, modify the style to include background color
		style := styledCell.Style
		if isHighlighted {
			if customStyle, ok := style.(*widget.CustomTextGridStyle); ok {
				// Create a new style with background color
				highlightedStyle := &widget.CustomTextGridStyle{
					FGColor: customStyle.FGColor,
					BGColor: bgColor,
				}
				style = highlightedStyle
			} else {
				// Create new custom style with background
				style = &widget.CustomTextGridStyle{
					BGColor: bgColor,
				}
			}
		}

		row.Cells[col] = widget.TextGridCell{
			Rune:  styledCell.Rune,
			Style: style,
		}
	}

	return row
}

// GetText returns the current text content
func (se *SyntaxEditor) GetText() string {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.textGrid.Text()
}

// Text returns the current text content (alias for GetText for compatibility)
func (se *SyntaxEditor) Text() string {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.textGrid.Text()
}

// SetOnChanged sets the callback for text changes
func (se *SyntaxEditor) SetOnChanged(callback func(string)) {
	se.onChange = callback
}

// SetHighlightingEnabled enables or disables syntax highlighting
func (se *SyntaxEditor) SetHighlightingEnabled(enabled bool) {
	se.mu.Lock()
	defer se.mu.Unlock()

	se.enabled = enabled
	if !enabled {
		// Reload text without syntax highlighting
		text := se.textGrid.Text()
		se.textGrid.SetText(text)
		se.textGrid.Refresh()
	} else {
		// Re-apply syntax highlighting
		text := se.textGrid.Text()
		// Need to unlock before calling SetText to avoid deadlock
		se.mu.Unlock()
		se.SetText(text)
		se.mu.Lock()
	}
}

// GetTextGrid returns the underlying TextGrid widget
func (se *SyntaxEditor) GetTextGrid() *widget.TextGrid {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.textGrid
}

// CreateRenderer implements fyne.Widget interface
func (se *SyntaxEditor) CreateRenderer() fyne.WidgetRenderer {
	// Return the TextGrid's renderer
	return widget.NewSimpleRenderer(se.textGrid)
}

// MinSize returns the minimum size of the widget
func (se *SyntaxEditor) MinSize() fyne.Size {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.textGrid.MinSize()
}

// Resize sets the size of the widget
func (se *SyntaxEditor) Resize(size fyne.Size) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.BaseWidget.Resize(size)
	se.textGrid.Resize(size)
}

// UpdateLineRange updates only a range of lines (for better performance)
func (se *SyntaxEditor) UpdateLineRange(startLine, endLine int, text string) {
	se.mu.Lock()
	defer se.mu.Unlock()

	if !se.enabled {
		return
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNum := startLine + i
		if lineNum > endLine {
			break
		}
		// Create the styled row with line number (lineNum is 0-indexed, display as 1-indexed)
		isHighlighted := (lineNum + 1) == se.highlightedLine
		row := se.createStyledRow(lineNum+1, line, se.maxLineNumWidth, isHighlighted)
		// Set the row in TextGrid
		if lineNum < len(se.textGrid.Rows) {
			se.textGrid.SetRow(lineNum, row)
		}
	}

	se.textGrid.Refresh()
}

// SetHighlightedLine sets the line to highlight (1-indexed, 0 to clear)
func (se *SyntaxEditor) SetHighlightedLine(lineNum int) {
	se.mu.Lock()
	defer se.mu.Unlock()

	oldLine := se.highlightedLine
	if oldLine == lineNum {
		return // No change needed
	}

	se.highlightedLine = lineNum

	// Only update the affected lines instead of rebuilding everything
	lines := strings.Split(se.textGrid.Text(), "\n")

	// Update old highlighted line (remove highlight)
	if oldLine > 0 && oldLine-1 < len(lines) {
		row := se.createStyledRow(oldLine, lines[oldLine-1], se.maxLineNumWidth, false)
		if oldLine-1 < len(se.textGrid.Rows) {
			se.textGrid.SetRow(oldLine-1, row)
		}
	}

	// Update new highlighted line (add highlight)
	if lineNum > 0 && lineNum-1 < len(lines) {
		row := se.createStyledRow(lineNum, lines[lineNum-1], se.maxLineNumWidth, true)
		if lineNum-1 < len(se.textGrid.Rows) {
			se.textGrid.SetRow(lineNum-1, row)
		}
	}

	se.textGrid.Refresh()
}
