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
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"

	arrowadapter "github.com/magpierre/fyne-datatable/adapters/arrow"
	"github.com/magpierre/fyne-datatable/datatable"
	dtwidget "github.com/magpierre/fyne-datatable/widget"
)

// Data holds information about a table tab.
type Data struct {
	model      *datatable.TableModel
	dataTable  *dtwidget.DataTable
	tab        *container.TabItem
	tableName  string
	arrowTable arrow.Table // Keep reference for export
}

// DataBrowser manages the display of Delta Sharing table data.
type DataBrowser struct {
	w              fyne.Window
	Data           []Data
	innerTabs      *container.DocTabs
	docTabs        *container.DocTabs
	browserTab     *container.TabItem
	tabDataMap     map[*container.TabItem]*Data
	statusCallback func(string)
}

// CreateWindow initializes the data browser.
func (t *DataBrowser) CreateWindow(docTabs *container.DocTabs, statusCallback func(string)) {
	t.w = fyne.CurrentApp().Driver().AllWindows()[0]
	t.docTabs = docTabs
	t.Data = make([]Data, 0)
	t.tabDataMap = make(map[*container.TabItem]*Data)
	t.statusCallback = statusCallback

	// Create persistent inner tabs for individual tables
	t.innerTabs = container.NewDocTabs()
	t.innerTabs.SetTabLocation(container.TabLocationBottom)

	// Set up close intercept to clean up memory when tabs are closed
	t.innerTabs.CloseIntercept = func(ti *container.TabItem) {
		// Find and clean up the data associated with this tab
		if data, exists := t.tabDataMap[ti]; exists {
			// Release Arrow resources
			if data.arrowTable != nil {
				data.arrowTable.Release()
			}
			// Remove from map
			delete(t.tabDataMap, ti)
		}
		// Remove the tab
		t.innerTabs.Remove(ti)

		// Update status bar to reflect the currently selected tab (if any)
		if t.innerTabs.Selected() != nil {
			t.updateStatusForTab(t.innerTabs.Selected())
		} else {
			// No tabs left, clear status
			if t.statusCallback != nil {
				t.statusCallback("Ready")
			}
		}
	}

	// Set up tab selection callback to update status bar
	t.innerTabs.OnSelected = func(ti *container.TabItem) {
		t.updateStatusForTab(ti)
	}

	// Create persistent Browser tab
	t.browserTab = container.NewTabItem("Browser", t.innerTabs)
	t.docTabs.Append(t.browserTab)
}

// updateStatusForTab updates the status bar with information about the given tab.
func (t *DataBrowser) updateStatusForTab(ti *container.TabItem) {
	if ti == nil || t.statusCallback == nil {
		return
	}

	// Get the data associated with this tab
	if data, exists := t.tabDataMap[ti]; exists {
		model := data.model
		totalRows := model.OriginalRowCount()
		totalCols := model.OriginalColumnCount()
		visibleRows := model.VisibleRowCount()
		visibleCols := model.VisibleColumnCount()

		// Build status text showing total and filtered counts if applicable
		var statusText string
		if visibleRows != totalRows || visibleCols != totalCols {
			statusText = fmt.Sprintf("Table %s (showing %d/%d columns x %d/%d rows)",
				data.tableName, visibleCols, totalCols, visibleRows, totalRows)
		} else {
			statusText = fmt.Sprintf("Table %s (%d columns x %d rows)",
				data.tableName, totalCols, totalRows)
		}

		// Add filter/sort info
		sortState := model.GetSortState()
		if sortState.IsSorted() {
			colName, _ := model.VisibleColumnName(sortState.Column)
			direction := "↑"
			if sortState.Direction == datatable.SortDescending {
				direction = "↓"
			}
			statusText += fmt.Sprintf(" | Sorted: %s %s", colName, direction)
		}

		t.statusCallback(statusText)
	}
}

