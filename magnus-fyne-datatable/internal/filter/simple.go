package filter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/magpierre/fyne-datatable/datatable"
)

// CompareOp represents a comparison operator.
type CompareOp int

const (
	// OpEqual checks for equality.
	OpEqual CompareOp = iota
	// OpNotEqual checks for inequality.
	OpNotEqual
	// OpGreaterThan checks if left > right.
	OpGreaterThan
	// OpLessThan checks if left < right.
	OpLessThan
	// OpGreaterOrEqual checks if left >= right.
	OpGreaterOrEqual
	// OpLessOrEqual checks if left <= right.
	OpLessOrEqual
	// OpContains checks if string contains substring.
	OpContains
	// OpStartsWith checks if string starts with prefix.
	OpStartsWith
	// OpEndsWith checks if string ends with suffix.
	OpEndsWith
)

// String returns the string representation of a CompareOp.
func (op CompareOp) String() string {
	switch op {
	case OpEqual:
		return "="
	case OpNotEqual:
		return "!="
	case OpGreaterThan:
		return ">"
	case OpLessThan:
		return "<"
	case OpGreaterOrEqual:
		return ">="
	case OpLessOrEqual:
		return "<="
	case OpContains:
		return "contains"
	case OpStartsWith:
		return "starts_with"
	case OpEndsWith:
		return "ends_with"
	default:
		return fmt.Sprintf("unknown(%d)", op)
	}
}

// SimpleFilter performs a comparison on a single column.
type SimpleFilter struct {
	// Column is the name of the column to filter on.
	Column string

	// Operator is the comparison operator.
	Operator CompareOp

	// Value is the value to compare against.
	Value interface{}
}

// Evaluate implements the Filter interface.
func (f *SimpleFilter) Evaluate(row []datatable.Value, columnNames []string) (bool, error) {
	// Find the column index
	colIdx := -1
	for i, name := range columnNames {
		if name == f.Column {
			colIdx = i
			break
		}
	}

	if colIdx < 0 {
		return false, fmt.Errorf("%w: %s", datatable.ErrColumnNotFound, f.Column)
	}

	if colIdx >= len(row) {
		return false, fmt.Errorf("%w: %d", datatable.ErrInvalidColumn, colIdx)
	}

	cellValue := row[colIdx]

	// Handle null values
	if cellValue.IsNull {
		return false, nil // Nulls don't match any comparison
	}

	// Perform comparison based on operator
	return f.compare(cellValue, f.Value, f.Operator)
}

// Description implements the Filter interface.
func (f *SimpleFilter) Description() string {
	return fmt.Sprintf("%s %s %v", f.Column, f.Operator, f.Value)
}

// compare performs the actual comparison based on the operator and value types.
func (f *SimpleFilter) compare(cellValue datatable.Value, filterValue interface{}, op CompareOp) (bool, error) {
	// Handle string operations
	if op == OpContains || op == OpStartsWith || op == OpEndsWith {
		cellStr := cellValue.Formatted
		filterStr := fmt.Sprintf("%v", filterValue)

		switch op {
		case OpContains:
			return strings.Contains(strings.ToLower(cellStr), strings.ToLower(filterStr)), nil
		case OpStartsWith:
			return strings.HasPrefix(strings.ToLower(cellStr), strings.ToLower(filterStr)), nil
		case OpEndsWith:
			return strings.HasSuffix(strings.ToLower(cellStr), strings.ToLower(filterStr)), nil
		}
	}

	// For numeric comparisons, try to parse as numbers
	cellNum, cellIsNum := parseNumber(cellValue.Formatted)
	filterNum, filterIsNum := parseNumber(fmt.Sprintf("%v", filterValue))

	if cellIsNum && filterIsNum {
		return compareNumbers(cellNum, filterNum, op)
	}

	// Fall back to string comparison
	cellStr := cellValue.Formatted
	filterStr := fmt.Sprintf("%v", filterValue)

	switch op {
	case OpEqual:
		return strings.EqualFold(cellStr, filterStr), nil
	case OpNotEqual:
		return !strings.EqualFold(cellStr, filterStr), nil
	case OpGreaterThan:
		return cellStr > filterStr, nil
	case OpLessThan:
		return cellStr < filterStr, nil
	case OpGreaterOrEqual:
		return cellStr >= filterStr, nil
	case OpLessOrEqual:
		return cellStr <= filterStr, nil
	default:
		return false, fmt.Errorf("%w: unsupported operator %s", datatable.ErrInvalidFilter, op)
	}
}

// parseNumber attempts to parse a string as a float64.
func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

// compareNumbers compares two numbers using the given operator.
func compareNumbers(a, b float64, op CompareOp) (bool, error) {
	switch op {
	case OpEqual:
		return a == b, nil
	case OpNotEqual:
		return a != b, nil
	case OpGreaterThan:
		return a > b, nil
	case OpLessThan:
		return a < b, nil
	case OpGreaterOrEqual:
		return a >= b, nil
	case OpLessOrEqual:
		return a <= b, nil
	default:
		return false, fmt.Errorf("%w: numeric comparison not supported for operator %s", datatable.ErrInvalidFilter, op)
	}
}
