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
	"fmt"

	"github.com/magpierre/fyne-datatable/datatable"
)

// ModelIterator implements RowIterator for a TableModel.
// It iterates over visible rows in the current view.
type ModelIterator struct {
	model       datatable.DataSource
	visibleRows []int // Indices of visible rows
	columnNames []string
	columnTypes []datatable.DataType
	currentRow  int
	err         error
}

// NewModelIterator creates an iterator from a DataSource and visible row indices.
func NewModelIterator(
	source datatable.DataSource,
	visibleRows []int,
) (*ModelIterator, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}

	// Get column information
	colCount := source.ColumnCount()
	columnNames := make([]string, colCount)
	columnTypes := make([]datatable.DataType, colCount)

	for i := 0; i < colCount; i++ {
		name, err := source.ColumnName(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get column name %d: %w", i, err)
		}
		columnNames[i] = name

		colType, err := source.ColumnType(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get column type %d: %w", i, err)
		}
		columnTypes[i] = colType
	}

	// If no visible rows specified, use all rows
	if visibleRows == nil {
		rowCount := source.RowCount()
		visibleRows = make([]int, rowCount)
		for i := 0; i < rowCount; i++ {
			visibleRows[i] = i
		}
	}

	return &ModelIterator{
		model:       source,
		visibleRows: visibleRows,
		columnNames: columnNames,
		columnTypes: columnTypes,
		currentRow:  -1, // Start before first row
	}, nil
}

// Next advances to the next row.
func (it *ModelIterator) Next() bool {
	if it.err != nil {
		return false
	}

	it.currentRow++
	return it.currentRow < len(it.visibleRows)
}

// Row returns the current row's values.
func (it *ModelIterator) Row() ([]datatable.Value, error) {
	if it.currentRow < 0 {
		return nil, fmt.Errorf("Next() not called yet")
	}

	if it.currentRow >= len(it.visibleRows) {
		return nil, fmt.Errorf("iterator exhausted")
	}

	// Get the original row index
	originalRowIdx := it.visibleRows[it.currentRow]

	// Get the row from the data source
	row, err := it.model.Row(originalRowIdx)
	if err != nil {
		it.err = err
		return nil, fmt.Errorf("failed to get row %d: %w", originalRowIdx, err)
	}

	return row, nil
}

// RowNumber returns the current row number (0-based in the visible rows).
func (it *ModelIterator) RowNumber() int {
	return it.currentRow
}

// TotalRows returns the total number of rows to iterate.
func (it *ModelIterator) TotalRows() int {
	return len(it.visibleRows)
}

// ColumnNames returns the column names.
func (it *ModelIterator) ColumnNames() []string {
	// Return a copy to prevent modification
	names := make([]string, len(it.columnNames))
	copy(names, it.columnNames)
	return names
}

// ColumnTypes returns the column data types.
func (it *ModelIterator) ColumnTypes() []datatable.DataType {
	// Return a copy to prevent modification
	types := make([]datatable.DataType, len(it.columnTypes))
	copy(types, it.columnTypes)
	return types
}

// Err returns any error encountered during iteration.
func (it *ModelIterator) Err() error {
	return it.err
}

// Reset resets the iterator to the beginning.
func (it *ModelIterator) Reset() {
	it.currentRow = -1
	it.err = nil
}
