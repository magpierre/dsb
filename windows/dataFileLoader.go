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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	arrowadapter "github.com/magpierre/fyne-datatable/adapters/arrow"
	csvadapter "github.com/magpierre/fyne-datatable/adapters/csv"
	sliceadapter "github.com/magpierre/fyne-datatable/adapters/slice"
	"github.com/magpierre/fyne-datatable/datatable"
	fynewidget "github.com/magpierre/fyne-datatable/widget"
)

// FileType represents the type of data file
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeCSV
	FileTypeParquet
	FileTypeJSON
	FileTypeDeltaSharingProfile
)

// DetectFileType determines the type of file based on extension and content
func DetectFileType(filePath string, content string) FileType {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv":
		return FileTypeCSV
	case ".parquet":
		return FileTypeParquet
	case ".json", ".share", ".txt":
		// Try to detect if it's a Delta Sharing profile or JSON data
		if isDeltaSharingProfile(content) {
			return FileTypeDeltaSharingProfile
		}
		return FileTypeJSON
	default:
		return FileTypeUnknown
	}
}

// isDeltaSharingProfile checks if the content looks like a Delta Sharing profile
func isDeltaSharingProfile(content string) bool {
	// A Delta Sharing profile typically has shareCredentialsVersion, endpoint, bearerToken
	var profile map[string]interface{}
	if err := json.Unmarshal([]byte(content), &profile); err != nil {
		return false
	}

	// Check for required fields
	_, hasVersion := profile["shareCredentialsVersion"]
	_, hasEndpoint := profile["endpoint"]
	_, hasBearerToken := profile["bearerToken"]

	return hasVersion && hasEndpoint && hasBearerToken
}

// detectCSVSeparator tries to detect the CSV separator from the first line
func detectCSVSeparator(filePath string) (rune, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ',', fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		// Empty file or error, use default comma
		return ',', nil
	}

	firstLine := scanner.Text()
	if firstLine == "" {
		return ',', nil
	}

	// Count occurrences of common separators
	separators := map[rune]int{
		',':  strings.Count(firstLine, ","),
		';':  strings.Count(firstLine, ";"),
		'\t': strings.Count(firstLine, "\t"),
		'|':  strings.Count(firstLine, "|"),
	}

	// Find the separator with the highest count
	maxCount := 0
	detectedSep := ','
	for sep, count := range separators {
		if count > maxCount {
			maxCount = count
			detectedSep = sep
		}
	}

	// If no separator was found (all counts are 0), default to comma
	if maxCount == 0 {
		return ',', nil
	}

	return detectedSep, nil
}

// getSeparatorName returns a human-readable name for the separator
func getSeparatorName(sep rune) string {
	switch sep {
	case ',':
		return "comma"
	case ';':
		return "semicolon"
	case '\t':
		return "tab"
	case '|':
		return "pipe"
	default:
		return string(sep)
	}
}

// LoadDataFile loads a data file using the appropriate adapter and displays it
func (t *MainWindow) LoadDataFile(filePath string) error {
	fileType := DetectFileType(filePath, "")

	switch fileType {
	case FileTypeCSV:
		return t.loadCSVFile(filePath)
	case FileTypeParquet:
		return t.loadParquetFile(filePath)
	case FileTypeJSON:
		return t.loadJSONFile(filePath)
	default:
		return fmt.Errorf("unsupported file type")
	}
}

// loadCSVFile loads a CSV file using the CSV adapter
func (t *MainWindow) loadCSVFile(filePath string) error {
	t.SetStatus("Loading CSV file: " + filepath.Base(filePath))

	// Detect the CSV separator from the first line
	separator, err := detectCSVSeparator(filePath)
	if err != nil {
		separator = ','
	}

	// Use CSV adapter to load the file with detected separator
	config := csvadapter.DefaultConfig()
	config.HasHeaders = true
	config.TrimSpace = true
	config.Delimiter = separator

	dataSource, err := csvadapter.NewFromFile(filePath, config)
	if err != nil {
		return fmt.Errorf("failed to load CSV file: %w", err)
	}

	// Create datatable model
	model, err := datatable.NewTableModel(dataSource)
	if err != nil {
		return fmt.Errorf("failed to create table model: %w", err)
	}

	// Display the data
	t.displayDataTable(model, filepath.Base(filePath))

	// Show which separator was detected
	separatorName := getSeparatorName(separator)
	t.SetStatus(fmt.Sprintf("Loaded CSV file: %s (%d rows, %d columns, separator: %s)",
		filepath.Base(filePath), dataSource.RowCount(), dataSource.ColumnCount(), separatorName))

	return nil
}

