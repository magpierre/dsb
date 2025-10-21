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

package datatable

import (
	"fmt"
	"sync"
)

// TableModel manages the state of table data and coordinates transformations.
// It provides view management (filtering, sorting, column visibility) without
// modifying the underlying data source.
type TableModel struct {
	// Dependencies (immutable after creation)
	source DataSource

	// Thread safety
	mu sync.RWMutex

	// Original data dimensions (cached for performance)
	originalRows int
	originalCols int

	// View state (mutable, protected by mu)
	visibleRows []int // Indices of visible rows in original data
	visibleCols []int // Indices of visible columns

	// Sort state
	sortState SortState

	// Filter state
	activeFilters []Filter
	filterMask    []bool // Quick lookup: is row i visible after filtering?
}

// NewTableModel creates a new TableModel from a DataSource.
// Returns ErrNoDataSource if source is nil.
func NewTableModel(source DataSource) (*TableModel, error) {
	if source == nil {
		return nil, ErrNoDataSource
	}

	rowCount := source.RowCount()
	colCount := source.ColumnCount()

	// Initialize visible rows and columns to show everything
	visibleRows := make([]int, rowCount)
	for i := range visibleRows {
		visibleRows[i] = i
	}

	visibleCols := make([]int, colCount)
	for i := range visibleCols {
		visibleCols[i] = i
	}

	// Initialize filter mask (all rows visible)
	filterMask := make([]bool, rowCount)
	for i := range filterMask {
		filterMask[i] = true
	}

	return &TableModel{
		source:        source,
		originalRows:  rowCount,
		originalCols:  colCount,
		visibleRows:   visibleRows,
		visibleCols:   visibleCols,
		sortState:     SortState{Column: -1, Direction: SortNone},
		activeFilters: make([]Filter, 0),
		filterMask:    filterMask,
	}, nil
}

// --- View Queries (Read-only, thread-safe) ---

// VisibleRowCount returns the number of currently visible rows.
func (m *TableModel) VisibleRowCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.visibleRows)
}

// VisibleColumnCount returns the number of currently visible columns.
func (m *TableModel) VisibleColumnCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.visibleCols)
}

// OriginalRowCount returns the total number of rows in the data source.
func (m *TableModel) OriginalRowCount() int {
	return m.originalRows
}

// OriginalColumnCount returns the total number of columns in the data source.
func (m *TableModel) OriginalColumnCount() int {
	return m.originalCols
}

// VisibleCell returns the value at the specified visible row and column.
// The indices are relative to the current view (after filtering/sorting).
// Returns ErrInvalidRow if row is out of visible range.
// Returns ErrInvalidColumn if col is out of visible range.
func (m *TableModel) VisibleCell(row, col int) (Value, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if row < 0 || row >= len(m.visibleRows) {
		return Value{}, fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidRow, row, len(m.visibleRows)-1)
	}

	if col < 0 || col >= len(m.visibleCols) {
		return Value{}, fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidColumn, col, len(m.visibleCols)-1)
	}

	// Map to original indices
	originalRow := m.visibleRows[row]
	originalCol := m.visibleCols[col]

	return m.source.Cell(originalRow, originalCol)
}

// VisibleRow returns all values for the specified visible row.
// Returns ErrInvalidRow if row is out of visible range.
func (m *TableModel) VisibleRow(row int) ([]Value, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if row < 0 || row >= len(m.visibleRows) {
		return nil, fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidRow, row, len(m.visibleRows)-1)
	}

	originalRow := m.visibleRows[row]

	// Get full row from source
	fullRow, err := m.source.Row(originalRow)
	if err != nil {
		return nil, err
	}

	// Filter to visible columns
	result := make([]Value, len(m.visibleCols))
	for i, colIdx := range m.visibleCols {
		result[i] = fullRow[colIdx]
	}

	return result, nil
}

