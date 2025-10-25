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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

// QueryOptions holds the query configuration for table data loading.
//
// NOTE: These options are currently applied CLIENT-SIDE after data is fetched.
// Future enhancement: Push these to the Delta Sharing server via query parameters
// to reduce network data transfer (requires delta_sharing library API update).
type QueryOptions struct {
	SelectedColumns []string // Columns to include (empty = all columns)
	Predicate       string   // SQL WHERE clause for filtering (e.g., "age > 25 AND status = 'active'")
	Limit           int64    // Maximum rows to return (-1 = no limit)
}

// QueryOptionsDialog creates a dialog for configuring query options
type QueryOptionsDialog struct {
	dialog         dialog.Dialog
	window         fyne.Window
	schema         *delta_sharing.SparkSchema
	columnChecks   map[string]*widget.Check
	predicateEntry *widget.Entry
	limitEntry     *widget.Entry
	callback       func(*QueryOptions)
}

// NewQueryOptionsDialog creates a new query options dialog
func NewQueryOptionsDialog(w fyne.Window, schema *delta_sharing.SparkSchema, callback func(*QueryOptions)) *QueryOptionsDialog {
	qod := &QueryOptionsDialog{
		window:       w,
		schema:       schema,
		columnChecks: make(map[string]*widget.Check),
		callback:     callback,
	}
	qod.createDialog()
	return qod
}

func (qod *QueryOptionsDialog) createDialog() {
	// Column selection
	columnSelectLabel := widget.NewLabel("Select Columns:")
	columnSelectLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create checkboxes for each column
	columnCheckboxes := container.NewVBox()

	// Add "Select All" / "Deselect All" buttons
	selectAllBtn := widget.NewButton("Select All", func() {
		for _, check := range qod.columnChecks {
			check.SetChecked(true)
		}
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		for _, check := range qod.columnChecks {
			check.SetChecked(false)
		}
	})

	selectButtons := container.NewHBox(selectAllBtn, deselectAllBtn)

	if qod.schema != nil {
		for _, field := range qod.schema.Fields {
			// Format the type for display
			typeStr := fmt.Sprintf("%v", field.Type)
			check := widget.NewCheck(fmt.Sprintf("%s (%s)", field.Name, typeStr), nil)
			check.SetChecked(true) // Default to all columns selected
			qod.columnChecks[field.Name] = check
			columnCheckboxes.Add(check)
		}
	}

	columnScroll := container.NewVScroll(columnCheckboxes)
	columnScroll.SetMinSize(fyne.NewSize(400, 200))

	// Predicate input
	predicateLabel := widget.NewLabel("Filter Predicate (SQL WHERE clause):")
	predicateLabel.TextStyle = fyne.TextStyle{Bold: true}

	qod.predicateEntry = widget.NewMultiLineEntry()
	qod.predicateEntry.SetPlaceHolder("e.g., age > 25 AND status = 'active'")
	qod.predicateEntry.SetMinRowsVisible(3)

	predicateHelp := widget.NewLabel("Leave empty for no filtering. Use column names and standard SQL operators.")
	predicateHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Limit input
	limitLabel := widget.NewLabel("Row Limit:")
	limitLabel.TextStyle = fyne.TextStyle{Bold: true}

	qod.limitEntry = widget.NewEntry()
	qod.limitEntry.SetText("1000") // Default to 1000 rows
	qod.limitEntry.SetPlaceHolder("Leave empty for all rows, or enter a number (e.g., 1000)")

	limitHelp := widget.NewLabel("Maximum number of rows to return. Leave empty to return all rows.")
	limitHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Create form layout
	content := container.NewVBox(
		columnSelectLabel,
		selectButtons,
		columnScroll,
		widget.NewSeparator(),
		predicateLabel,
		qod.predicateEntry,
		predicateHelp,
		widget.NewSeparator(),
		limitLabel,
		qod.limitEntry,
		limitHelp,
	)

	// Create dialog with custom buttons
	qod.dialog = dialog.NewCustomConfirm(
		"Query Options",
		"Load Data",
		"Cancel",
		content,
		func(confirmed bool) {
			if confirmed {
				qod.handleConfirm()
			}
		},
		qod.window,
	)

	qod.dialog.Resize(fyne.NewSize(500, 600))
}

func (qod *QueryOptionsDialog) handleConfirm() {
	options := &QueryOptions{
		SelectedColumns: make([]string, 0),
	}

	// Collect selected columns
	for colName, check := range qod.columnChecks {
		if check.Checked {
			options.SelectedColumns = append(options.SelectedColumns, colName)
		}
	}

	// If no columns selected, show error
	if len(options.SelectedColumns) == 0 {
		dialog.ShowError(fmt.Errorf("please select at least one column"), qod.window)
		return
	}

	// Get predicate
	options.Predicate = strings.TrimSpace(qod.predicateEntry.Text)

	// Get limit
	limitText := strings.TrimSpace(qod.limitEntry.Text)
	if limitText != "" {
		limit, err := strconv.ParseInt(limitText, 10, 64)
		if err != nil || limit <= 0 {
			dialog.ShowError(fmt.Errorf("invalid limit: must be a positive number"), qod.window)
			return
		}
		options.Limit = limit
	} else {
		options.Limit = -1 // No limit
	}

	// Call the callback
	if qod.callback != nil {
		qod.callback(options)
	}
}

func (qod *QueryOptionsDialog) Show() {
	qod.dialog.Show()
}

// ShowQueryOptionsDialogWithSchema loads table schema and shows enhanced query options dialog
//
// NOTE: Query options (predicateHints, limitHint, column selection) are currently applied
// CLIENT-SIDE after data is fetched from the Delta Sharing server. This means all data
// matching the table is transferred over the network before filtering.
//
// TODO: Once the delta_sharing library exposes query pushdown parameters in its public API,
// update this to push predicates and limits to the server to reduce data transfer.
// The internal protocol already supports this (see protocol.data struct), but it's not
// currently exposed in ListFilesInTable or LoadArrowTable methods.
func ShowQueryOptionsDialogWithSchema(w fyne.Window, profile string, table delta_sharing.Table, callback func(*QueryOptions)) {
	// Create and show progress dialog on calling thread (which should be main/UI thread)
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	progressDialog := dialog.NewCustomWithoutButtons("Loading Schema", progressBar, w)
	progressDialog.Resize(fyne.NewSize(300, 100))
	progressDialog.Show()

	// Launch single background goroutine to load schema
	go func() {
		// Perform schema loading (network I/O, no UI operations)
		ds, err := delta_sharing.NewSharingClientV2FromString(profile)
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to create client: %w", err), w)
			return
		}

		// Use GetTableMetadata to fetch schema without loading actual data
		// This is much faster than loading a file just to read the schema
		metadata, err := ds.GetTableMetadata(context.Background(), table)
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to get table metadata: %w", err), w)
			return
		}

		// Extract Spark schema from metadata
		sparkSchema, err := metadata.GetSparkSchema()
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to parse schema: %w", err), w)
			return
		}

		// Close progress dialog
		progressBar.Stop()
		progressDialog.Hide()

		// Brief delay to allow dialog to fully close
		time.Sleep(100 * time.Millisecond)

		// Create and show query options dialog
		qod := NewQueryOptionsDialog(w, sparkSchema, callback)
		qod.Show()
	}()
}
