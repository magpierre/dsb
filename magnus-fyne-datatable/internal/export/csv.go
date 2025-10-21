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

package export

import (
	"encoding/csv"
	"fmt"
	"io"
)

// CSVConfig configures CSV export options.
type CSVConfig struct {
	// Delimiter is the field delimiter (default: comma)
	Delimiter rune

	// IncludeHeaders determines if column names are written as first row
	IncludeHeaders bool

	// UseCRLF determines if lines end with \r\n instead of \n
	UseCRLF bool
}

// DefaultCSVConfig returns the default CSV configuration.
func DefaultCSVConfig() CSVConfig {
	return CSVConfig{
		Delimiter:      ',',
		IncludeHeaders: true,
		UseCRLF:        false,
	}
}

// CSVExporter exports data in CSV format.
type CSVExporter struct {
	config CSVConfig
}

// NewCSVExporter creates a new CSV exporter with default configuration.
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{
		config: DefaultCSVConfig(),
	}
}

// NewCSVExporterWithConfig creates a CSV exporter with custom configuration.
func NewCSVExporterWithConfig(config CSVConfig) *CSVExporter {
	return &CSVExporter{
		config: config,
	}
}

// Export writes data in CSV format.
func (e *CSVExporter) Export(
	writer io.Writer,
	iterator RowIterator,
	progress ProgressCallback,
) (int, error) {
	if writer == nil {
		return 0, fmt.Errorf("writer cannot be nil")
	}
	if iterator == nil {
		return 0, fmt.Errorf("iterator cannot be nil")
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	csvWriter.Comma = e.config.Delimiter
	csvWriter.UseCRLF = e.config.UseCRLF

	// Write headers if requested
	if e.config.IncludeHeaders {
		headers := iterator.ColumnNames()
		if err := csvWriter.Write(headers); err != nil {
			return 0, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Export rows
	rowCount := 0
	totalRows := iterator.TotalRows()

	for iterator.Next() {
		row, err := iterator.Row()
		if err != nil {
			return rowCount, fmt.Errorf("failed to get row %d: %w", rowCount, err)
		}

		// Convert Values to strings
		record := make([]string, len(row))
		for i, val := range row {
			if val.IsNull {
				record[i] = "" // Empty string for null values
			} else {
				record[i] = val.Formatted
			}
		}

		// Write the record
		if err := csvWriter.Write(record); err != nil {
			return rowCount, fmt.Errorf("failed to write row %d: %w", rowCount, err)
		}

		rowCount++

		// Report progress if callback provided
		if progress != nil {
			if !progress(rowCount, totalRows) {
				// User cancelled
				csvWriter.Flush()
				return rowCount, fmt.Errorf("export cancelled by user")
			}
		}
	}

	// Check for iteration errors
	if err := iterator.Err(); err != nil {
		return rowCount, fmt.Errorf("iterator error: %w", err)
	}

	// Flush any buffered data
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return rowCount, fmt.Errorf("failed to flush CSV data: %w", err)
	}

	return rowCount, nil
}

// FileExtension returns "csv".
func (e *CSVExporter) FileExtension() string {
	return "csv"
}

// MimeType returns the CSV MIME type.
func (e *CSVExporter) MimeType() string {
	return "text/csv"
}

// Description returns a human-readable description.
func (e *CSVExporter) Description() string {
	if e.config.Delimiter == ',' {
		return "Comma-Separated Values (CSV)"
	} else if e.config.Delimiter == '\t' {
		return "Tab-Separated Values (TSV)"
	}
	return fmt.Sprintf("Delimited Values (delimiter: %q)", e.config.Delimiter)
}

// GetConfig returns the current configuration.
func (e *CSVExporter) GetConfig() CSVConfig {
	return e.config
}

// SetConfig updates the configuration.
func (e *CSVExporter) SetConfig(config CSVConfig) {
	e.config = config
}
