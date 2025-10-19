// Package memory provides an in-memory implementation of DataSource.
package memory

import (
	"fmt"
	"sync"

	"github.com/magpierre/fyne-datatable/datatable"
)

// MemoryDataSource is a thread-safe in-memory implementation of DataSource.
// It's suitable for small to medium datasets that fit in memory.
type MemoryDataSource struct {
	mu          sync.RWMutex
	data        [][]datatable.Value
	columnNames []string
	columnTypes []datatable.DataType
	metadata    datatable.Metadata
}

// NewDataSource creates a new in-memory data source from string data.
// The data is converted to Values with type inference.
func NewDataSource(data [][]string, columnNames []string) (*MemoryDataSource, error) {
	if len(columnNames) == 0 {
		return nil, fmt.Errorf("%w: no columns provided", datatable.ErrEmptyData)
	}

	// Infer types from first row of data
	columnTypes := make([]datatable.DataType, len(columnNames))
	for i := range columnTypes {
		columnTypes[i] = datatable.TypeString // Default to string
	}

	// Convert string data to Values
	values := make([][]datatable.Value, len(data))
	for i, row := range data {
		if len(row) != len(columnNames) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", i, len(row), len(columnNames))
		}
		values[i] = make([]datatable.Value, len(row))
		for j, cell := range row {
			values[i][j] = datatable.NewValue(cell, datatable.TypeString)
		}
	}

	return &MemoryDataSource{
		data:        values,
		columnNames: columnNames,
		columnTypes: columnTypes,
		metadata:    make(datatable.Metadata),
	}, nil
}

// NewDataSourceFromValues creates a new in-memory data source from typed Values.
func NewDataSourceFromValues(data [][]datatable.Value, columnNames []string, columnTypes []datatable.DataType) (*MemoryDataSource, error) {
	if len(columnNames) == 0 {
		return nil, fmt.Errorf("%w: no columns provided", datatable.ErrEmptyData)
	}

	if len(columnNames) != len(columnTypes) {
		return nil, fmt.Errorf("column names (%d) and types (%d) length mismatch", len(columnNames), len(columnTypes))
	}

	// Validate data dimensions
	for i, row := range data {
		if len(row) != len(columnNames) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", i, len(row), len(columnNames))
		}
	}

	return &MemoryDataSource{
		data:        data,
		columnNames: columnNames,
		columnTypes: columnTypes,
		metadata:    make(datatable.Metadata),
	}, nil
}

// RowCount returns the total number of rows.
func (m *MemoryDataSource) RowCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// ColumnCount returns the total number of columns.
func (m *MemoryDataSource) ColumnCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.columnNames)
}

// ColumnName returns the name of the column at the given index.
func (m *MemoryDataSource) ColumnName(col int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if col < 0 || col >= len(m.columnNames) {
		return "", fmt.Errorf("%w: %d (valid range: 0-%d)", datatable.ErrInvalidColumn, col, len(m.columnNames)-1)
	}

	return m.columnNames[col], nil
}

// ColumnType returns the data type of the column at the given index.
func (m *MemoryDataSource) ColumnType(col int) (datatable.DataType, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if col < 0 || col >= len(m.columnTypes) {
		return datatable.TypeString, fmt.Errorf("%w: %d (valid range: 0-%d)", datatable.ErrInvalidColumn, col, len(m.columnTypes)-1)
	}

	return m.columnTypes[col], nil
}

// Cell returns the value at the specified row and column.
func (m *MemoryDataSource) Cell(row, col int) (datatable.Value, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if row < 0 || row >= len(m.data) {
		return datatable.Value{}, fmt.Errorf("%w: %d (valid range: 0-%d)", datatable.ErrInvalidRow, row, len(m.data)-1)
	}

	if col < 0 || col >= len(m.columnNames) {
		return datatable.Value{}, fmt.Errorf("%w: %d (valid range: 0-%d)", datatable.ErrInvalidColumn, col, len(m.columnNames)-1)
	}

	return m.data[row][col], nil
}

// Row returns all values for the specified row.
func (m *MemoryDataSource) Row(row int) ([]datatable.Value, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if row < 0 || row >= len(m.data) {
		return nil, fmt.Errorf("%w: %d (valid range: 0-%d)", datatable.ErrInvalidRow, row, len(m.data)-1)
	}

	// Return a copy to prevent modification
	result := make([]datatable.Value, len(m.data[row]))
	copy(result, m.data[row])
	return result, nil
}

// Metadata returns metadata about the data source.
func (m *MemoryDataSource) Metadata() datatable.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent modification
	result := make(datatable.Metadata)
	for k, v := range m.metadata {
		result[k] = v
	}
	return result
}

// SetMetadata sets a metadata value (not part of DataSource interface).
func (m *MemoryDataSource) SetMetadata(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = value
}
