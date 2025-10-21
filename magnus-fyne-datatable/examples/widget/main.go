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

package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/magpierre/fyne-datatable/adapters/slice"
	"github.com/magpierre/fyne-datatable/datatable"
	dtwidget "github.com/magpierre/fyne-datatable/widget"
)

func main() {
	// Create Fyne application
	myApp := app.New()
	window := myApp.NewWindow("DataTable Widget Demo")
	window.Resize(fyne.Size{Width: 800, Height: 600})

	// Create sample data
	data := [][]interface{}{
		{"Alice", 30, "Engineer", 75000.50, true},
		{"Bob", 25, "Designer", 65000.00, true},
		{"Charlie", 35, "Manager", 85000.75, true},
		{"Diana", 28, "Developer", 70000.25, true},
		{"Eve", 32, "Engineer", 76000.00, false},
		{"Frank", 27, "Designer", 66000.50, true},
		{"Grace", 31, "Developer", 72000.00, true},
		{"Henry", 29, "Engineer", 74000.25, false},
		{"Iris", 33, "Manager", 88000.00, true},
		{"Jack", 26, "Developer", 68000.75, true},
	}

	headers := []string{"Name", "Age", "Role", "Salary", "Active"}

	// Create data source
	source, err := slice.NewFromInterfaces(data, headers)
	if err != nil {
		log.Fatal("Failed to create data source:", err)
	}

	// Create table model
	model, err := datatable.NewTableModel(source)
	if err != nil {
		log.Fatal("Failed to create table model:", err)
	}

	// Create DataTable widget with all features enabled
	config := dtwidget.DefaultConfig()
	config.ShowColumnSelector = true                 // Enable column selection accordion
	config.ShowFilterBar = true                      // Enable row filtering (already default)
	config.ShowStatusBar = true                      // Enable status bar (already default)
	config.ShowSettingsButton = true                 // Enable settings button (already default)
	config.AutoAdjustColumnWidths = true             // Auto-adjust columns to fit headers
	config.SelectionMode = dtwidget.SelectionModeRow // Enable row selection for copy functionality
	table := dtwidget.NewDataTableWithConfig(model, config)

	// Set window reference for settings dialog
	table.SetWindow(window)

	// Optional: Add selection handler for logging (handles both cell and row modes)
	table.OnCellSelected(func(row, col int) {
		if col == -1 {
			// Row selection mode
			rowData, err := model.VisibleRow(row)
			if err != nil {
				log.Printf("Row selection error: %v\n", err)
				return
			}
			log.Printf("Row %d selected: %v\n", row, rowData)
		} else {
			// Cell selection mode
			cell, err := model.VisibleCell(row, col)
			if err != nil {
				log.Printf("Cell selection error: %v\n", err)
				return
			}
			colName, _ := model.VisibleColumnName(col)
			log.Printf("Cell selected: [%d, %d] (%s) = %s\n", row, col, colName, cell.Formatted)
		}
	})

	// Built-in features (no manual setup required):
	// ✓ Column sorting - Click headers to cycle through None → Asc → Desc
	// ✓ Column selection - Use accordion to show/hide columns
	// ✓ Row filtering - Use filter bar to query data
	// ✓ Status display - View row counts and sort/filter status
	// ✓ Auto-adjust columns - Headers are sized to fit their text
	// ✓ Ellipsis truncation - Long cell text shows "..." when truncated
	// ✓ Tooltips - Hover over cells to see full content

	// Wrap with tooltip layer to enable tooltips
	content := dtwidget.WrapWithTooltips(table, window.Canvas())

	// Keyboard shortcuts are now handled automatically by the DataTable widget
	// CMD+C (Mac) / Ctrl+C (Windows/Linux) is registered in SetWindow()
	// To test: select rows using the checkboxes and press CMD+C or Ctrl+C to copy

	window.SetContent(content)
	window.ShowAndRun()
}
