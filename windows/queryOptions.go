package windows

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/apache/arrow-go/v18/arrow"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

// QueryOptions holds the query configuration for table data loading
type QueryOptions struct {
	SelectedColumns []string
	Predicate       string
	Limit           int64
}

// QueryOptionsDialog creates a dialog for configuring query options
type QueryOptionsDialog struct {
	dialog         dialog.Dialog
	window         fyne.Window
	schema         *arrow.Schema
	columnChecks   map[string]*widget.Check
	predicateEntry *widget.Entry
	limitEntry     *widget.Entry
	callback       func(*QueryOptions)
}

// NewQueryOptionsDialog creates a new query options dialog
func NewQueryOptionsDialog(w fyne.Window, schema *arrow.Schema, callback func(*QueryOptions)) *QueryOptionsDialog {
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
		for _, field := range qod.schema.Fields() {
			check := widget.NewCheck(fmt.Sprintf("%s (%s)", field.Name, field.Type), nil)
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

// SimpleQueryOptionsDialog creates a simplified query options dialog without schema
func SimpleQueryOptionsDialog(w fyne.Window, callback func(*QueryOptions)) {
	predicateEntry := widget.NewMultiLineEntry()
	predicateEntry.SetPlaceHolder("e.g., age > 25 AND status = 'active'")
	predicateEntry.SetMinRowsVisible(3)

	limitEntry := widget.NewEntry()
	limitEntry.SetPlaceHolder("Leave empty for all rows")

	predicateLabel := widget.NewLabel("Filter Predicate (SQL WHERE clause):")
	predicateLabel.TextStyle = fyne.TextStyle{Bold: true}

	limitLabel := widget.NewLabel("Row Limit:")
	limitLabel.TextStyle = fyne.TextStyle{Bold: true}

	limitHelp := widget.NewLabel("Maximum number of rows to return.")
	limitHelp.TextStyle = fyne.TextStyle{Italic: true}

	columnEntry := widget.NewEntry()
	columnEntry.SetPlaceHolder("Leave empty for all columns, or comma-separated list")

	columnLabel := widget.NewLabel("Select Columns:")
	columnLabel.TextStyle = fyne.TextStyle{Bold: true}

	columnHelp := widget.NewLabel("Comma-separated column names (e.g., id,name,age)")
	columnHelp.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		columnLabel,
		columnEntry,
		columnHelp,
		widget.NewSeparator(),
		predicateLabel,
		predicateEntry,
		widget.NewLabel("Leave empty for no filtering."),
		widget.NewSeparator(),
		limitLabel,
		limitEntry,
		limitHelp,
	)

	d := dialog.NewCustomConfirm(
		"Query Options",
		"Load Data",
		"Cancel",
		content,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			options := &QueryOptions{
				SelectedColumns: make([]string, 0),
			}

			// Parse columns
			colText := strings.TrimSpace(columnEntry.Text)
			if colText != "" {
				cols := strings.Split(colText, ",")
				for _, col := range cols {
					trimmed := strings.TrimSpace(col)
					if trimmed != "" {
						options.SelectedColumns = append(options.SelectedColumns, trimmed)
					}
				}
			}

			// Get predicate
			options.Predicate = strings.TrimSpace(predicateEntry.Text)

			// Get limit
			limitText := strings.TrimSpace(limitEntry.Text)
			if limitText != "" {
				limit, err := strconv.ParseInt(limitText, 10, 64)
				if err != nil || limit <= 0 {
					dialog.ShowError(fmt.Errorf("invalid limit: must be a positive number"), w)
					return
				}
				options.Limit = limit
			} else {
				options.Limit = -1 // No limit
			}

			if callback != nil {
				callback(options)
			}
		},
		w,
	)

	d.Resize(fyne.NewSize(500, 500))
	d.Show()
}

// ListWithContextMenu creates a list widget with context menu support
func NewListWithContextMenu(data binding.StringList, onSelected func(widget.ListItemID), onContextMenu func(widget.ListItemID)) *widget.List {
	list := widget.NewListWithData(data, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(di binding.DataItem, co fyne.CanvasObject) {
		co.(*widget.Label).Bind(di.(binding.String))
	})

	if onSelected != nil {
		list.OnSelected = onSelected
	}

	return list
}

// ShowQueryOptionsDialogWithSchema loads table schema and shows enhanced query options dialog
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
		ds, err := delta_sharing.NewSharingClientFromString(profile)
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to create client: %w", err), w)
			return
		}

		resp, err := ds.ListFilesInTable(context.Background(), table)
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to list files: %w", err), w)
			return
		}

		if len(resp.AddFiles) == 0 {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("no files available for table"), w)
			return
		}

		fileID := resp.AddFiles[0].Id
		arrowTable, err := delta_sharing.LoadArrowTable(context.Background(), ds, table, fileID)
		if err != nil {
			progressBar.Stop()
			progressDialog.Hide()
			time.Sleep(50 * time.Millisecond)
			dialog.ShowError(fmt.Errorf("failed to load schema: %w", err), w)
			return
		}

		schema := arrowTable.Schema()
		arrowTable.Release()

		// Close progress dialog
		progressBar.Stop()
		progressDialog.Hide()

		// Brief delay to allow dialog to fully close
		time.Sleep(100 * time.Millisecond)

		// Create and show query options dialog
		qod := NewQueryOptionsDialog(w, schema, callback)
		qod.Show()
	}()
}
