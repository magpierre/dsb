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

// Package export provides streaming export functionality for table data.
package export

import (
	"fmt"
	"io"

	"github.com/magpierre/fyne-datatable/datatable"
)

// ProgressCallback is called during export to report progress.
// It receives the number of rows exported so far and the total number of rows.
// Return false to cancel the export.
type ProgressCallback func(current, total int) bool

// RowIterator provides row-by-row access to table data.
// This enables streaming export without loading all data into memory.
type RowIterator interface {
	// Next advances to the next row. Returns false when no more rows.
	Next() bool

	// Row returns the current row's values.
	// Returns error if called before first Next() or after Next() returns false.
	Row() ([]datatable.Value, error)

	// RowNumber returns the current row number (0-based).
	RowNumber() int

	// TotalRows returns the total number of rows, if known.
	// Returns -1 if total is unknown (e.g., infinite stream).
	TotalRows() int

	// ColumnNames returns the column names.
	ColumnNames() []string

	// ColumnTypes returns the column data types.
	ColumnTypes() []datatable.DataType

	// Err returns any error encountered during iteration.
	Err() error
}

// Exporter exports data in a specific format.
type Exporter interface {
	// Export writes data from the iterator to the writer.
	// Calls progress callback periodically if provided (can be nil).
	// Returns number of rows exported and any error.
	Export(writer io.Writer, iterator RowIterator, progress ProgressCallback) (int, error)

	// FileExtension returns the recommended file extension (e.g., "csv", "json").
	FileExtension() string

	// MimeType returns the MIME type for this format.
	MimeType() string

	// Description returns a human-readable description of this format.
	Description() string
}

// Engine coordinates export operations.
type Engine struct {
	// Stateless - no fields needed
}

// NewEngine creates a new ExportEngine.
func NewEngine() *Engine {
	return &Engine{}
}

// Export exports data using the specified exporter.
// This is a convenience method that delegates to the exporter.
func (e *Engine) Export(
	writer io.Writer,
	iterator RowIterator,
	exporter Exporter,
	progress ProgressCallback,
) (int, error) {
	if writer == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}
	if iterator == nil {
		return 0, fmt.Errorf("iterator cannot be nil")
	}
	if exporter == nil {
		return 0, fmt.Errorf("exporter cannot be nil")
	}

	return exporter.Export(writer, iterator, progress)
}

// ExportToFile is a helper for exporting to a file.
// The caller is responsible for creating and closing the file.
func (e *Engine) ExportToFile(
	writer io.Writer,
	iterator RowIterator,
	exporter Exporter,
	progress ProgressCallback,
) (int, error) {
	return e.Export(writer, iterator, exporter, progress)
}

// ValidateExporter checks if an exporter is properly implemented.
func (e *Engine) ValidateExporter(exporter Exporter) error {
	if exporter == nil {
		return fmt.Errorf("exporter cannot be nil")
	}

	if exporter.FileExtension() == "" {
		return fmt.Errorf("exporter must provide file extension")
	}

	if exporter.MimeType() == "" {
		return fmt.Errorf("exporter must provide MIME type")
	}

	if exporter.Description() == "" {
		return fmt.Errorf("exporter must provide description")
	}

	return nil
}
