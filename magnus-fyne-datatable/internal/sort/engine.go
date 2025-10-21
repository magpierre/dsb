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

// Package sort provides sorting functionality for table data.
package sort

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/magpierre/fyne-datatable/datatable"
)

// Engine sorts data indices based on column values.
// It is stateless - all methods are pure functions.
type Engine struct {
	// Stateless - no fields needed
}

// NewEngine creates a new SortEngine.
func NewEngine() *Engine {
	return &Engine{}
}

// SortSpec specifies how to sort a column.
type SortSpec struct {
	// Column is the index of the column to sort by.
	Column int

	// Direction specifies ascending or descending.
	Direction datatable.SortDirection

	// DataType helps with type-aware sorting.
	DataType datatable.DataType
}

// Sort sorts row indices based on the values in a specified column.
// The sort is stable - equal elements maintain their original order.
// Returns a new slice of sorted indices.
func (e *Engine) Sort(
	source datatable.DataSource,
	indices []int,
	spec SortSpec,
) ([]int, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}

	if spec.Direction == datatable.SortNone {
		// No sorting - return copy of original indices
		result := make([]int, len(indices))
		copy(result, indices)
		return result, nil
	}

	if spec.Column < 0 || spec.Column >= source.ColumnCount() {
		return nil, fmt.Errorf("%w: %d", datatable.ErrInvalidSortColumn, spec.Column)
	}

	// Create a copy of indices to sort
	result := make([]int, len(indices))
	copy(result, indices)

	// Get column type for type-aware sorting
	colType, err := source.ColumnType(spec.Column)
	if err != nil {
		colType = spec.DataType // Fall back to provided type
	}

	// Perform stable sort
	sort.SliceStable(result, func(i, j int) bool {
		rowI := result[i]
		rowJ := result[j]

		// Get cell values
		cellI, errI := source.Cell(rowI, spec.Column)
		cellJ, errJ := source.Cell(rowJ, spec.Column)

		// Handle errors - place error rows at end
		if errI != nil || errJ != nil {
			return errI == nil // Non-error rows come first
		}

		// Compare values
		cmp := compareValues(cellI, cellJ, colType)

		// Apply direction
		if spec.Direction == datatable.SortAscending {
			return cmp < 0
		}
		return cmp > 0 // Descending
	})

	return result, nil
}

// MultiSort sorts by multiple columns in order of precedence.
// The first SortSpec has highest priority.
func (e *Engine) MultiSort(
	source datatable.DataSource,
	indices []int,
	specs []SortSpec,
) ([]int, error) {
	if source == nil {
		return nil, datatable.ErrNoDataSource
	}

	if len(specs) == 0 {
		// No sorting - return copy of original indices
		result := make([]int, len(indices))
		copy(result, indices)
		return result, nil
	}

	// Validate all specs
	for i, spec := range specs {
		if spec.Column < 0 || spec.Column >= source.ColumnCount() {
			return nil, fmt.Errorf("spec %d: %w: %d", i, datatable.ErrInvalidSortColumn, spec.Column)
		}
	}

	// Create a copy of indices to sort
	result := make([]int, len(indices))
	copy(result, indices)

	// Get column types
	colTypes := make([]datatable.DataType, len(specs))
	for i, spec := range specs {
		colType, err := source.ColumnType(spec.Column)
		if err != nil {
			colType = spec.DataType // Fall back to provided type
		}
		colTypes[i] = colType
	}

	// Perform stable sort with multiple columns
	sort.SliceStable(result, func(i, j int) bool {
		rowI := result[i]
		rowJ := result[j]

		// Compare by each column in order
		for specIdx, spec := range specs {
			// Get cell values
			cellI, errI := source.Cell(rowI, spec.Column)
			cellJ, errJ := source.Cell(rowJ, spec.Column)

			// Handle errors
			if errI != nil || errJ != nil {
				return errI == nil
			}

			// Compare values
			cmp := compareValues(cellI, cellJ, colTypes[specIdx])

			if cmp != 0 {
				// Values differ - apply direction and return
				if spec.Direction == datatable.SortAscending {
					return cmp < 0
				}
				return cmp > 0
			}

			// Values equal - continue to next column
		}

		// All columns equal - maintain original order (stable sort)
		return false
	})

	return result, nil
}

// compareValues compares two Value objects based on their data type.
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareValues(a, b datatable.Value, dataType datatable.DataType) int {
	// Null handling - nulls sort to end
	if a.IsNull && b.IsNull {
		return 0
	}
	if a.IsNull {
		return 1 // Nulls sort after non-nulls
	}
	if b.IsNull {
		return -1
	}

	// Type-aware comparison
	switch dataType {
	case datatable.TypeInt, datatable.TypeFloat, datatable.TypeDecimal:
		return compareNumeric(a.Formatted, b.Formatted)

	case datatable.TypeDate, datatable.TypeTimestamp:
		return compareDateTime(a.Formatted, b.Formatted)

	case datatable.TypeBool:
		return compareBool(a.Formatted, b.Formatted)

	default:
		// String comparison (case-insensitive)
		return compareString(a.Formatted, b.Formatted)
	}
}

// compareNumeric compares two values as numbers.
func compareNumeric(a, b string) int {
	aNum, aErr := strconv.ParseFloat(strings.TrimSpace(a), 64)
	bNum, bErr := strconv.ParseFloat(strings.TrimSpace(b), 64)

	// If parsing fails, fall back to string comparison
	if aErr != nil || bErr != nil {
		return compareString(a, b)
	}

	if aNum < bNum {
		return -1
	}
	if aNum > bNum {
		return 1
	}
	return 0
}

// compareDateTime compares two values as dates/timestamps.
func compareDateTime(a, b string) int {
	// Try multiple common date formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC822,
	}

	var aTime, bTime time.Time
	var aErr, bErr error

	for _, format := range formats {
		aTime, aErr = time.Parse(format, a)
		if aErr == nil {
			break
		}
	}

	for _, format := range formats {
		bTime, bErr = time.Parse(format, b)
		if bErr == nil {
			break
		}
	}

	// If parsing fails, fall back to string comparison
	if aErr != nil || bErr != nil {
		return compareString(a, b)
	}

	if aTime.Before(bTime) {
		return -1
	}
	if aTime.After(bTime) {
		return 1
	}
	return 0
}

// compareBool compares two boolean values.
func compareBool(a, b string) int {
	aBool, aErr := strconv.ParseBool(strings.TrimSpace(a))
	bBool, bErr := strconv.ParseBool(strings.TrimSpace(b))

	// If parsing fails, fall back to string comparison
	if aErr != nil || bErr != nil {
		return compareString(a, b)
	}

	// false < true
	if !aBool && bBool {
		return -1
	}
	if aBool && !bBool {
		return 1
	}
	return 0
}

// compareString compares two strings (case-insensitive).
func compareString(a, b string) int {
	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)

	if aLower < bLower {
		return -1
	}
	if aLower > bLower {
		return 1
	}
	return 0
}
