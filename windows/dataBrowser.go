package windows

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

type Data struct {
	data              [][]string
	header            []string
	arrow_table       arrow.Table
	arrow_rec         arrow.Record
	tab               *container.TabItem
	tableName         string
	isFiltered        bool
	filteredData      [][]string
	filteredHeader    []string
	visibleColumns    []int
	filteredRowIndices []int // Maps filtered row index to original row index
}

type DataBrowser struct {
	w            fyne.Window
	Data         []Data
	innerTabs    *container.DocTabs
	docTabs      *container.DocTabs
	browserTab   *container.TabItem
	tabDataMap   map[*container.TabItem]*Data
}

func (t *DataBrowser) CreateWindow(docTabs *container.DocTabs) {
	t.w = fyne.CurrentApp().Driver().AllWindows()[0]
	t.docTabs = docTabs
	t.Data = make([]Data, 0)
	t.tabDataMap = make(map[*container.TabItem]*Data)

	// Create persistent inner tabs for individual tables
	t.innerTabs = container.NewDocTabs()
	t.innerTabs.SetTabLocation(container.TabLocationBottom)

	// Set up close intercept to clean up memory when tabs are closed
	t.innerTabs.CloseIntercept = func(ti *container.TabItem) {
		// Find and clean up the data associated with this tab
		if data, exists := t.tabDataMap[ti]; exists {
			// Release Arrow resources if they haven't been released yet
			if data.arrow_rec != nil {
				data.arrow_rec.Release()
			}
			if data.arrow_table != nil {
				data.arrow_table.Release()
			}
			// Clear the data arrays to help GC
			data.data = nil
			data.header = nil
			// Remove from map
			delete(t.tabDataMap, ti)
		}
		// Remove the tab
		t.innerTabs.Remove(ti)
	}

	// Create persistent Browser tab
	t.browserTab = container.NewTabItem("Browser", t.innerTabs)
	t.docTabs.Append(t.browserTab)
}

