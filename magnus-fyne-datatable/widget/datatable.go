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

package widget

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"

	"github.com/magpierre/fyne-datatable/datatable"
	sortengine "github.com/magpierre/fyne-datatable/internal/sort"
)

// DataTable is a widget that displays tabular data with sorting and filtering capabilities.
type DataTable struct {
	widget.BaseWidget

	model *datatable.TableModel

	// Callbacks
	headerClickHandler func(col int)
	cellSelectHandler  func(row, col int)

	// Internal state
	table          *widget.Table
	filterBar      *FilterBar
	statusBar      *StatusBar
	columnSelector *ColumnSelector
	settingsButton *widget.Button
	window         fyne.Window
	container      *fyne.Container
	selectedRow    int          // Currently selected row (-1 if none)
	selectedRows   map[int]bool // Multiple selected rows (row index -> selected)
	selectedCell   struct {     // Currently selected cell (for cell selection mode)
		row int // -1 if no cell selected
		col int // -1 if no cell selected
	}
	config Config
}

// NewDataTable creates a new DataTable widget with default configuration.
func NewDataTable(model *datatable.TableModel) *DataTable {
	config := DefaultConfig()
	return NewDataTableWithConfig(model, config)
}

// NewDataTableWithConfig creates a new DataTable widget with custom configuration.
func NewDataTableWithConfig(model *datatable.TableModel, config Config) *DataTable {
	dt := &DataTable{
		model:        model,
		config:       config,
		selectedRow:  -1,                 // No row selected initially
		selectedRows: make(map[int]bool), // Initialize multi-selection map
	}
	dt.selectedCell.row = -1 // No cell selected initially
	dt.selectedCell.col = -1

	dt.ExtendBaseWidget(dt)
	dt.setupDefaultSorting() // Set up default sorting behavior
	dt.buildTable(config)
	dt.buildLayout()

	return dt
}

// setupDefaultSorting configures the default header click sorting behavior.
func (dt *DataTable) setupDefaultSorting() {
	dt.headerClickHandler = func(col int) {
		// Cycle through sort states: None → Asc → Desc → None
		currentSort := dt.model.GetSortState()

		var newDirection datatable.SortDirection
		if currentSort.Column == col {
			// Same column - cycle through states
			switch currentSort.Direction {
			case datatable.SortNone:
				newDirection = datatable.SortAscending
			case datatable.SortAscending:
				newDirection = datatable.SortDescending
			case datatable.SortDescending:
				newDirection = datatable.SortNone
			}
		} else {
			// Different column - start with ascending
			newDirection = datatable.SortAscending
		}

		// Apply the sort
		if newDirection == datatable.SortNone {
			dt.ClearSort()
		} else {
			if err := dt.SortByColumn(col, newDirection); err != nil {
				// Sort error - could log or handle silently
				_ = err
			}
		}
	}
}

