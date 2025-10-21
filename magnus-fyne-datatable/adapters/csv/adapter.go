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

// Package csv provides a DataSource adapter for CSV files.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/magpierre/fyne-datatable/datatable"
)

// Config configures CSV file loading.
type Config struct {
	// Delimiter is the field delimiter (default: comma)
	Delimiter rune

	// HasHeaders indicates if the first row contains column names
	HasHeaders bool

	// TrimSpace removes leading/trailing whitespace from fields
	TrimSpace bool

	// Comment character to ignore lines (0 = no comments)
	Comment rune

	// LazyQuotes allows lazy quote parsing
	LazyQuotes bool
}

// DefaultConfig returns the default CSV configuration.
func DefaultConfig() Config {
	return Config{
		Delimiter:  ',',
		HasHeaders: true,
		TrimSpace:  true,
		Comment:    0,
		LazyQuotes: false,
	}
}

// CSVDataSource implements DataSource for CSV files.
type CSVDataSource struct {
	mu          sync.RWMutex
	data        [][]datatable.Value
	columnNames []string
	columnTypes []datatable.DataType
	metadata    datatable.Metadata
}

// NewFromFile loads a CSV file and creates a DataSource.
func NewFromFile(filename string, config Config) (*CSVDataSource, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return NewFromReader(file, config)
}

// NewFromReader loads CSV data from an io.Reader.
func NewFromReader(reader io.Reader, config Config) (*CSVDataSource, error) {
	// Create CSV reader
	csvReader := csv.NewReader(reader)
	csvReader.Comma = config.Delimiter
	csvReader.Comment = config.Comment
	csvReader.TrimLeadingSpace = config.TrimSpace
	csvReader.LazyQuotes = config.LazyQuotes

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Extract headers if present
	var columnNames []string
	var dataStart int

	if config.HasHeaders {
		if len(records) < 1 {
			return nil, fmt.Errorf("CSV has headers but no data rows")
		}
		columnNames = records[0]
		dataStart = 1
	} else {
		// Generate column names: Col1, Col2, ...
		if len(records) > 0 {
			numCols := len(records[0])
			columnNames = make([]string, numCols)
			for i := 0; i < numCols; i++ {
				columnNames[i] = fmt.Sprintf("Col%d", i+1)
			}
		}
		dataStart = 0
	}

	// Validate column count consistency
	expectedCols := len(columnNames)
	for i := dataStart; i < len(records); i++ {
		if len(records[i]) != expectedCols {
			return nil, fmt.Errorf("inconsistent column count at row %d: expected %d, got %d",
				i, expectedCols, len(records[i]))
		}
	}

	// Convert data to Value types
	dataRows := make([][]datatable.Value, len(records)-dataStart)
	for i := dataStart; i < len(records); i++ {
		row := records[i]
		valueRow := make([]datatable.Value, len(row))
		for j, cell := range row {
			// Apply trimming if configured
			if config.TrimSpace {
				cell = trimSpace(cell)
			}
			valueRow[j] = datatable.NewValue(cell, datatable.TypeString)
		}
		dataRows[i-dataStart] = valueRow
	}

	// Infer column types from data
	columnTypes := inferColumnTypes(dataRows, len(columnNames))

	// Update Value types based on inferred types
	for i := range dataRows {
		for j := range dataRows[i] {
			dataRows[i][j].Type = columnTypes[j]
		}
	}

	return &CSVDataSource{
		data:        dataRows,
		columnNames: columnNames,
		columnTypes: columnTypes,
		metadata:    make(datatable.Metadata),
	}, nil
}

// inferColumnTypes attempts to infer data types from the data.
func inferColumnTypes(data [][]datatable.Value, numCols int) []datatable.DataType {
	types := make([]datatable.DataType, numCols)

	// Initialize all as string
	for i := range types {
		types[i] = datatable.TypeString
	}

	// Sample first few rows to infer types
	sampleSize := 100
	if len(data) < sampleSize {
		sampleSize = len(data)
	}

	for col := 0; col < numCols; col++ {
		allInts := true
		allFloats := true
		allBools := true

		for row := 0; row < sampleSize; row++ {
			if col >= len(data[row]) {
				continue
			}

			value := data[row][col].Formatted

			// Check if empty/null
			if value == "" {
				continue
			}

			// Try int
			if allInts {
				if !isInt(value) {
					allInts = false
				}
			}

			// Try float
			if allFloats {
				if !isFloat(value) {
					allFloats = false
				}
			}

			// Try bool
			if allBools {
				if !isBool(value) {
					allBools = false
				}
			}
		}

		// Assign type based on what passed
		if allBools {
			types[col] = datatable.TypeBool
		} else if allInts {
			types[col] = datatable.TypeInt
		} else if allFloats {
			types[col] = datatable.TypeFloat
		} else {
			types[col] = datatable.TypeString
		}
	}

	return types
}

// Helper functions for type inference
func isInt(s string) bool {
	if s == "" {
		return false
	}
	// Simple check for integer
	for i, c := range s {
		if i == 0 && (c == '-' || c == '+') {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isFloat(s string) bool {
	if s == "" {
		return false
	}
	dotSeen := false
	for i, c := range s {
		if i == 0 && (c == '-' || c == '+') {
			continue
		}
		if c == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isBool(s string) bool {
	s = trimSpace(s)
	lower := toLower(s)
	return lower == "true" || lower == "false" || lower == "1" || lower == "0" ||
		lower == "yes" || lower == "no" || lower == "y" || lower == "n"
}

func trimSpace(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)

	for start < end && isSpace(s[start]) {
		start++
	}

	for end > start && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// DataSource interface implementation

func (ds *CSVDataSource) RowCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.data)
}

func (ds *CSVDataSource) ColumnCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.columnNames)
}

func (ds *CSVDataSource) ColumnName(col int) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if col < 0 || col >= len(ds.columnNames) {
		return "", datatable.ErrInvalidColumn
	}

	return ds.columnNames[col], nil
}

func (ds *CSVDataSource) ColumnType(col int) (datatable.DataType, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if col < 0 || col >= len(ds.columnTypes) {
		return datatable.TypeString, datatable.ErrInvalidColumn
	}

	return ds.columnTypes[col], nil
}

func (ds *CSVDataSource) Cell(row, col int) (datatable.Value, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if row < 0 || row >= len(ds.data) {
		return datatable.Value{}, datatable.ErrInvalidRow
	}

	if col < 0 || col >= len(ds.columnNames) {
		return datatable.Value{}, datatable.ErrInvalidColumn
	}

	return ds.data[row][col], nil
}

func (ds *CSVDataSource) Row(row int) ([]datatable.Value, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if row < 0 || row >= len(ds.data) {
		return nil, datatable.ErrInvalidRow
	}

	// Return a copy
	result := make([]datatable.Value, len(ds.data[row]))
	copy(result, ds.data[row])
	return result, nil
}

func (ds *CSVDataSource) Metadata() datatable.Metadata {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.metadata
}