// CreateDataBrowser creates a new tab with the DataTable widget.
func (t *DataBrowser) CreateDataBrowser(
	arrowTable arrow.Table,
	delta_table delta_sharing.Table,
	statusCallback func(string),
) {
	// Create Arrow adapter
	source, err := arrowadapter.NewFromArrowTable(arrowTable)
	if err != nil {
		log.Printf("Failed to create Arrow adapter: %v", err)
		if statusCallback != nil {
			statusCallback(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	// Create model
	model, err := datatable.NewTableModel(source)
	if err != nil {
		log.Printf("Failed to create table model: %v", err)
		if statusCallback != nil {
			statusCallback(fmt.Sprintf("Error: %v", err))
		}
		return
	}

	// Create widget with configuration - all features enabled
	config := dtwidget.DefaultConfig()
	config.ShowFilterBar = true
	config.ShowStatusBar = true
	config.ShowColumnSelector = true                 // Enable built-in column selector
	config.ShowSettingsButton = true                 // Enable settings button
	config.AutoAdjustColumnWidths = true             // Auto-adjust columns to fit headers
	config.SelectionMode = dtwidget.SelectionModeRow // Enable row selection for copy functionality
	config.MinColumnWidth = 100

	dataTable := dtwidget.NewDataTableWithConfig(model, config)

	// Set window reference for settings dialog
	dataTable.SetWindow(t.w)

	// Sorting is now automatic! No handler needed.
	// Column selection is now automatic! No manual UI needed.

	// Optional: Setup selection handler for debugging (handles both cell and row modes)
	dataTable.OnCellSelected(func(row, col int) {
		if col == -1 {
			// Row selection mode
			rowData, err := model.VisibleRow(row)
			if err != nil {
				log.Printf("Row selection error: %v", err)
				return
			}
			log.Printf("Row %d selected: %v", row, rowData)
		} else {
			// Cell selection mode
			cell, err := model.VisibleCell(row, col)
			if err != nil {
				log.Printf("Cell selection error: %v", err)
				return
			}
			colName, _ := model.VisibleColumnName(col)
			log.Printf("Cell selected: [%d, %d] (%s) = %s", row, col, colName, cell.Formatted)
		}
	})

	// Wrap dataTable with tooltip layer to enable tooltips on cells
	content := dtwidget.WrapWithTooltips(dataTable, t.w.Canvas())

	// Keyboard shortcuts are now handled automatically by the DataTable widget
	// CMD+C (Mac) / Ctrl+C (Windows/Linux) is registered in SetWindow()
	// Plain C key is handled by the widget's TypedKey when it has focus

	// Create tab with the wrapped content
	tabName := delta_table.Name
	newTab := container.NewTabItem(tabName, content)

	// Store data
	data := &Data{
		model:      model,
		dataTable:  dataTable,
		tab:        newTab,
		tableName:  delta_table.Name,
		arrowTable: arrowTable, // Keep reference for export
	}

	// Retain Arrow table to prevent it from being released
	arrowTable.Retain()

	t.Data = append(t.Data, *data)
	t.tabDataMap[newTab] = data

	t.innerTabs.Append(newTab)
	t.innerTabs.Select(newTab)

	// Check if Browser tab still exists in docTabs, if not recreate it
	browserExists := false
	for _, item := range t.docTabs.Items {
		if item == t.browserTab {
			browserExists = true
			break
		}
	}

	if !browserExists {
		t.browserTab = container.NewTabItem("Browser", t.innerTabs)
		t.docTabs.Append(t.browserTab)
	}

	// Select the Browser tab
	t.docTabs.Select(t.browserTab)

	// Update status
	t.updateStatusForTab(newTab)
}

// GetData fetches data from Delta Sharing and creates a browser tab.
func (t *DataBrowser) GetData(profile string, table delta_sharing.Table, file_id string, options *QueryOptions) {
	c := make(chan bool)
	go func(c chan bool) {
		pbi := widget.NewProgressBarInfinite()
		di := dialog.NewCustomWithoutButtons(fmt.Sprintf("Loading %s...", table.Name), pbi, t.w)
		di.Resize(fyne.NewSize(300, 100))
		di.Show()
		pbi.Start()
		for {
			select {
			case <-c:
				di.Hide()
				pbi.Stop()
				return
			default:
				time.Sleep(time.Millisecond * 500)
			}
		}
	}(c)

	ds, err := delta_sharing.NewSharingClientV2FromString(profile)
	if err != nil {
		dialog.NewError(err, t.w).Show()
		c <- true
		return
	}

	resp, err := ds.ListFilesInTable(context.Background(), table)
	if err != nil {
		dialog.NewError(err, t.w).Show()
		c <- true
		return
	}

	for _, v := range resp.AddFiles {
		if v.Id == file_id {
			arrow_table, err := delta_sharing.LoadArrowTable(context.Background(), ds, table, file_id)
			if err != nil {
				dialog.NewError(err, t.w).Show()
				c <- true
				return
			}

			// Apply query options if provided
			if options != nil {
				arrow_table, err = t.applyQueryOptions(arrow_table, options)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to apply query options: %w", err), t.w)
					c <- true
					return
				}
			}

			// Use the new CreateDataBrowser
			t.CreateDataBrowser(arrow_table, table, t.statusCallback)

			c <- true
			t.w.Content().Refresh()
			return
		}
	}

	c <- true
}

// applyQueryOptions applies column selection and row limiting to the Arrow table.
// This is Delta Sharing-specific and kept from the original implementation.
func (d *DataBrowser) applyQueryOptions(table arrow.Table, options *QueryOptions) (arrow.Table, error) {
	if options == nil {
		return table, nil
	}

	// Apply column selection if specified
	if len(options.SelectedColumns) > 0 {
		// Build list of column indices to keep
		schema := table.Schema()
		colIndices := make([]int, 0)
		colNames := make(map[string]bool)

		for _, colName := range options.SelectedColumns {
			colNames[colName] = true
		}

		for i, field := range schema.Fields() {
			if colNames[field.Name] {
				colIndices = append(colIndices, i)
			}
		}

		if len(colIndices) == 0 {
			return nil, fmt.Errorf("no matching columns found")
		}

		// Create new schema with selected columns
		selectedFields := make([]arrow.Field, len(colIndices))
		for i, idx := range colIndices {
			selectedFields[i] = schema.Field(idx)
		}
		newSchema := arrow.NewSchema(selectedFields, nil)

		// Create new columns array
		columns := make([]arrow.Column, len(colIndices))
		for i, idx := range colIndices {
			col := table.Column(idx)
			columns[i] = *col
		}

		// Create new table with selected columns
		table = array.NewTable(newSchema, columns, table.NumRows())
	}

	// Apply row limit if specified
	if options.Limit > 0 && options.Limit < table.NumRows() {
		// Create a new table with limited rows
		numCols := int(table.NumCols())
		columns := make([]arrow.Column, numCols)
		for i := 0; i < numCols; i++ {
			col := table.Column(i)
			// Get the chunked array and slice it
			chunks := col.Data().Chunks()
			newChunks := make([]arrow.Array, 0)
			rowCount := int64(0)

			for _, chunk := range chunks {
				if rowCount >= options.Limit {
					break
				}
				remaining := options.Limit - rowCount
				if int64(chunk.Len()) <= remaining {
					newChunks = append(newChunks, chunk)
					rowCount += int64(chunk.Len())
				} else {
					// Slice the chunk
					sliced := array.NewSlice(chunk, 0, remaining)
					newChunks = append(newChunks, sliced)
					rowCount += remaining
				}
			}

			chunked := arrow.NewChunked(col.DataType(), newChunks)
			columns[i] = *arrow.NewColumn(col.Field(), chunked)
		}

		table = array.NewTable(table.Schema(), columns, options.Limit)
	}

	return table, nil
}

// exportData handles the export of data to different formats.
func (t *DataBrowser) exportData(dataItem *Data, format ExportFormat, tableName string) {
	// Determine file extension based on format
	var ext string
	switch format {
	case FormatParquet:
		ext = ".parquet"
	case FormatCSV:
		ext = ".csv"
	case FormatJSON:
		ext = ".json"
	}

	// Create file save dialog
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, t.w)
			return
		}
		if writer == nil {
			// User cancelled
			return
		}
		defer writer.Close()

		// Get the file path
		filePath := writer.URI().Path()

		// Create channel to control progress dialog
		c := make(chan bool)

		// Show progress indicator in a goroutine (following the GetData pattern)
		go func(c chan bool) {
			pbi := widget.NewProgressBarInfinite()
			progressDialog := dialog.NewCustomWithoutButtons("Exporting...", pbi, t.w)
			progressDialog.Resize(fyne.NewSize(300, 100))
			progressDialog.Show()
			pbi.Start()
			for {
				select {
				case <-c:
					progressDialog.Hide()
					pbi.Stop()
					return
				default:
					time.Sleep(time.Millisecond * 500)
				}
			}
		}(c)

		// Do export work on main callback thread (not in a separate goroutine)
		var exportErr error

		switch format {
		case FormatParquet:
			// Use existing Parquet export (Arrow-specific)
			// Create filtered Arrow table from current view
			filteredTable, convErr := t.createFilteredArrowTable(dataItem)
			if convErr != nil {
				exportErr = fmt.Errorf("failed to prepare filtered data: %w", convErr)
			} else {
				exportErr = ExportToParquet(filteredTable, filePath)
				filteredTable.Release()
			}

		case FormatCSV:
			// Use existing CSV export (Arrow-specific)
			filteredTable, convErr := t.createFilteredArrowTable(dataItem)
			if convErr != nil {
				exportErr = fmt.Errorf("failed to prepare filtered data: %w", convErr)
			} else {
				exportErr = ExportToCSV(filteredTable, filePath)
				filteredTable.Release()
			}

		case FormatJSON:
			// Use existing JSON export (Arrow-specific)
			filteredTable, convErr := t.createFilteredArrowTable(dataItem)
			if convErr != nil {
				exportErr = fmt.Errorf("failed to prepare filtered data: %w", convErr)
			} else {
				exportErr = ExportToJSON(filteredTable, filePath)
				filteredTable.Release()
			}
		}

		// Signal progress dialog to stop
		c <- true

		// Show result dialog on main thread
		if exportErr != nil {
			dialog.ShowError(fmt.Errorf("export failed: %w", exportErr), t.w)
		} else {
			dialog.ShowInformation("Export Successful",
				fmt.Sprintf("Data exported successfully to:\n%s", filePath), t.w)
		}
	}, t.w)

	// Set default filename
	defaultName := cleanFilename(tableName) + ext
	saveDialog.SetFileName(defaultName)

	saveDialog.Show()
}

