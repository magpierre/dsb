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
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/magpierre/fyne-datatable/datatable"
)

// ColumnSelector provides UI for selecting which columns are visible.
type ColumnSelector struct {
	widget.BaseWidget

	dataTable *DataTable
	model     *datatable.TableModel

	// UI components
	checkboxes  map[int]*widget.Check
	accordion   *widget.Accordion
	container   *fyne.Container
	columnNames []string
}

// NewColumnSelector creates a new column selector for the given DataTable.
func NewColumnSelector(dt *DataTable) *ColumnSelector {
	cs := &ColumnSelector{
		dataTable:  dt,
		model:      dt.model,
		checkboxes: make(map[int]*widget.Check),
	}

	cs.ExtendBaseWidget(cs)
	cs.buildUI()

	return cs
}

// buildUI constructs the column selector's UI.
func (cs *ColumnSelector) buildUI() {
	// Get all column names from the data source
	totalColumns := cs.model.OriginalColumnCount()
	cs.columnNames = make([]string, totalColumns)

	for i := 0; i < totalColumns; i++ {
		colName, err := cs.model.GetDataSource().ColumnName(i)
		if err != nil {
			colName = "Column " + string(rune(i+'A'))
		}
		cs.columnNames[i] = colName
	}

	// Create checkboxes container
	columnFilterContainer := container.NewVBox()

	// Create checkboxes for each column
	for i, colName := range cs.columnNames {
		idx := i // Capture for closure
		check := widget.NewCheck(colName, func(checked bool) {
			cs.applyColumnVisibility()
		})
		check.Checked = true // All columns visible by default
		cs.checkboxes[idx] = check
		columnFilterContainer.Add(check)
	}

	// Create scrollable container
	columnFilterScroll := container.NewVScroll(columnFilterContainer)
	columnFilterScroll.SetMinSize(fyne.NewSize(200, 150))

	// Create card
	columnFilterCard := widget.NewCard("", "Select Columns", columnFilterScroll)

	// Create Select All / Deselect All buttons
	selectAllBtn := widget.NewButton("Select All", func() {
		cs.SelectAll()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		cs.DeselectAll()
	})

	filterButtons := container.NewHBox(selectAllBtn, deselectAllBtn)

	// Create accordion
	cs.accordion = widget.NewAccordion(
		widget.NewAccordionItem("Column Filter",
			container.NewBorder(filterButtons, nil, nil, nil, columnFilterCard)),
	)

	cs.container = container.NewVBox(cs.accordion)
}

// applyColumnVisibility updates the model based on selected checkboxes.
func (cs *ColumnSelector) applyColumnVisibility() {
	// Build list of visible column indices
	visibleIndices := make([]int, 0)
	totalColumns := cs.model.OriginalColumnCount()

	for i := 0; i < totalColumns; i++ {
		if check, exists := cs.checkboxes[i]; exists && check.Checked {
			visibleIndices = append(visibleIndices, i)
		}
	}

	// Update model with visible columns (require at least one column)
	if len(visibleIndices) > 0 {
		err := cs.model.SetVisibleColumns(visibleIndices)
		if err == nil {
			cs.dataTable.Refresh()
		}
	}
}

// SelectAll checks all column checkboxes.
func (cs *ColumnSelector) SelectAll() {
	totalColumns := cs.model.OriginalColumnCount()
	for i := 0; i < totalColumns; i++ {
		if check, exists := cs.checkboxes[i]; exists {
			check.SetChecked(true)
		}
	}
	cs.applyColumnVisibility()
}

// DeselectAll unchecks all but the first column checkbox.
func (cs *ColumnSelector) DeselectAll() {
	totalColumns := cs.model.OriginalColumnCount()
	for i := 0; i < totalColumns; i++ {
		if check, exists := cs.checkboxes[i]; exists {
			// Keep first column checked to ensure at least one is visible
			check.SetChecked(i == 0)
		}
	}
	cs.applyColumnVisibility()
}

// GetVisibleColumns returns the indices of currently visible columns.
func (cs *ColumnSelector) GetVisibleColumns() []int {
	visibleIndices := make([]int, 0)
	totalColumns := cs.model.OriginalColumnCount()

	for i := 0; i < totalColumns; i++ {
		if check, exists := cs.checkboxes[i]; exists && check.Checked {
			visibleIndices = append(visibleIndices, i)
		}
	}

	return visibleIndices
}

// SetVisibleColumns updates the checkboxes to match the given column indices.
func (cs *ColumnSelector) SetVisibleColumns(indices []int) {
	// Create a map of visible indices
	visibleMap := make(map[int]bool)
	for _, idx := range indices {
		visibleMap[idx] = true
	}

	// Update checkboxes
	totalColumns := cs.model.OriginalColumnCount()
	for i := 0; i < totalColumns; i++ {
		if check, exists := cs.checkboxes[i]; exists {
			check.SetChecked(visibleMap[i])
		}
	}
}

// CreateRenderer returns the widget's renderer.
func (cs *ColumnSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(cs.container)
}