// buildTable constructs the underlying Fyne table widget.
func (dt *DataTable) buildTable(config Config) {
	dt.table = widget.NewTable(
		func() (int, int) {
			return dt.model.VisibleRowCount(), dt.model.VisibleColumnCount()
		},
		func() fyne.CanvasObject {
			label := ttwidget.NewLabel("")
			// Enable ellipsis truncation for text that's too long
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*ttwidget.Label)
			value, err := dt.model.VisibleCell(id.Row, id.Col)
			if err != nil {
				label.SetText("Error")
				label.SetToolTip("")
				return
			}

			text := value.Formatted
			label.SetText(text)

			// Always set tooltip to show full cell content
			label.SetToolTip(text)

			// Highlight entire row if in row selection mode and this row is selected
			if dt.config.SelectionMode == SelectionModeRow && (dt.selectedRow == id.Row || dt.selectedRows[id.Row]) {
				label.Importance = widget.HighImportance
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.Importance = widget.MediumImportance
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	// Configure selection mode
	// Always show row numbers for better data navigation
	dt.table.ShowHeaderColumn = true

	if config.SelectionMode == SelectionModeRow {
		// Row selection mode - select entire row with checkboxes
		// Set minimum size for header column buttons to make them more visible
		dt.table.SetColumnWidth(0, 120) // Make row number column wider for checkboxes
	} else {
		// Cell selection mode (default) - show simple row numbers
		dt.table.SetColumnWidth(0, 60) // Narrower column for simple row numbers
	}

	// Enable and configure column headers
	dt.table.ShowHeaderRow = true
	dt.table.CreateHeader = func() fyne.CanvasObject {
		// Create a button that can be used for both row numbers and column headers
		btn := widget.NewButton("", nil)
		btn.Importance = widget.MediumImportance // Medium importance for better centered text
		// Size will be set by UpdateHeader and AutoAdjustColumns
		return btn
	}
	dt.table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		// Handle row number buttons (header column)
		if id.Col == -1 {
			btn := cell.(*widget.Button)

			if config.SelectionMode == SelectionModeRow {
				// Row selection mode - show toggle button with row number
				rowIndex := id.Row

				// Update button text to show toggle state and row number
				if dt.selectedRows[rowIndex] {
					btn.SetText(fmt.Sprintf("☑ %d", id.Row+1)) // Checked with row number
				} else {
					btn.SetText(fmt.Sprintf("☐ %d", id.Row+1)) // Unchecked with row number
				}

				// Set proper sizing for row number buttons
				btn.Importance = widget.MediumImportance
				btn.Resize(fyne.NewSize(120, 35))

				// Set click handler for toggle functionality
				btn.OnTapped = func() {
					// Toggle selection state
					dt.selectedRows[rowIndex] = !dt.selectedRows[rowIndex]

					// Update single selection if this is the only selected row
					if dt.selectedRows[rowIndex] {
						dt.selectedRow = rowIndex
					} else {
						dt.selectedRow = -1
					}

					// Refresh to show updated highlighting
					dt.table.Refresh()
					dt.Refresh()

					// Notify the cell selection handler
					if dt.cellSelectHandler != nil {
						dt.cellSelectHandler(rowIndex, -1) // -1 indicates full row selection
					}
				}
			} else {
				// Cell selection mode - show simple row number
				btn.SetText(fmt.Sprintf("%d", id.Row+1))
				btn.Importance = widget.LowImportance
				btn.Resize(fyne.NewSize(50, 30))
				btn.OnTapped = nil // No action in cell selection mode
			}
			return
		}

		// Handle column headers
		btn := cell.(*widget.Button)

		// Use medium importance for better centered text appearance
		btn.Importance = widget.MediumImportance
		// Don't set fixed size here - let AutoAdjustColumns control width

		colName, err := dt.model.VisibleColumnName(id.Col)
		if err != nil {
			btn.SetText("Column " + string(rune(id.Col+'A')))
			return
		}

		// Add sort indicator if this column is sorted
		sortState := dt.model.GetSortState()
		headerText := colName
		if sortState.IsSorted() && sortState.Column == id.Col {
			if sortState.Direction == datatable.SortAscending {
				headerText += " ↑"
			} else if sortState.Direction == datatable.SortDescending {
				headerText += " ↓"
			}
		}
		btn.SetText(headerText)

		// Set click handler for this column
		colIndex := id.Col
		btn.OnTapped = func() {
			if dt.headerClickHandler != nil {
				dt.headerClickHandler(colIndex)
			}
		}
	}

	// Set selection handler based on mode
	if config.SelectionMode == SelectionModeRow {
		// Row selection mode - notify with full row
		dt.table.OnSelected = func(id widget.TableCellID) {
			// Toggle the row in multi-selection map
			dt.selectedRows[id.Row] = !dt.selectedRows[id.Row]

			// Update single selection to track the most recently clicked row
			if dt.selectedRows[id.Row] {
				dt.selectedRow = id.Row
			} else {
				// If deselected, set to -1 if no other rows selected
				hasSelection := false
				for _, selected := range dt.selectedRows {
					if selected {
						hasSelection = true
						break
					}
				}
				if !hasSelection {
					dt.selectedRow = -1
				}
			}

			dt.table.Refresh()
			dt.Refresh() // Also refresh the DataTable widget itself

			if dt.cellSelectHandler != nil {
				// Call handler with row and -1 to indicate full row selection
				dt.cellSelectHandler(id.Row, -1)
			}
		}
	} else {
		// Cell selection mode - notify with specific cell
		dt.table.OnSelected = func(id widget.TableCellID) {
			// Store the selected cell coordinates
			dt.selectedCell.row = id.Row
			dt.selectedCell.col = id.Col

			// Clear row selection in cell mode and refresh
			dt.selectedRow = -1
			dt.table.Refresh() // Ensure immediate visual update
			dt.Refresh()       // Also refresh the DataTable widget itself

			if dt.cellSelectHandler != nil {
				dt.cellSelectHandler(id.Row, id.Col)
			}
		}
	}

	// Set minimum column width if specified
	if config.MinColumnWidth > 0 {
		for i := 0; i < dt.model.VisibleColumnCount(); i++ {
			dt.table.SetColumnWidth(i, float32(config.MinColumnWidth))
		}
	}

	// Auto-adjust column widths if enabled
	if config.AutoAdjustColumnWidths {
		dt.AutoAdjustColumns()
	}

	// Enable focus for keyboard shortcuts
	// Note: Table widget doesn't have OnTapped, so we'll handle focus differently
}

// AutoAdjustColumns adjusts all column widths to fit their header text.
// This method can be called at any time to resize columns based on current headers.
func (dt *DataTable) AutoAdjustColumns() {
	if dt.table == nil || dt.model == nil {
		return
	}

	colCount := dt.model.VisibleColumnCount()

	// Create a temporary button to measure text size
	tempButton := widget.NewButton("", nil)
	tempButton.Importance = widget.LowImportance

	for col := 0; col < colCount; col++ {
		// Get column name
		colName, err := dt.model.VisibleColumnName(col)
		if err != nil {
			continue
		}

		// Add extra space for sort indicator (which could appear)
		headerText := colName + " ↓" // Account for widest indicator

		// Use medium importance button for accurate measurement
		tempButton.Importance = widget.MediumImportance
		tempButton.SetText(headerText)
		minSize := tempButton.MinSize()

		// Add generous padding for comfortable display and center alignment
		// Buttons need extra space on both sides for centered text to look good
		width := minSize.Width + 40 // Increased padding for better centering

		// Apply minimum width if configured
		if dt.config.MinColumnWidth > 0 && width < float32(dt.config.MinColumnWidth) {
			width = float32(dt.config.MinColumnWidth)
		}

		// Set the column width
		dt.table.SetColumnWidth(col, width)
	}

	// Refresh the table to apply changes
	dt.table.Refresh()
}

// buildLayout creates the layout with optional filter bar, column selector, and status bar.
func (dt *DataTable) buildLayout() {
	var top, bottom fyne.CanvasObject

	// Build top section with FilterBar and ColumnSelector if enabled
	topComponents := make([]fyne.CanvasObject, 0)

	if dt.config.ShowColumnSelector {
		dt.columnSelector = NewColumnSelector(dt)
		topComponents = append(topComponents, dt.columnSelector)
	}

	if dt.config.ShowFilterBar {
		dt.filterBar = NewFilterBar(dt)
		topComponents = append(topComponents, dt.filterBar)
	}

	if len(topComponents) > 0 {
		top = container.NewVBox(topComponents...)
	}

	// Create bottom section with StatusBar and Settings button if enabled
	bottomComponents := make([]fyne.CanvasObject, 0)

	if dt.config.ShowStatusBar {
		dt.statusBar = NewStatusBar(dt)
		bottomComponents = append(bottomComponents, dt.statusBar)
	}

	if dt.config.ShowSettingsButton {
		dt.settingsButton = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			if dt.window != nil {
				settingsDialog := NewSettingsDialog(dt, dt.window)
				settingsDialog.Show()
			}
		})
		dt.settingsButton.Importance = widget.LowImportance
		bottomComponents = append(bottomComponents, dt.settingsButton)
	}

	if len(bottomComponents) > 0 {
		bottom = container.NewBorder(nil, nil, nil, dt.settingsButton, dt.statusBar)
	}

	// Build container with border layout (no right component now)
	dt.container = container.NewBorder(top, bottom, nil, nil, dt.table)
}