func (t *DataBrowser) CreateDataBrowser(dataItem *Data, delta_table delta_sharing.Table) {
	// Initialize filtered data with full dataset
	dataItem.filteredData = dataItem.data
	dataItem.filteredHeader = dataItem.header
	dataItem.visibleColumns = make([]int, len(dataItem.header))
	for i := range dataItem.visibleColumns {
		dataItem.visibleColumns[i] = i
	}
	// Initialize row indices to match all rows
	dataItem.filteredRowIndices = make([]int, len(dataItem.data))
	for i := range dataItem.filteredRowIndices {
		dataItem.filteredRowIndices[i] = i
	}

	// Create the table widget
	table := widget.NewTableWithHeaders(func() (rows int, cols int) {
		if len(dataItem.filteredData) == 0 {
			return 0, len(dataItem.filteredHeader)
		}
		return len(dataItem.filteredData), len(dataItem.filteredHeader)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("template.............")
	}, func(tci widget.TableCellID, co fyne.CanvasObject) {
		if tci.Row < len(dataItem.filteredData) && tci.Col < len(dataItem.filteredHeader) {
			co.(*widget.Label).SetText(dataItem.filteredData[tci.Row][tci.Col])
			co.(*widget.Label).Truncation = fyne.TextTruncateClip
		}
	})

	table.ShowHeaderColumn = false
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		if id.Col < len(dataItem.filteredHeader) {
			template.(*widget.Label).SetText(dataItem.filteredHeader[id.Col])
			template.(*widget.Label).Truncation = fyne.TextTruncateClip
		}
	}

	// Calculate and set column widths based on header text
	for i, headerText := range dataItem.header {
		// Measure the header text width and add padding
		textSize := fyne.MeasureText(headerText, theme.TextSize(), fyne.TextStyle{})
		columnWidth := textSize.Width + theme.Padding()*4 // Add padding for better spacing

		// Ensure a minimum width
		if columnWidth < 80 {
			columnWidth = 80
		}

		table.SetColumnWidth(i, columnWidth)
	}

	// Create search/filter controls
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search: text or column=value, column>10, name~'John' AND age>=18 OR status='active'")

	// Create column filter UI
	columnChecks := make(map[string]*widget.Check)
	columnFilterContainer := container.NewVBox()

	for _, colName := range dataItem.header {
		check := widget.NewCheck(colName, nil)
		check.Checked = true
		columnChecks[colName] = check
		columnFilterContainer.Add(check)
	}

	columnFilterScroll := container.NewVScroll(columnFilterContainer)
	columnFilterScroll.SetMinSize(fyne.NewSize(200, 200))

	columnFilterCard := widget.NewCard("", "Select Columns", columnFilterScroll)

	// Create query parser
	queryParser := NewQueryParser(dataItem.header)
	var errorLabel *widget.Label

	// Track last valid query to avoid re-filtering on incomplete expressions
	var lastValidQuery *Query

	// Declare applyFilters function variable
	var applyFilters func()

	// Create search button (after declaring applyFilters)
	searchButton := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		applyFilters()
	})

	// Create clear search button
	clearSearchBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		searchEntry.SetText("")
		applyFilters() // Apply filters to show all records
	})

	// Define the applyFilters function
	applyFilters = func() {
		searchText := strings.TrimSpace(searchEntry.Text)

		// Filter columns
		visibleCols := make([]int, 0)
		filteredHeader := make([]string, 0)

		for i, colName := range dataItem.header {
			if check, exists := columnChecks[colName]; exists && check.Checked {
				visibleCols = append(visibleCols, i)
				filteredHeader = append(filteredHeader, colName)
			}
		}

		dataItem.visibleColumns = visibleCols
		dataItem.filteredHeader = filteredHeader

		// Parse query
		query, err := queryParser.ParseQuery(searchText)
		if err != nil {
			// Show error but don't filter - use last valid query
			if errorLabel != nil {
				errorLabel.SetText(fmt.Sprintf("Query error: %v", err))
				errorLabel.Show()
			}
			// Keep using last valid query for filtering
			query = lastValidQuery
		} else {
			// Hide error label if query is valid
			if errorLabel != nil {
				errorLabel.Hide()
			}
			// Update last valid query
			lastValidQuery = query
		}

		// Filter rows based on query
		if query == nil || len(query.Expressions) == 0 {
			// No search filter - show all rows with visible columns
			filteredData := make([][]string, len(dataItem.data))
			filteredRowIndices := make([]int, len(dataItem.data))
			for i, row := range dataItem.data {
				newRow := make([]string, len(visibleCols))
				for j, colIdx := range visibleCols {
					newRow[j] = row[colIdx]
				}
				filteredData[i] = newRow
				filteredRowIndices[i] = i
			}
			dataItem.filteredData = filteredData
			dataItem.filteredRowIndices = filteredRowIndices
		} else {
			// Apply query filter
			filteredData := make([][]string, 0)
			filteredRowIndices := make([]int, 0)
			for rowIdx, row := range dataItem.data {
				// Evaluate query against full row
				if queryParser.EvaluateRow(query, row, dataItem.header) {
					// Create filtered row with only visible columns
					newRow := make([]string, len(visibleCols))
					for j, colIdx := range visibleCols {
						newRow[j] = row[colIdx]
					}
					filteredData = append(filteredData, newRow)
					filteredRowIndices = append(filteredRowIndices, rowIdx)
				}
			}
			dataItem.filteredData = filteredData
			dataItem.filteredRowIndices = filteredRowIndices
		}

		table.Refresh()
	}

	// Connect search entry to filter only on Enter key
	searchEntry.OnSubmitted = func(string) {
		applyFilters()
	}

	// Add change handlers to column checkboxes
	for _, check := range columnChecks {
		check.OnChanged = func(bool) {
			applyFilters()
		}
	}

	// Create filter buttons for column selection
	selectAllBtn := widget.NewButton("Select All", func() {
		for _, check := range columnChecks {
			check.SetChecked(true)
		}
		applyFilters()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		for _, check := range columnChecks {
			check.SetChecked(false)
		}
		applyFilters()
	})

	filterButtons := container.NewHBox(selectAllBtn, deselectAllBtn)

	// Combine search and column filter in an accordion
	filterAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Column Filter", container.NewBorder(filterButtons, nil, nil, nil, columnFilterCard)),
	)

	// Create export menu
	exportMenu := fyne.NewMenu("Export",
		fyne.NewMenuItem("Export as Parquet", func() {
			t.exportData(dataItem, FormatParquet, delta_table.Name)
		}),
		fyne.NewMenuItem("Export as CSV", func() {
			t.exportData(dataItem, FormatCSV, delta_table.Name)
		}),
		fyne.NewMenuItem("Export as JSON", func() {
			t.exportData(dataItem, FormatJSON, delta_table.Name)
		}),
	)

	// Create export button with menu
	var exportMenuBtn *widget.Button
	exportMenuBtn = widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), func() {
		widget.ShowPopUpMenuAtPosition(exportMenu, t.w.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(exportMenuBtn))
	})

	// Create toolbar with button aligned to the right
	exportToolbar := container.NewBorder(nil, nil, nil, exportMenuBtn)

	// Create table card with column and row count
	rowCount := len(dataItem.data)
	colCount := len(dataItem.header)
	cardTitle := fmt.Sprintf("Table %s (%d columns x %d rows)", delta_table.Name, colCount, rowCount)

	// Create error label for query parsing errors
	errorLabel = widget.NewLabel("")
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.Importance = widget.HighImportance
	errorLabel.Hide()

	// Create help text for query syntax
	helpText := widget.NewLabel("Query Syntax:\n" +
		"• Simple search: john (searches all columns)\n" +
		"• Exact match: name = john or name = 'john'\n" +
		"• Not equal: status != active\n" +
		"• Comparison: age > 18, price <= 100, score >= 90\n" +
		"• Contains: name ~ john (case-insensitive)\n" +
		"• Logic: age > 18 AND status = active\n" +
		"• Multiple conditions: age > 18 OR age < 65")
	helpText.Wrapping = fyne.TextWrapWord

	helpAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Query Help", helpText),
	)

	// Create search bar with entry and buttons
	searchButtonsContainer := container.NewHBox(searchButton, clearSearchBtn)
	searchBar := container.NewBorder(nil, nil, nil, searchButtonsContainer, searchEntry)

	// Combine search, filters, and table in a vertical layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewCard("", cardTitle, nil),
			exportToolbar,
			widget.NewSeparator(),
			widget.NewLabel("Row Search (Press Enter or click Search button):"),
			searchBar,
			errorLabel,
			helpAccordion,
			filterAccordion,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		table,
	)

	// Add new tab to the persistent inner tabs
	// Append " filtered" to tab name if filtering was applied
	tabName := delta_table.Name
	if dataItem.isFiltered {
		tabName = delta_table.Name + " filtered"
	}
	newTab := container.NewTabItem(tabName, content)

	// Store the tab reference in the data item and register in map
	dataItem.tab = newTab
	dataItem.tableName = delta_table.Name
	t.tabDataMap[newTab] = dataItem

	t.innerTabs.Append(newTab)

	// Select the newly added tab
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
		// Recreate the Browser tab
		t.browserTab = container.NewTabItem("Browser", t.innerTabs)
		t.docTabs.Append(t.browserTab)
	}

	// Select the Browser tab in the main tabs
	t.docTabs.Select(t.browserTab)
}

