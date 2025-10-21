package windows

import (
	"fmt"
	"strings"

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
	placeholder     string
	maxLineNumWidth int // Width needed for line numbers
}

// NewSyntaxEditor creates a new syntax editor widget
func NewSyntaxEditor() *SyntaxEditor {
	se := &SyntaxEditor{
		enabled:  true,
		textGrid: widget.NewTextGrid(),
	}

	// Set initial empty content
	se.textGrid.SetText("")

	se.ExtendBaseWidget(se)

	return se
}

// SetText sets the text content and applies syntax highlighting
func (se *SyntaxEditor) SetText(text string) {
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
		rows[lineNum] = se.createStyledRow(lineNum+1, line, se.maxLineNumWidth)
	}

	// Update TextGrid with all rows
	se.textGrid.SetText(text) // Set raw text first
	for i, row := range rows {
		if i < len(se.textGrid.Rows) {
			se.textGrid.SetRow(i, row)
		}
	}

	se.textGrid.Refresh()

	// Call onChange callback
	if se.onChange != nil {
		se.onChange(text)
	}
}

// createStyledRow parses a line and creates a styled TextGrid row with line number
func (se *SyntaxEditor) createStyledRow(lineNum int, lineText string, maxLineNumWidth int) widget.TextGridRow {
	// Parse the line to get styled cells
	cells := ParseGoLine(lineText)

	// Format the line number with right padding
	lineNumStr := fmt.Sprintf("%*d", maxLineNumWidth, lineNum)
	separator := " │ "

	// Calculate total prefix length (line number + separator)
	prefixLen := len(lineNumStr) + len(separator)

	// Create TextGrid row with prefix + cells
	row := widget.TextGridRow{
		Cells: make([]widget.TextGridCell, prefixLen+len(cells)),
	}

	// Add line number cells (dimmed color)
	lineNumStyle := &widget.CustomTextGridStyle{
		FGColor: theme.DisabledColor(),
	}
	for i, ch := range lineNumStr {
		row.Cells[i] = widget.TextGridCell{
			Rune:  ch,
			Style: lineNumStyle,
		}
	}

	// Add separator cells
	separatorStyle := &widget.CustomTextGridStyle{
		FGColor: theme.DisabledColor(),
	}
	for i, ch := range separator {
		row.Cells[len(lineNumStr)+i] = widget.TextGridCell{
			Rune:  ch,
			Style: separatorStyle,
		}
	}

	// Add the styled code cells
	for col, styledCell := range cells {
		row.Cells[prefixLen+col] = widget.TextGridCell{
			Rune:  styledCell.Rune,
			Style: styledCell.Style,
		}
	}

	return row
}

// updateLine parses and styles a single line with line number
func (se *SyntaxEditor) updateLine(lineNum int, lineText string) {
	// Create the styled row with line number (lineNum is 0-indexed, display as 1-indexed)
	row := se.createStyledRow(lineNum+1, lineText, se.maxLineNumWidth)

	// Set the row in TextGrid
	if lineNum < len(se.textGrid.Rows) {
		se.textGrid.SetRow(lineNum, row)
	}
}

// GetText returns the current text content
func (se *SyntaxEditor) GetText() string {
	return se.textGrid.Text()
}

// Text returns the current text content (alias for GetText for compatibility)
func (se *SyntaxEditor) Text() string {
	return se.textGrid.Text()
}

// SetPlaceHolder sets placeholder text (shown when empty)
func (se *SyntaxEditor) SetPlaceHolder(text string) {
	se.placeholder = text
	// Note: TextGrid doesn't have native placeholder support
	// Could implement by showing gray text when empty
}

// SetOnChanged sets the callback for text changes
func (se *SyntaxEditor) SetOnChanged(callback func(string)) {
	se.onChange = callback
}

// SetHighlightingEnabled enables or disables syntax highlighting
func (se *SyntaxEditor) SetHighlightingEnabled(enabled bool) {
	se.enabled = enabled
	if !enabled {
		// Reload text without syntax highlighting
		text := se.textGrid.Text()
		se.textGrid.SetText(text)
		se.textGrid.Refresh()
	} else {
		// Re-apply syntax highlighting
		text := se.textGrid.Text()
		se.SetText(text)
	}
}

// GetTextGrid returns the underlying TextGrid widget
func (se *SyntaxEditor) GetTextGrid() *widget.TextGrid {
	return se.textGrid
}

// CreateRenderer implements fyne.Widget interface
func (se *SyntaxEditor) CreateRenderer() fyne.WidgetRenderer {
	// Return the TextGrid's renderer
	return widget.NewSimpleRenderer(se.textGrid)
}

// MinSize returns the minimum size of the widget
func (se *SyntaxEditor) MinSize() fyne.Size {
	return se.textGrid.MinSize()
}

// Resize sets the size of the widget
func (se *SyntaxEditor) Resize(size fyne.Size) {
	se.BaseWidget.Resize(size)
	se.textGrid.Resize(size)
}

// UpdateLineRange updates only a range of lines (for better performance)
func (se *SyntaxEditor) UpdateLineRange(startLine, endLine int, text string) {
	if !se.enabled {
		return
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lineNum := startLine + i
		if lineNum > endLine {
			break
		}
		se.updateLine(lineNum, line)
	}

	se.textGrid.Refresh()
}