// SetWindow sets the window reference for the DataTable.
// This is required for the settings dialog to work properly.
// It also registers keyboard shortcuts for copy operations.
func (dt *DataTable) SetWindow(window fyne.Window) {
	dt.window = window

	// Register keyboard shortcuts for CMD+C (Mac) / Ctrl+C (Windows/Linux)
	if window != nil {
		copyHandler := func(shortcut fyne.Shortcut) {
			// Use appropriate copy method based on selection mode
			if dt.config.SelectionMode == SelectionModeRow {
				_ = dt.CopySelectedRows()
			} else {
				_ = dt.CopySelectedCell()
			}
		}

		// Register Ctrl+C for Windows/Linux
		ctrlCShortcut := &desktop.CustomShortcut{
			KeyName:  fyne.KeyC,
			Modifier: fyne.KeyModifierControl,
		}
		window.Canvas().AddShortcut(ctrlCShortcut, copyHandler)

		// Register Super+C (CMD+C) for Mac
		cmdCShortcut := &desktop.CustomShortcut{
			KeyName:  fyne.KeyC,
			Modifier: fyne.KeyModifierSuper,
		}
		window.Canvas().AddShortcut(cmdCShortcut, copyHandler)
	}
}

// OnHeaderClick sets a callback for when a column header is clicked.
// This overrides the default sorting behavior. If you want custom behavior
// in addition to sorting, call the default sorting logic within your handler.
//
// Default behavior (if not overridden):
//   - Clicking a column header cycles through: None → Ascending → Descending → None
//   - Clicking a different column starts at Ascending
//   - Sort indicators (↑ ↓) are automatically updated in the headers
func (dt *DataTable) OnHeaderClick(handler func(col int)) {
	dt.headerClickHandler = handler

	// Update the table's selection handler to call header click for header rows
	if dt.table != nil {
		dt.table.OnSelected = func(id widget.TableCellID) {
			// Check if it's a header click (Fyne uses row -1 or special handling for headers)
			// Since headers are shown as a separate row, we need to check differently
			// For now, assume regular cell clicks call cellSelectHandler
			if dt.cellSelectHandler != nil {
				dt.cellSelectHandler(id.Row, id.Col)
			}
		}
	}
}