func (t *DataBrowser) GetData(profile string, table delta_sharing.Table, file_id string, options *QueryOptions) {
	c := make(chan bool)
	go func(c chan bool) {
		pbi := widget.NewProgressBarInfinite()

		di := dialog.NewCustomWithoutButtons("Please wait", pbi, t.w)
		di.Resize(fyne.NewSize(200, 100))
		di.Show()
		pbi.Start()
		for {
			select {
			case <-c:
				di.Hide()
				pbi.Stop()
				return
			default:
				time.Sleep(time.Millisecond + 500)
			}
		}
	}(c)
	ds, err := delta_sharing.NewSharingClientFromString(context.Background(), profile, "")
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	resp, err := ds.ListFilesInTable(table)
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	var data Data
	for _, v := range resp.AddFiles {
		if v.Id == file_id {
			arrow_table, err := delta_sharing.LoadArrowTable(ds, table, file_id)
			if err != nil {
				dialog.NewError(err, t.w).Show()
			}
			data.arrow_table = arrow_table

			// Apply query options if provided
			if options != nil {
				data.arrow_table, err = t.applyQueryOptions(data.arrow_table, options)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to apply query options: %w", err), t.w)
					c <- true
					return
				}
				// Mark data as filtered when options are applied
				data.isFiltered = true
			}

			data.arrow_table, err = t.test(data.arrow_table)
			if err != nil {
				fmt.Println(err)
				c <- true
				return
			}
			var header []string = make([]string, data.arrow_table.NumCols())
			for i, f := range data.arrow_table.Schema().Fields() {
				header[i] = f.Name
			}

			data.data = make([][]string, 0)
			data.header = header

			// Determine batch size based on limit
			batchSize := int64(1000)
			if options != nil && options.Limit > 0 && options.Limit < batchSize {
				batchSize = options.Limit
			}

			tr := array.NewTableReader(data.arrow_table, batchSize)
			tr.Retain()
			tr.Next()
			data.arrow_rec = tr.Record()
			t.Data = append(t.Data, data)
			dt := t.parseRecord(options)

			t.CreateDataBrowser(dt, table)

			c <- true
			t.w.Content().Refresh()
		}
	}
}