// VisibleColumnName returns the name of the specified visible column.
// Returns ErrInvalidColumn if col is out of visible range.
func (m *TableModel) VisibleColumnName(col int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if col < 0 || col >= len(m.visibleCols) {
		return "", fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidColumn, col, len(m.visibleCols)-1)
	}

	originalCol := m.visibleCols[col]
	return m.source.ColumnName(originalCol)
}

// VisibleColumnType returns the data type of the specified visible column.
// Returns ErrInvalidColumn if col is out of visible range.
func (m *TableModel) VisibleColumnType(col int) (DataType, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if col < 0 || col >= len(m.visibleCols) {
		return TypeString, fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidColumn, col, len(m.visibleCols)-1)
	}

	originalCol := m.visibleCols[col]
	return m.source.ColumnType(originalCol)
}

// --- State Queries ---

// GetSortState returns the current sort state.
func (m *TableModel) GetSortState() SortState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortState
}

// IsSorted returns true if the table is currently sorted.
func (m *TableModel) IsSorted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortState.IsSorted()
}

// IsFiltered returns true if any filters are active or columns are hidden.
func (m *TableModel) IsFiltered() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.activeFilters) > 0 || len(m.visibleCols) != m.originalCols
}

// --- State Mutations (validated, return errors) ---

// SetVisibleColumns sets which columns are visible.
// Columns are specified by their original indices.
// Returns ErrInvalidColumn if any column index is out of range.
func (m *TableModel) SetVisibleColumns(cols []int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate all indices
	for _, col := range cols {
		if col < 0 || col >= m.originalCols {
			return fmt.Errorf("%w: %d (valid range: 0-%d)", ErrInvalidColumn, col, m.originalCols-1)
		}
	}

	// Check for duplicates
	seen := make(map[int]bool)
	for _, col := range cols {
		if seen[col] {
			return fmt.Errorf("%w: duplicate column %d", ErrInvalidColumn, col)
		}
		seen[col] = true
	}

	// If we're currently sorted by a column, check if it will still be visible
	// We need to do this BEFORE updating m.visibleCols
	sortedOriginalCol := -1
	if m.sortState.IsSorted() && m.sortState.Column >= 0 && m.sortState.Column < len(m.visibleCols) {
		sortedOriginalCol = m.visibleCols[m.sortState.Column]
	}

	// Update visible columns
	m.visibleCols = make([]int, len(cols))
	copy(m.visibleCols, cols)

	// Check if the sorted column is still visible
	if sortedOriginalCol >= 0 {
		colStillVisible := false
		for _, col := range cols {
			if col == sortedOriginalCol {
				colStillVisible = true
				break
			}
		}
		if !colStillVisible {
			m.sortState = SortState{Column: -1, Direction: SortNone}
		}
	}

	return nil
}

// ResetVisibleColumns makes all columns visible.
func (m *TableModel) ResetVisibleColumns() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.visibleCols = make([]int, m.originalCols)
	for i := range m.visibleCols {
		m.visibleCols[i] = i
	}

	return nil
}

// ClearSort removes any active sorting, returning data to filtered order.
func (m *TableModel) ClearSort() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sortState = SortState{Column: -1, Direction: SortNone}

	// Reset visible rows to filtered order
	m.rebuildVisibleRows()

	return nil
}

// rebuildVisibleRows updates visibleRows based on filterMask.
// Must be called with lock held.
func (m *TableModel) rebuildVisibleRows() {
	newVisibleRows := make([]int, 0, m.originalRows)
	for i, visible := range m.filterMask {
		if visible {
			newVisibleRows = append(newVisibleRows, i)
		}
	}
	m.visibleRows = newVisibleRows
}

// GetDataSource returns the underlying data source (read-only).
func (m *TableModel) GetDataSource() DataSource {
	return m.source
}

// GetVisibleColumnIndices returns a copy of the current visible column indices.
func (m *TableModel) GetVisibleColumnIndices() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]int, len(m.visibleCols))
	copy(result, m.visibleCols)
	return result
}

