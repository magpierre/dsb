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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StatusBar displays information about the current table state.
type StatusBar struct {
	widget.BaseWidget

	dataTable *DataTable

	// UI components
	rowCountLabel *widget.Label
	filterLabel   *widget.Label
	sortLabel     *widget.Label
	container     *fyne.Container
}

// NewStatusBar creates a new status bar for the given DataTable.
func NewStatusBar(dt *DataTable) *StatusBar {
	sb := &StatusBar{
		dataTable: dt,
	}

	sb.ExtendBaseWidget(sb)
	sb.buildUI()
	sb.Update()

	return sb
}

// buildUI constructs the status bar's UI.
func (sb *StatusBar) buildUI() {
	sb.rowCountLabel = widget.NewLabel("")
	sb.filterLabel = widget.NewLabel("")
	sb.sortLabel = widget.NewLabel("")

	sb.container = container.NewHBox(
		sb.rowCountLabel,
		widget.NewLabel("|"),
		sb.filterLabel,
		widget.NewLabel("|"),
		sb.sortLabel,
	)
}

// Update updates the status bar's display.
func (sb *StatusBar) Update() {
	// Update row count
	visibleRows := sb.dataTable.model.VisibleRowCount()
	totalRows := sb.dataTable.model.OriginalRowCount()

	if visibleRows == totalRows {
		sb.rowCountLabel.SetText(fmt.Sprintf("Rows: %d", totalRows))
	} else {
		sb.rowCountLabel.SetText(fmt.Sprintf("Rows: %d of %d", visibleRows, totalRows))
	}

	// Update filter status
	if sb.dataTable.model.IsFiltered() {
		filters := sb.dataTable.model.GetActiveFilters()
		if len(filters) > 0 {
			sb.filterLabel.SetText(fmt.Sprintf("Filtered: %s", filters[0].Description()))
		} else {
			sb.filterLabel.SetText("Filtered: (columns hidden)")
		}
	} else {
		sb.filterLabel.SetText("No filter")
	}

	// Update sort status
	sortState := sb.dataTable.model.GetSortState()
	if sortState.IsSorted() {
		colName, err := sb.dataTable.model.VisibleColumnName(sortState.Column)
		if err != nil {
			sb.sortLabel.SetText("Sorted")
		} else {
			direction := "↑"
			if sortState.Direction == 2 { // SortDescending
				direction = "↓"
			}
			sb.sortLabel.SetText(fmt.Sprintf("Sorted: %s %s", colName, direction))
		}
	} else {
		sb.sortLabel.SetText("No sort")
	}

	sb.Refresh()
}

// CreateRenderer returns the widget's renderer.
func (sb *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sb.container)
}