func (t *DataBrowser) parseRecord(options *QueryOptions) *Data {
	dp := len(t.Data) - 1
	maxRows := int(t.Data[dp].arrow_rec.NumRows())

	// Apply row limit if specified
	if options != nil && options.Limit > 0 && int(options.Limit) < maxRows {
		maxRows = int(options.Limit)
	}

	for pos := 0; pos < maxRows; pos++ {
		var v []string = make([]string, t.Data[dp].arrow_rec.NumCols())
		for i, col := range t.Data[dp].arrow_rec.Columns() {
			switch col.DataType().ID() {
			case arrow.STRUCT:
				s := col.(*array.Struct)

				b, err := s.MarshalJSON()
				if err != nil {
					log.Fatal(err)
				}
				v[i] = string(b)

			case arrow.LIST:
				as := array.NewSlice(col, int64(pos), int64(pos+1))
				str := fmt.Sprintf("%v", as)
				if len(str) > 253 {
					v[i] = str[1:253] + "..."
				} else {
					v[i] = str
				}
			case arrow.STRING:
				s := col.(*array.String)
				v[i] = s.Value(pos)
			case arrow.BINARY:
				b := col.(*array.Binary)
				v[i] = string(b.Value(pos))
			case arrow.BOOL:
				b := col.(*array.Boolean)
				v[i] = fmt.Sprintf("%v", b.Value(pos))
			case arrow.DATE32:
				d32 := col.(*array.Date32)
				v[i] = d32.Value(pos).ToTime().String()
			case arrow.DATE64:
				d64 := col.(*array.Date64)
				v[i] = d64.Value(pos).ToTime().String()
			case arrow.DECIMAL:
				d128 := col.(*array.Decimal128)
				v[i] = d128.Value(pos).BigInt().String()
			case arrow.INT8:
				i8 := col.(*array.Int8)
				v[i] = fmt.Sprintf("%d", i8.Value(pos))
			case arrow.INT16:
				i16 := col.(*array.Int16)
				v[i] = fmt.Sprintf("%d", i16.Value(pos))
			case arrow.INT32:
				i32 := col.(*array.Int32)
				v[i] = fmt.Sprintf("%d", i32.Value(pos))
			case arrow.INT64:
				i64 := col.(*array.Int64)
				v[i] = fmt.Sprintf("%d", i64.Value(pos))
			case arrow.FLOAT16:
				f16 := col.(*array.Float16)
				v[i] = f16.Value(pos).String()
			case arrow.FLOAT32:
				f32 := col.(*array.Float32)
				v[i] = fmt.Sprintf("%.2f", f32.Value(pos))
			case arrow.FLOAT64:
				f64 := col.(*array.Float64)
				v[i] = fmt.Sprintf("%.2f", f64.Value(pos))
			case arrow.INTERVAL_MONTHS:
				intV := col.(*array.DayTimeInterval)
				v[i] = fmt.Sprintf("%v", intV.Value(pos))
			case arrow.INTERVAL_DAY_TIME:
				intV := col.(*array.DayTimeInterval)
				v[i] = fmt.Sprintf("%v", intV.Value(pos))
			case arrow.TIMESTAMP:
				ts := col.(*array.Timestamp)
				v[i] = ts.Value(pos).ToTime(arrow.Nanosecond).String()
			}
		}
		t.Data[dp].data = append(t.Data[dp].data, v)
	}
	// Don't release Arrow resources here - they will be released when the tab is closed
	// t.Data[dp].arrow_rec.Release()
	// t.Data[dp].arrow_table.Release()
	return &t.Data[dp]
}