// SetFilter applies a filter to the table, updating visible rows.
// The filter is evaluated against the original data source.
// Previous filters are replaced by the new filter.
// Pass nil to clear all filters.
func (m *TableModel) SetFilter(filter Filter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if filter == nil {
		// Clear filter
		m.activeFilters = make([]Filter, 0)
		for i := range m.filterMask {
			m.filterMask[i] = true
		}
		m.rebuildVisibleRows()
		return nil
	}

	// Apply filter to all rows
	columnNames := make([]string, m.originalCols)
	for i := 0; i < m.originalCols; i++ {
		name, err := m.source.ColumnName(i)
		if err != nil {
			return fmt.Errorf("failed to get column name %d: %w", i, err)
		}
		columnNames[i] = name
	}

	// Evaluate filter for each row
	for i := 0; i < m.originalRows; i++ {
		row, err := m.source.Row(i)
		if err != nil {
			return fmt.Errorf("failed to get row %d: %w", i, err)
		}

		passes, err := filter.Evaluate(row, columnNames)
		if err != nil {
			return fmt.Errorf("filter evaluation failed for row %d: %w", i, err)
		}

		m.filterMask[i] = passes
	}

	// Update active filters
	m.activeFilters = []Filter{filter}

	// Rebuild visible rows
	m.rebuildVisibleRows()

	// If we have a sort state, we need to re-sort the filtered rows
	if m.sortState.IsSorted() {
		// Note: This requires access to sort engine, which we'll handle
		// by having the caller re-apply sort after filter
		// For now, clear sort state to maintain consistency
		m.sortState = SortState{Column: -1, Direction: SortNone}
	}

	return nil
}

// SetSort applies sorting to the currently visible (filtered) rows.
// The column parameter is the visible column index (not original).
// Returns ErrInvalidColumn if column is out of visible range.
func (m *TableModel) SetSort(column int, direction SortDirection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if column < 0 || column >= len(m.visibleCols) {
		return fmt.Errorf("%w: %d (visible range: 0-%d)", ErrInvalidColumn, column, len(m.visibleCols)-1)
	}

	if direction == SortNone {
		// Clear sort
		m.sortState = SortState{Column: -1, Direction: SortNone}
		m.rebuildVisibleRows()
		return nil
	}

	// Update sort state
	m.sortState = SortState{
		Column:    column,
		Direction: direction,
	}

	// Note: The actual sorting is deferred to the sort engine
	// This method just updates the state
	// The caller (UI layer) should use the sort engine to get sorted indices
	// and then update visibleRows accordingly

	return nil
}

// ApplySortedIndices updates the visible rows with pre-sorted indices.
// This is called by the sort engine integration layer.
// The indices must be valid row indices from the current visibleRows.
func (m *TableModel) ApplySortedIndices(sortedIndices []int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate that sortedIndices is same length as current visibleRows
	if len(sortedIndices) != len(m.visibleRows) {
		return fmt.Errorf("sorted indices length %d does not match visible rows length %d",
			len(sortedIndices), len(m.visibleRows))
	}

	// Validate that all indices are valid
	for _, idx := range sortedIndices {
		if idx < 0 || idx >= m.originalRows {
			return fmt.Errorf("%w: %d", ErrInvalidRow, idx)
		}
	}

	// Update visible rows
	m.visibleRows = make([]int, len(sortedIndices))
	copy(m.visibleRows, sortedIndices)

	return nil
}

// GetActiveFilters returns a copy of the currently active filters.
func (m *TableModel) GetActiveFilters() []Filter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Filter, len(m.activeFilters))
	copy(result, m.activeFilters)
	return result
}

// GetVisibleRowIndices returns a copy of the current visible row indices.
// These are the original row indices after applying filters and sorting.
// Useful for export operations that need to iterate over visible rows.
func (m *TableModel) GetVisibleRowIndices() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]int, len(m.visibleRows))
	copy(result, m.visibleRows)
	return result
}

// Filter interface for extensibility (to be implemented in filter package)
type Filter interface {
	// Evaluate returns true if the row passes the filter
	Evaluate(row []Value, columnNames []string) (bool, error)

	// Description returns a human-readable description of the filter
	Description() string
}