// loadParquetFile loads a Parquet file using the Arrow adapter
func (t *MainWindow) loadParquetFile(filePath string) error {
	t.SetStatus("Loading Parquet file: " + filepath.Base(filePath))

	// Open the parquet file
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer f.Close()

	// Get file info for size
	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create a parquet file reader
	pf, err := file.NewParquetReader(f, file.WithReadProps(&parquet.ReaderProperties{}))
	if err != nil {
		return fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer pf.Close()

	// Convert parquet to Arrow table
	mem := memory.NewGoAllocator()
	arrowReader, err := pqarrow.NewFileReader(pf, pqarrow.ArrowReadProperties{}, mem)
	if err != nil {
		return fmt.Errorf("failed to create arrow reader: %w", err)
	}

	// Read all data into an Arrow table
	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return fmt.Errorf("failed to read parquet data: %w", err)
	}
	defer table.Release()

	// Create Arrow adapter
	dataSource, err := arrowadapter.NewFromArrowTable(table)
	if err != nil {
		return fmt.Errorf("failed to create arrow data source: %w", err)
	}

	// Create datatable model
	model, err := datatable.NewTableModel(dataSource)
	if err != nil {
		return fmt.Errorf("failed to create table model: %w", err)
	}

	// Display the data
	t.displayDataTable(model, filepath.Base(filePath))
	t.SetStatus(fmt.Sprintf("Loaded Parquet file: %s (%d rows, %d columns, %.2f MB)",
		filepath.Base(filePath), dataSource.RowCount(), dataSource.ColumnCount(),
		float64(fileInfo.Size())/(1024*1024)))

	return nil
}

// loadJSONFile loads a JSON file using the slice adapter
func (t *MainWindow) loadJSONFile(filePath string) error {
	t.SetStatus("Loading JSON file: " + filepath.Base(filePath))

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Try to parse as array of objects
	var data []map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		// Try as single object
		var singleObj map[string]interface{}
		if err := json.Unmarshal(content, &singleObj); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
		data = []map[string]interface{}{singleObj}
	}

	if len(data) == 0 {
		return fmt.Errorf("JSON file is empty or has no records")
	}

	// Use slice adapter to create data source
	dataSource, err := sliceadapter.NewFromMaps(data)
	if err != nil {
		return fmt.Errorf("failed to create data source from JSON: %w", err)
	}

	// Create datatable model
	model, err := datatable.NewTableModel(dataSource)
	if err != nil {
		return fmt.Errorf("failed to create table model: %w", err)
	}

	// Display the data
	t.displayDataTable(model, filepath.Base(filePath))
	t.SetStatus(fmt.Sprintf("Loaded JSON file: %s (%d rows, %d columns)",
		filepath.Base(filePath), dataSource.RowCount(), dataSource.ColumnCount()))

	return nil
}

// displayDataTable creates and displays a datatable widget with the given model
func (t *MainWindow) displayDataTable(model *datatable.TableModel, tabName string) {
	// Create widget with configuration - row selection enabled for copy functionality
	config := fynewidget.DefaultConfig()
	config.ShowFilterBar = true
	config.ShowStatusBar = true
	config.ShowColumnSelector = true
	config.ShowSettingsButton = true
	config.AutoAdjustColumnWidths = true
	config.SelectionMode = fynewidget.SelectionModeRow // Enable row selection for copy functionality (default)
	config.MinColumnWidth = 100

	dt := fynewidget.NewDataTableWithConfig(model, config)

	// Set window reference for keyboard shortcuts
	dt.SetWindow(t.w)

	scroll := container.NewScroll(dt)
	// Create a scroll container for the datatable

	// Create a card to hold the datatable
	card := widget.NewCard("", tabName, scroll)

	// Add to doc tabs
	if t.docTabs != nil {
		// Check if a tab with this name already exists
		for _, tab := range t.docTabs.Items {
			if tab.Text == tabName {
				// Update existing tab
				tab.Content = card
				t.docTabs.Select(tab)
				return
			}
		}

		// Create new tab
		tabItem := container.NewTabItem(tabName, card)
		t.docTabs.Append(tabItem)
		t.docTabs.Select(tabItem)
	}
}

// Helper method to handle file loading with error dialogs
func (t *MainWindow) handleDataFileLoad(filePath string) {
	go func() {
		err := t.LoadDataFile(filePath)
		if err != nil {
			// Show error on UI thread by creating a closure that captures the error
			errMsg := err.Error()
			t.a.SendNotification(&fyne.Notification{
				Title:   "Error Loading File",
				Content: errMsg,
			})
			fmt.Println("Error loading file: " + errMsg)
			t.SetStatus("Error loading file: " + errMsg)
		}
	}()
}