func (d *DataBrowser) test(t arrow.Table) (arrow.Table, error) {
	/* table := t

	pool := compute.GetAllocator(context.Background())
	intBuilder := array.NewFloat32Builder(pool)
	intBuilder.Append(21)
	a := intBuilder.NewArray()
	//c := compute.Equal(compute.NewFieldRef("deaths"), compute.NewLiteral(21))
	tab2, err := compute.FilterTable(context.Background(), table, compute.NewDatum(a), compute.DefaultFilterOptions())
	if err != nil {
		return nil, err
	}
	*/
	return t, nil
}

// applyQueryOptions applies column selection and row limiting to the Arrow table
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

	// Note: Predicate filtering would require more complex SQL parsing
	// For now, we'll display a message if a predicate is provided
	if options.Predicate != "" {
		// Predicate filtering is complex and would require SQL parsing
		// This is a placeholder for future implementation
		log.Printf("Predicate filtering requested but not yet implemented: %s", options.Predicate)
	}

	return table, nil
}

// exportData handles the export of data to different formats
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

		// Show progress indicator
		pbi := widget.NewProgressBarInfinite()
		progressDialog := dialog.NewCustomWithoutButtons("Exporting...", pbi, t.w)
		progressDialog.Resize(fyne.NewSize(300, 100))
		progressDialog.Show()
		pbi.Start()

		// Export in a goroutine
		go func() {
			var exportErr error

			// Convert filtered data to Arrow table for export
			filteredTable, convErr := t.createFilteredArrowTable(dataItem)
			if convErr != nil {
				pbi.Stop()
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("failed to prepare filtered data: %w", convErr), t.w)
				return
			}
			defer filteredTable.Release()

			switch format {
			case FormatParquet:
				exportErr = ExportToParquet(filteredTable, filePath)
			case FormatCSV:
				exportErr = ExportToCSV(filteredTable, filePath)
			case FormatJSON:
				exportErr = ExportToJSON(filteredTable, filePath)
			}

			// Hide progress dialog
			pbi.Stop()
			progressDialog.Hide()

			// Show result
			if exportErr != nil {
				dialog.ShowError(fmt.Errorf("export failed: %w", exportErr), t.w)
			} else {
				dialog.ShowInformation("Export Successful",
					fmt.Sprintf("Data exported successfully to:\n%s", filePath), t.w)
			}
		}()
	}, t.w)

	// Set default filename
	defaultName := strings.ReplaceAll(tableName, " ", "_") + ext
	saveDialog.SetFileName(defaultName)

	// Set file filter
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{ext}))

	saveDialog.Show()
}

