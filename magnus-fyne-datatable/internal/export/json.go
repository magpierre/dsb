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
	"encoding/json"
	"fmt"
	"io"
)

// JSONConfig configures JSON export options.
type JSONConfig struct {
	// PrettyPrint enables formatted output with indentation
	PrettyPrint bool

	// Indent specifies the indentation string (used if PrettyPrint is true)
	Indent string
}

// DefaultJSONConfig returns the default JSON configuration.
func DefaultJSONConfig() JSONConfig {
	return JSONConfig{
		PrettyPrint: false,
		Indent:      "  ",
	}
}

// JSONExporter exports data in JSON format.
// Output is an array of objects, where each object represents a row.
type JSONExporter struct {
	config JSONConfig
}

// NewJSONExporter creates a new JSON exporter with default configuration.
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{
		config: DefaultJSONConfig(),
	}
}

// NewJSONExporterWithConfig creates a JSON exporter with custom configuration.
func NewJSONExporterWithConfig(config JSONConfig) *JSONExporter {
	return &JSONExporter{
		config: config,
	}
}

// Export writes data in JSON format as an array of objects.
// Each row becomes an object with column names as keys.
func (e *JSONExporter) Export(
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

	columnNames := iterator.ColumnNames()
	totalRows := iterator.TotalRows()
	rowCount := 0

	// Write opening bracket
	if _, err := writer.Write([]byte("[")); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	// Track if we need to write a comma before the next object
	needsComma := false

	for iterator.Next() {
		row, err := iterator.Row()
		if err != nil {
			return rowCount, fmt.Errorf("failed to get row %d: %w", rowCount, err)
		}

		// Write comma before object (except for first object)
		if needsComma {
			if _, err := writer.Write([]byte(",")); err != nil {
				return rowCount, fmt.Errorf("failed to write comma: %w", err)
			}
		}

		// Add newline and indentation if pretty printing
		if e.config.PrettyPrint {
			if _, err := writer.Write([]byte("\n" + e.config.Indent)); err != nil {
				return rowCount, fmt.Errorf("failed to write newline: %w", err)
			}
		}

		// Build object as map
		obj := make(map[string]interface{}, len(columnNames))
		for i, val := range row {
			if i >= len(columnNames) {
				// Safety check: skip extra columns
				break
			}

			colName := columnNames[i]
			if val.IsNull {
				obj[colName] = nil
			} else {
				// Use Raw value if available, otherwise Formatted string
				if val.Raw != nil {
					obj[colName] = val.Raw
				} else {
					obj[colName] = val.Formatted
				}
			}
		}

		// Marshal object to JSON
		var jsonBytes []byte
		if e.config.PrettyPrint {
			// For pretty print, we use indentation for nested objects
			// but not for the array wrapping (we handle that manually)
			jsonBytes, err = json.Marshal(obj)
		} else {
			jsonBytes, err = json.Marshal(obj)
		}

		if err != nil {
			return rowCount, fmt.Errorf("failed to marshal row %d: %w", rowCount, err)
		}

		// Write the JSON object
		if _, err := writer.Write(jsonBytes); err != nil {
			return rowCount, fmt.Errorf("failed to write row %d: %w", rowCount, err)
		}

		needsComma = true
		rowCount++

		// Report progress if callback provided
		if progress != nil {
			if !progress(rowCount, totalRows) {
				// User cancelled - still write valid JSON
				if e.config.PrettyPrint {
					writer.Write([]byte("\n"))
				}
				writer.Write([]byte("]"))
				return rowCount, fmt.Errorf("export cancelled by user")
			}
		}
	}

	// Check for iteration errors
	if err := iterator.Err(); err != nil {
		return rowCount, fmt.Errorf("iterator error: %w", err)
	}

	// Write closing bracket
	if e.config.PrettyPrint {
		if _, err := writer.Write([]byte("\n")); err != nil {
			return rowCount, fmt.Errorf("failed to write newline: %w", err)
		}
	}

	if _, err := writer.Write([]byte("]")); err != nil {
		return rowCount, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Add final newline for pretty print
	if e.config.PrettyPrint {
		if _, err := writer.Write([]byte("\n")); err != nil {
			return rowCount, fmt.Errorf("failed to write final newline: %w", err)
		}
	}

	return rowCount, nil
}

// FileExtension returns "json".
func (e *JSONExporter) FileExtension() string {
	return "json"
}

// MimeType returns the JSON MIME type.
func (e *JSONExporter) MimeType() string {
	return "application/json"
}

// Description returns a human-readable description.
func (e *JSONExporter) Description() string {
	if e.config.PrettyPrint {
		return "JSON (Pretty Printed)"
	}
	return "JSON (Compact)"
}

// GetConfig returns the current configuration.
func (e *JSONExporter) GetConfig() JSONConfig {
	return e.config
}

// SetConfig updates the configuration.
func (e *JSONExporter) SetConfig(config JSONConfig) {
	e.config = config
}
