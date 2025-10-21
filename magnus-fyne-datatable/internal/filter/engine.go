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

// Package filter provides filtering functionality for table data.
package filter

import (
	"fmt"

	"github.com/magpierre/fyne-datatable/datatable"
)

// Engine applies filters to data sources.
// It is stateless - all methods are pure functions.
type Engine struct {
	// Stateless - no fields needed
}

// NewEngine creates a new FilterEngine.
func NewEngine() *Engine {
	return &Engine{}
}

// Apply applies a single filter to a data source and returns the indices
// of rows that pass the filter.
func (e *Engine) Apply(
	source datatable.DataSource,
	filter datatable.Filter,
) ([]int, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}
	if filter == nil {
		return nil, datatable.ErrInvalidFilter
	}

	rowCount := source.RowCount()
	colCount := source.ColumnCount()

	// Build column names array
	columnNames := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		name, err := source.ColumnName(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get column name: %w", err)
		}
		columnNames[i] = name
	}

	// Apply filter to each row
	result := make([]int, 0, rowCount) // Pre-allocate for efficiency

	for rowIdx := 0; rowIdx < rowCount; rowIdx++ {
		row, err := source.Row(rowIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to get row %d: %w", rowIdx, err)
		}

		passes, err := filter.Evaluate(row, columnNames)
		if err != nil {
			return nil, fmt.Errorf("filter evaluation failed on row %d: %w", rowIdx, err)
		}

		if passes {
			result = append(result, rowIdx)
		}
	}

	return result, nil
}

// ApplyMultiple applies multiple filters with AND logic.
// Returns indices of rows that pass ALL filters.
func (e *Engine) ApplyMultiple(
	source datatable.DataSource,
	filters []datatable.Filter,
) ([]int, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}
	if len(filters) == 0 {
		// No filters - return all rows
		rowCount := source.RowCount()
		result := make([]int, rowCount)
		for i := range result {
			result[i] = i
		}
		return result, nil
	}

	// Use composite filter with AND logic
	composite := &CompositeFilter{
		Filters: filters,
		Logic:   LogicAND,
	}

	return e.Apply(source, composite)
}

// Validate checks if a filter is valid for the given data source.
// This can be used before applying a filter to catch errors early.
func (e *Engine) Validate(
	filter datatable.Filter,
	source datatable.DataSource,
) error {
	if filter == nil {
		return datatable.ErrInvalidFilter
	}
	if source == nil {
		return datatable.ErrNoDataSource
	}

	// For now, just check that we can evaluate the filter on the first row
	// More sophisticated validation could be added per filter type
	if source.RowCount() > 0 {
		colCount := source.ColumnCount()
		columnNames := make([]string, colCount)
		for i := 0; i < colCount; i++ {
			name, err := source.ColumnName(i)
			if err != nil {
				return err
			}
			columnNames[i] = name
		}

		row, err := source.Row(0)
		if err != nil {
			return err
		}

		_, err = filter.Evaluate(row, columnNames)
		return err
	}

	return nil
}

// ApplyToIndices applies a filter to a subset of rows (specified by indices).
// This is useful for applying filters after sorting or previous filtering.
func (e *Engine) ApplyToIndices(
	source datatable.DataSource,
	filter datatable.Filter,
	indices []int,
) ([]int, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}
	if filter == nil {
		return nil, datatable.ErrInvalidFilter
	}

	colCount := source.ColumnCount()

	// Build column names array
	columnNames := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		name, err := source.ColumnName(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get column name: %w", err)
		}
		columnNames[i] = name
	}

	// Apply filter to specified rows only
	result := make([]int, 0, len(indices))

	for _, rowIdx := range indices {
		row, err := source.Row(rowIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to get row %d: %w", rowIdx, err)
		}

		passes, err := filter.Evaluate(row, columnNames)
		if err != nil {
			return nil, fmt.Errorf("filter evaluation failed on row %d: %w", rowIdx, err)
		}

		if passes {
			result = append(result, rowIdx)
		}
	}

	return result, nil
}