// createFilteredArrowTable creates an Arrow table from the filtered data
func (t *DataBrowser) createFilteredArrowTable(dataItem *Data) (arrow.Table, error) {
	if len(dataItem.filteredData) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	// Get the original schema to determine column types
	originalSchema := dataItem.arrow_table.Schema()

	// Build new schema with only visible columns
	newFields := make([]arrow.Field, len(dataItem.visibleColumns))
	for i, colIdx := range dataItem.visibleColumns {
		newFields[i] = originalSchema.Field(colIdx)
	}
	schema := arrow.NewSchema(newFields, nil)

	// Create memory pool
	pool := memory.NewGoAllocator()

	// Get table reader to access typed values (shared across all columns)
	tr := array.NewTableReader(dataItem.arrow_table, dataItem.arrow_table.NumRows())
	defer tr.Release()
	tr.Next()
	rec := tr.Record()

	// Build Arrow arrays for each column using the tracked row indices
	columns := make([]arrow.Column, len(dataItem.visibleColumns))
	for i, colIdx := range dataItem.visibleColumns {
		field := originalSchema.Field(colIdx)

		// Create builder based on data type
		builder := array.NewBuilder(pool, field.Type)
		defer builder.Release()

		// Append values from the original Arrow column using tracked indices
		// filteredRowIndices maps filtered row position to original row position
		for _, originalRowIdx := range dataItem.filteredRowIndices {
			col := rec.Column(colIdx)
			appendValueToBuilder(builder, col, originalRowIdx)
		}

		// Build the array
		arr := builder.NewArray()
		defer arr.Release()

		// Create chunked array
		chunked := arrow.NewChunked(field.Type, []arrow.Array{arr})
		columns[i] = *arrow.NewColumn(field, chunked)
	}

	// Create and return the table
	return array.NewTable(schema, columns, int64(len(dataItem.filteredData))), nil
}

// formatValueFromArray formats a value from an Arrow array (helper for matching)
func formatValueFromArray(col arrow.Array, pos int) string {
	if col.IsNull(pos) {
		return ""
	}

	switch col.DataType().ID() {
	case arrow.STRUCT:
		s := col.(*array.Struct)
		b, _ := s.MarshalJSON()
		return string(b)
	case arrow.LIST:
		as := array.NewSlice(col, int64(pos), int64(pos+1))
		str := fmt.Sprintf("%v", as)
		if len(str) > 253 {
			return str[1:253] + "..."
		}
		return str
	case arrow.STRING:
		s := col.(*array.String)
		return s.Value(pos)
	case arrow.BINARY:
		b := col.(*array.Binary)
		return string(b.Value(pos))
	case arrow.BOOL:
		b := col.(*array.Boolean)
		return fmt.Sprintf("%v", b.Value(pos))
	case arrow.DATE32:
		d32 := col.(*array.Date32)
		return d32.Value(pos).ToTime().String()
	case arrow.DATE64:
		d64 := col.(*array.Date64)
		return d64.Value(pos).ToTime().String()
	case arrow.DECIMAL:
		d128 := col.(*array.Decimal128)
		return d128.Value(pos).BigInt().String()
	case arrow.INT8:
		i8 := col.(*array.Int8)
		return fmt.Sprintf("%d", i8.Value(pos))
	case arrow.INT16:
		i16 := col.(*array.Int16)
		return fmt.Sprintf("%d", i16.Value(pos))
	case arrow.INT32:
		i32 := col.(*array.Int32)
		return fmt.Sprintf("%d", i32.Value(pos))
	case arrow.INT64:
		i64 := col.(*array.Int64)
		return fmt.Sprintf("%d", i64.Value(pos))
	case arrow.FLOAT16:
		f16 := col.(*array.Float16)
		return f16.Value(pos).String()
	case arrow.FLOAT32:
		f32 := col.(*array.Float32)
		return fmt.Sprintf("%.2f", f32.Value(pos))
	case arrow.FLOAT64:
		f64 := col.(*array.Float64)
		return fmt.Sprintf("%.2f", f64.Value(pos))
	case arrow.TIMESTAMP:
		ts := col.(*array.Timestamp)
		return ts.Value(pos).ToTime(arrow.Nanosecond).String()
	default:
		return fmt.Sprintf("%v", col)
	}
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