// OnCellSelected sets a callback for when a cell or row is selected.
// In cell selection mode (default): handler receives (row, col) for the selected cell.
// In row selection mode: handler receives (row, -1) indicating the entire row is selected.
//
// Example for cell mode:
//
//	table.OnCellSelected(func(row, col int) {
//	    cell, _ := model.VisibleCell(row, col)
//	    fmt.Printf("Cell [%d,%d] = %s\n", row, col, cell.Formatted)
//	})
//
// Example for row mode:
//
//	table.OnCellSelected(func(row, col int) {
//	    if col == -1 {
//	        // Full row selected
//	        rowData, _ := model.VisibleRow(row)
//	        fmt.Printf("Row %d selected\n", row)
//	    }
//	})
func (dt *DataTable) OnCellSelected(handler func(row, col int)) {
	dt.cellSelectHandler = handler
	// Note: The actual handler is set in buildTable() based on SelectionMode
}

// SortByColumn sorts the table by the specified column.
func (dt *DataTable) SortByColumn(col int, direction datatable.SortDirection) error {
	// Set sort state in model
	if err := dt.model.SetSort(col, direction); err != nil {
		return err
	}

	// Get the original column index for sorting
	visibleCols := dt.model.GetVisibleColumnIndices()
	if col >= len(visibleCols) {
		return datatable.ErrInvalidColumn
	}
	originalCol := visibleCols[col]

	// Get column type for sort
	colType, _ := dt.model.GetDataSource().ColumnType(originalCol)

	// Perform sort using sort engine
	engine := sortengine.NewEngine()
	sortedIndices, err := engine.Sort(
		dt.model.GetDataSource(),
		dt.model.GetVisibleRowIndices(),
		sortengine.SortSpec{
			Column:    originalCol,
			Direction: direction,
			DataType:  colType,
		},
	)
	if err != nil {
		return err
	}

	// Apply sorted indices to model
	if err := dt.model.ApplySortedIndices(sortedIndices); err != nil {
		return err
	}

	dt.Refresh()
	return nil
}