// cleanFilename removes spaces and special characters from a filename.
func cleanFilename(name string) string {
	// Simple implementation - replace spaces with underscores
	result := ""
	for _, r := range name {
		if r == ' ' {
			result += "_"
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		}
	}
	return result
}

// createFilteredArrowTable creates an Arrow table from the filtered/sorted data.
// This reconstructs an Arrow table from the current view state.
func (t *DataBrowser) createFilteredArrowTable(dataItem *Data) (arrow.Table, error) {
	model := dataItem.model
	originalTable := dataItem.arrowTable

	// Get visible row and column indices
	visibleRows := model.GetVisibleRowIndices()
	visibleCols := model.GetVisibleColumnIndices()

	if len(visibleRows) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	// Get the original schema
	originalSchema := originalTable.Schema()

	// Build new schema with only visible columns
	newFields := make([]arrow.Field, len(visibleCols))
	for i, colIdx := range visibleCols {
		newFields[i] = originalSchema.Field(colIdx)
	}
	schema := arrow.NewSchema(newFields, nil)

	// Create memory pool
	pool := memory.NewGoAllocator()

	// Get table reader to access typed values
	tr := array.NewTableReader(originalTable, originalTable.NumRows())
	defer tr.Release()
	tr.Next()
	rec := tr.Record()

	// Build Arrow arrays for each column using the visible row indices
	columns := make([]arrow.Column, len(visibleCols))
	for i, colIdx := range visibleCols {
		field := originalSchema.Field(colIdx)

		// Create builder based on data type
		builder := array.NewBuilder(pool, field.Type)
		defer builder.Release()

		// Append values from the original Arrow column using visible indices
		for _, rowIdx := range visibleRows {
			col := rec.Column(colIdx)
			appendValueToBuilder(builder, col, rowIdx)
		}

		// Build the array
		arr := builder.NewArray()
		defer arr.Release()

		// Create chunked array
		chunked := arrow.NewChunked(field.Type, []arrow.Array{arr})
		columns[i] = *arrow.NewColumn(field, chunked)
	}

	// Create and return the table
	return array.NewTable(schema, columns, int64(len(visibleRows))), nil
}

