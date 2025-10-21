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

	"github.com/magpierre/fyne-datatable/internal/filter"
)

// FilterBar provides UI for filtering table data.
type FilterBar struct {
	widget.BaseWidget

	dataTable *DataTable

	// UI components
	queryEntry  *widget.Entry
	applyButton *widget.Button
	clearButton *widget.Button
	container   *fyne.Container
}

// NewFilterBar creates a new filter bar for the given DataTable.
func NewFilterBar(dt *DataTable) *FilterBar {
	fb := &FilterBar{
		dataTable: dt,
	}

	fb.ExtendBaseWidget(fb)
	fb.buildUI()

	return fb
}

// buildUI constructs the filter bar's UI.
func (fb *FilterBar) buildUI() {
	// Create query entry
	fb.queryEntry = widget.NewEntry()
	fb.queryEntry.SetPlaceHolder("Enter filter query (e.g., age > 25 AND role = 'Engineer')")
	fb.queryEntry.OnSubmitted = func(query string) {
		fb.applyFilter()
	}

	// Create apply button
	fb.applyButton = widget.NewButton("Apply Filter", func() {
		fb.applyFilter()
	})

	// Create clear button
	fb.clearButton = widget.NewButton("Clear", func() {
		fb.queryEntry.SetText("")
		fb.dataTable.ClearFilter()
	})

	// Build container
	fb.container = container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Filter:"),
		container.NewHBox(fb.applyButton, fb.clearButton),
		fb.queryEntry,
	)
}

// applyFilter applies the current query as a filter.
func (fb *FilterBar) applyFilter() {
	query := fb.queryEntry.Text

	if query == "" {
		fb.dataTable.ClearFilter()
		return
	}

	// Create query filter
	queryFilter := &filter.QueryFilter{
		Query: query,
	}

	// Apply filter
	if err := fb.dataTable.SetFilter(queryFilter); err != nil {
		// Show error (in a real app, this would show a dialog)
		fb.queryEntry.SetPlaceHolder("Error: " + err.Error())
	}
}

// CreateRenderer returns the widget's renderer.
func (fb *FilterBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fb.container)
}

// SetQuery sets the filter query text.
func (fb *FilterBar) SetQuery(query string) {
	fb.queryEntry.SetText(query)
}

// GetQuery returns the current filter query text.
func (fb *FilterBar) GetQuery() string {
	return fb.queryEntry.Text
}