// ClearSort removes any active sorting.
func (dt *DataTable) ClearSort() error {
	if err := dt.model.ClearSort(); err != nil {
		return err
	}
	dt.Refresh()
	return nil
}

// SetFilter applies a filter to the table.
func (dt *DataTable) SetFilter(filter datatable.Filter) error {
	if err := dt.model.SetFilter(filter); err != nil {
		return err
	}
	dt.Refresh()
	return nil
}

// ClearFilter removes any active filter.
func (dt *DataTable) ClearFilter() error {
	if err := dt.model.SetFilter(nil); err != nil {
		return err
	}
	dt.Refresh()
	return nil
}

// Reconfigure updates the DataTable with a new configuration.
// This rebuilds the table UI with the new settings.
// Use this method when you need to change configuration after table creation.
func (dt *DataTable) Reconfigure(newConfig Config) {
	dt.config = newConfig

	// Clear selection when reconfiguring
	dt.selectedRow = -1
	dt.selectedRows = make(map[int]bool) // Clear multi-selection
	dt.selectedCell.row = -1              // Clear cell selection
	dt.selectedCell.col = -1

	// Save reference to old container that renderer is using
	oldContainer := dt.container

	// Rebuild the table with new configuration
	dt.buildTable(newConfig)

	// Build new layout (creates a new container)
	dt.buildLayout()

	// Copy new container contents to old container (which renderer has reference to)
	if oldContainer != nil {
		oldContainer.Objects = dt.container.Objects
		oldContainer.Layout = dt.container.Layout

		// Update container reference to use the original one
		dt.container = oldContainer
	}

	// Refresh the container and widget
	if dt.container != nil {
		dt.container.Refresh()
	}
	dt.Refresh()
}

// Refresh updates the table display.
func (dt *DataTable) Refresh() {
	if dt.table != nil {
		// Refresh the table which also updates headers with sort indicators
		dt.table.Refresh()
	}
	if dt.statusBar != nil {
		dt.statusBar.Update()
	}
	dt.BaseWidget.Refresh()
}

// CreateRenderer returns the widget's renderer.
func (dt *DataTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dt.container)
}

// SelectionMode defines how table cells/rows can be selected.
type SelectionMode int

const (
	// SelectionModeCell allows individual cell selection (default)
	SelectionModeCell SelectionMode = iota
	// SelectionModeRow allows full row selection
	SelectionModeRow
)

// Config holds configuration options for DataTable.
type Config struct {
	ShowFilterBar          bool
	ShowStatusBar          bool
	ShowColumnSelector     bool
	ShowSettingsButton     bool
	AutoAdjustColumnWidths bool
	SelectionMode          SelectionMode
	MinColumnWidth         int
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		ShowFilterBar:          true,
		ShowStatusBar:          true,
		ShowColumnSelector:     false,
		ShowSettingsButton:     true, // Show settings button by default
		AutoAdjustColumnWidths: false,
		SelectionMode:          SelectionModeRow, // Default to row selection
		MinColumnWidth:         100,
	}
}

// Note: Cell text truncation with ellipsis (...) is always enabled for text
// that exceeds column width. This is a built-in feature and cannot be disabled.
//
// Note: Tooltips are enabled on all cells to show full content on hover.
// To enable tooltip display, wrap the DataTable with the tooltip layer when adding to window:
//
//   table := dtwidget.NewDataTable(model)
//   content := dtwidget.WrapWithTooltips(table, window.Canvas())
//   window.SetContent(content)
//
// Or use the helper: window.SetContent(dtwidget.WrapWithTooltips(table, window.Canvas()))

