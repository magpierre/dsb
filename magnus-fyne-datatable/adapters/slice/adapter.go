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

// Package slice provides a DataSource adapter for Go slices.
package slice

import (
	"fmt"
	"sync"

	"github.com/magpierre/fyne-datatable/datatable"
)

// SliceDataSource implements DataSource for [][]interface{}.
// It automatically converts interface{} values to appropriate types.
type SliceDataSource struct {
	mu          sync.RWMutex
	data        [][]datatable.Value
	columnNames []string
	columnTypes []datatable.DataType
	metadata    datatable.Metadata
}

// NewFromInterfaces creates a DataSource from [][]interface{}.
// Column names must be provided. Types are inferred from data.
func NewFromInterfaces(data [][]interface{}, columnNames []string) (*SliceDataSource, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	if columnNames == nil || len(columnNames) == 0 {
		return nil, fmt.Errorf("column names cannot be empty")
	}

	// Validate column count consistency
	expectedCols := len(columnNames)
	for i, row := range data {
		if len(row) != expectedCols {
			return nil, fmt.Errorf("inconsistent column count at row %d: expected %d, got %d",
				i, expectedCols, len(row))
		}
	}

	// Convert to Value types
	valueData := make([][]datatable.Value, len(data))
	for i, row := range data {
		valueRow := make([]datatable.Value, len(row))
		for j, cell := range row {
			valueRow[j] = convertToValue(cell)
		}
		valueData[i] = valueRow
	}

	// Infer column types
	columnTypes := inferTypes(valueData, len(columnNames))

	return &SliceDataSource{
		data:        valueData,
		columnNames: columnNames,
		columnTypes: columnTypes,
		metadata:    make(datatable.Metadata),
	}, nil
}

// NewFromStrings creates a DataSource from [][]string.
// This is a convenience function for string data.
func NewFromStrings(data [][]string, columnNames []string) (*SliceDataSource, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	if columnNames == nil || len(columnNames) == 0 {
		return nil, fmt.Errorf("column names cannot be empty")
	}

	// Convert strings to interface{}
	interfaceData := make([][]interface{}, len(data))
	for i, row := range data {
		interfaceRow := make([]interface{}, len(row))
		for j, cell := range row {
			interfaceRow[j] = cell
		}
		interfaceData[i] = interfaceRow
	}

	return NewFromInterfaces(interfaceData, columnNames)
}

// NewFromMaps creates a DataSource from []map[string]interface{}.
// Column names are extracted from map keys (first map determines column order).
func NewFromMaps(data []map[string]interface{}) (*SliceDataSource, error) {
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	// Extract column names from first map
	var columnNames []string
	for key := range data[0] {
		columnNames = append(columnNames, key)
	}

	if len(columnNames) == 0 {
		return nil, fmt.Errorf("no columns found in data")
	}

	// Convert maps to rows
	rows := make([][]interface{}, len(data))
	for i, rowMap := range data {
		row := make([]interface{}, len(columnNames))
		for j, colName := range columnNames {
			row[j] = rowMap[colName]
		}
		rows[i] = row
	}

	return NewFromInterfaces(rows, columnNames)
}

// convertToValue converts an interface{} to a Value with type inference.
func convertToValue(v interface{}) datatable.Value {
	if v == nil {
		return datatable.NewNullValue(datatable.TypeString)
	}

	switch val := v.(type) {
	case string:
		return datatable.NewValue(val, datatable.TypeString)
	case int:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case int32:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case int64:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case uint:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case uint32:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case uint64:
		return datatable.NewValue(fmt.Sprintf("%d", val), datatable.TypeInt)
	case float32:
		return datatable.NewValue(fmt.Sprintf("%f", val), datatable.TypeFloat)
	case float64:
		return datatable.NewValue(fmt.Sprintf("%f", val), datatable.TypeFloat)
	case bool:
		return datatable.NewValue(fmt.Sprintf("%t", val), datatable.TypeBool)
	default:
		// Fallback to string representation
		return datatable.NewValue(fmt.Sprintf("%v", val), datatable.TypeString)
	}
}

// inferTypes infers column types from the data.
func inferTypes(data [][]datatable.Value, numCols int) []datatable.DataType {
	types := make([]datatable.DataType, numCols)

	// Count types for each column
	for col := 0; col < numCols; col++ {
		typeCount := make(map[datatable.DataType]int)

		for row := 0; row < len(data); row++ {
			if col >= len(data[row]) {
				continue
			}

			value := data[row][col]
			if !value.IsNull {
				typeCount[value.Type]++
			}
		}

		// Use most common type, defaulting to String
		maxCount := 0
		types[col] = datatable.TypeString

		for dtype, count := range typeCount {
			if count > maxCount {
				maxCount = count
				types[col] = dtype
			}
		}
	}

	return types
}

// DataSource interface implementation

func (ds *SliceDataSource) RowCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.data)
}

func (ds *SliceDataSource) ColumnCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.columnNames)
}

func (ds *SliceDataSource) ColumnName(col int) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if col < 0 || col >= len(ds.columnNames) {
		return "", datatable.ErrInvalidColumn
	}

	return ds.columnNames[col], nil
}

func (ds *SliceDataSource) ColumnType(col int) (datatable.DataType, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if col < 0 || col >= len(ds.columnTypes) {
		return datatable.TypeString, datatable.ErrInvalidColumn
	}

	return ds.columnTypes[col], nil
}

func (ds *SliceDataSource) Cell(row, col int) (datatable.Value, error) {
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

func (ds *SliceDataSource) Row(row int) ([]datatable.Value, error) {
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

func (ds *SliceDataSource) Metadata() datatable.Metadata {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.metadata
}