// appendValueToBuilder appends a typed value from an Arrow array to a builder
func appendValueToBuilder(builder array.Builder, col arrow.Array, pos int) {
	if col.IsNull(pos) {
		builder.AppendNull()
		return
	}

	switch col.DataType().ID() {
	case arrow.STRING:
		b := builder.(*array.StringBuilder)
		s := col.(*array.String)
		b.Append(s.Value(pos))
	case arrow.BINARY:
		b := builder.(*array.BinaryBuilder)
		bin := col.(*array.Binary)
		b.Append(bin.Value(pos))
	case arrow.BOOL:
		b := builder.(*array.BooleanBuilder)
		bl := col.(*array.Boolean)
		b.Append(bl.Value(pos))
	case arrow.INT8:
		b := builder.(*array.Int8Builder)
		i8 := col.(*array.Int8)
		b.Append(i8.Value(pos))
	case arrow.INT16:
		b := builder.(*array.Int16Builder)
		i16 := col.(*array.Int16)
		b.Append(i16.Value(pos))
	case arrow.INT32:
		b := builder.(*array.Int32Builder)
		i32 := col.(*array.Int32)
		b.Append(i32.Value(pos))
	case arrow.INT64:
		b := builder.(*array.Int64Builder)
		i64 := col.(*array.Int64)
		b.Append(i64.Value(pos))
	case arrow.UINT8:
		b := builder.(*array.Uint8Builder)
		u8 := col.(*array.Uint8)
		b.Append(u8.Value(pos))
	case arrow.UINT16:
		b := builder.(*array.Uint16Builder)
		u16 := col.(*array.Uint16)
		b.Append(u16.Value(pos))
	case arrow.UINT32:
		b := builder.(*array.Uint32Builder)
		u32 := col.(*array.Uint32)
		b.Append(u32.Value(pos))
	case arrow.UINT64:
		b := builder.(*array.Uint64Builder)
		u64 := col.(*array.Uint64)
		b.Append(u64.Value(pos))
	case arrow.FLOAT16:
		b := builder.(*array.Float16Builder)
		f16 := col.(*array.Float16)
		b.Append(f16.Value(pos))
	case arrow.FLOAT32:
		b := builder.(*array.Float32Builder)
		f32 := col.(*array.Float32)
		b.Append(f32.Value(pos))
	case arrow.FLOAT64:
		b := builder.(*array.Float64Builder)
		f64 := col.(*array.Float64)
		b.Append(f64.Value(pos))
	case arrow.DATE32:
		b := builder.(*array.Date32Builder)
		d32 := col.(*array.Date32)
		b.Append(d32.Value(pos))
	case arrow.DATE64:
		b := builder.(*array.Date64Builder)
		d64 := col.(*array.Date64)
		b.Append(d64.Value(pos))
	case arrow.TIMESTAMP:
		b := builder.(*array.TimestampBuilder)
		ts := col.(*array.Timestamp)
		b.Append(ts.Value(pos))
	case arrow.DECIMAL128:
		b := builder.(*array.Decimal128Builder)
		d128 := col.(*array.Decimal128)
		b.Append(d128.Value(pos))
	case arrow.STRUCT:
		// For struct types, we need to handle nested builders
		b := builder.(*array.StructBuilder)
		s := col.(*array.Struct)
		b.Append(true) // Mark as valid
		// Copy field values
		for i := 0; i < s.NumField(); i++ {
			fieldBuilder := b.FieldBuilder(i)
			fieldCol := s.Field(i)
			appendValueToBuilder(fieldBuilder, fieldCol, pos)
		}
	case arrow.LIST:
		// For list types, handle nested values
		b := builder.(*array.ListBuilder)
		l := col.(*array.List)
		b.Append(true)
		valueBuilder := b.ValueBuilder()
		offsets := l.Offsets()
		start := int(offsets[pos])
		end := int(offsets[pos+1])
		values := l.ListValues()
		for i := start; i < end; i++ {
			appendValueToBuilder(valueBuilder, values, i)
		}
	default:
		// For unsupported types, append null
		builder.AppendNull()
	}
}