// CopySelectedRows copies the selected rows to the clipboard as tab-separated values.
// This method handles both single and multi-row selection.
func (dt *DataTable) CopySelectedRows() error {
	if dt.config.SelectionMode != SelectionModeRow {
		return fmt.Errorf("copy is only available in row selection mode")
	}

	// Get selected rows - check multi-selection map first
	var selectedRowIndices []int

	// Collect all selected rows from the map
	for rowIndex, selected := range dt.selectedRows {
		if selected {
			selectedRowIndices = append(selectedRowIndices, rowIndex)
		}
	}

	// If no rows in the map, fall back to single selection
	if len(selectedRowIndices) == 0 && dt.selectedRow != -1 {
		selectedRowIndices = []int{dt.selectedRow}
	}

	if len(selectedRowIndices) == 0 {
		return fmt.Errorf("no rows selected")
	}

	// Sort the indices to maintain row order in the clipboard
	sort.Ints(selectedRowIndices)

	// Build the copied data
	var rows []string

	// Add header row
	var headerRow []string
	for col := 0; col < dt.model.VisibleColumnCount(); col++ {
		colName, err := dt.model.VisibleColumnName(col)
		if err != nil {
			colName = fmt.Sprintf("Column %d", col)
		}
		headerRow = append(headerRow, colName)
	}
	rows = append(rows, strings.Join(headerRow, "\t"))

	// Add data rows
	for _, rowIndex := range selectedRowIndices {
		var rowData []string
		for col := 0; col < dt.model.VisibleColumnCount(); col++ {
			cell, err := dt.model.VisibleCell(rowIndex, col)
			if err != nil {
				rowData = append(rowData, "Error")
			} else {
				rowData = append(rowData, cell.Formatted)
			}
		}
		rows = append(rows, strings.Join(rowData, "\t"))
	}

	// Join all rows with newlines
	copiedText := strings.Join(rows, "\n")

	// Copy to clipboard
	if dt.window != nil {
		dt.window.Clipboard().SetContent(copiedText)
	}

	return nil
}

// CopySelectedCell copies the selected cell to the clipboard.
// This method is used in cell selection mode to copy individual cells.
func (dt *DataTable) CopySelectedCell() error {
	if dt.config.SelectionMode != SelectionModeCell {
		return fmt.Errorf("copy cell is only available in cell selection mode")
	}

	// Check if a cell is selected
	if dt.selectedCell.row == -1 || dt.selectedCell.col == -1 {
		return fmt.Errorf("no cell selected")
	}

	// Get the cell value
	cell, err := dt.model.VisibleCell(dt.selectedCell.row, dt.selectedCell.col)
	if err != nil {
		return fmt.Errorf("error getting cell value: %w", err)
	}

	// Copy to clipboard - just the cell value, no header
	if dt.window != nil {
		dt.window.Clipboard().SetContent(cell.Formatted)
	}

	return nil
}

// TypedKey handles keyboard events for the DataTable.
// This enables keyboard shortcuts like Cmd+C for copying selected rows or cells.
func (dt *DataTable) TypedKey(event *fyne.KeyEvent) {
	// For now, just handle C key (we'll add modifier detection later)
	if event.Name == fyne.KeyC {
		// Use appropriate copy method based on selection mode
		if dt.config.SelectionMode == SelectionModeRow {
			_ = dt.CopySelectedRows()
		} else {
			_ = dt.CopySelectedCell()
		}
	}
}

// FocusGained is called when the DataTable receives focus.
// This enables keyboard shortcuts to work.
func (dt *DataTable) FocusGained() {
}

// FocusLost is called when the DataTable loses focus.
func (dt *DataTable) FocusLost() {
}

// Tapped is called when the DataTable is clicked.
// This ensures the widget receives focus for keyboard shortcuts.
func (dt *DataTable) Tapped(event *fyne.PointEvent) {
	// Request focus when tapped
	dt.RequestFocus()
}

// RequestFocus requests focus for the DataTable widget.
func (dt *DataTable) RequestFocus() {
	if dt.window != nil {
		dt.window.Canvas().Focus(dt)
	}
}

// CanFocus returns true to indicate this widget can receive focus.
func (dt *DataTable) CanFocus() bool {
	return true
}

// TypedRune handles rune input for the DataTable.
// This is required to implement the fyne.Focusable interface.
func (dt *DataTable) TypedRune(r rune) {
	// Handle rune input if needed
}

// WrapWithTooltips wraps a DataTable with the tooltip layer to enable tooltip display.
// This must be called when setting the window content to enable tooltips on cells.
//
// Usage:
//
//	table := NewDataTable(model)
//	content := WrapWithTooltips(table, window.Canvas())
//	window.SetContent(content)
func WrapWithTooltips(table *DataTable, canvas fyne.Canvas) fyne.CanvasObject {
	return fynetooltip.AddWindowToolTipLayer(table, canvas)
}
