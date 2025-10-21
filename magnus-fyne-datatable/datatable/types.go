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

// Package datatable provides a reusable data table widget for Fyne applications.
package datatable

import "fmt"

// DataType represents the type of data in a column.
type DataType int

const (
	// TypeString represents string data.
	TypeString DataType = iota
	// TypeInt represents integer data (any size).
	TypeInt
	// TypeFloat represents floating-point data (any precision).
	TypeFloat
	// TypeBool represents boolean data.
	TypeBool
	// TypeDate represents date data (without time).
	TypeDate
	// TypeTimestamp represents timestamp data (date + time).
	TypeTimestamp
	// TypeBinary represents binary/blob data.
	TypeBinary
	// TypeDecimal represents decimal/numeric data (fixed precision).
	TypeDecimal
	// TypeStruct represents structured data (nested fields).
	TypeStruct
	// TypeList represents list/array data.
	TypeList
)

// String returns the string representation of a DataType.
func (dt DataType) String() string {
	switch dt {
	case TypeString:
		return "String"
	case TypeInt:
		return "Int"
	case TypeFloat:
		return "Float"
	case TypeBool:
		return "Bool"
	case TypeDate:
		return "Date"
	case TypeTimestamp:
		return "Timestamp"
	case TypeBinary:
		return "Binary"
	case TypeDecimal:
		return "Decimal"
	case TypeStruct:
		return "Struct"
	case TypeList:
		return "List"
	default:
		return fmt.Sprintf("Unknown(%d)", dt)
	}
}

// Value is a typed container for cell values.
// It holds the raw value, type information, and a pre-formatted string for display.
type Value struct {
	// Raw holds the underlying value.
	// The type depends on the DataType field.
	Raw interface{}

	// Type indicates the data type of this value.
	Type DataType

	// IsNull indicates whether this value is null/nil.
	IsNull bool

	// Formatted is a pre-formatted string representation for display.
	// This improves UI performance by avoiding repeated formatting.
	Formatted string
}

// NewValue creates a new Value from a raw value and type.
func NewValue(raw interface{}, dataType DataType) Value {
	if raw == nil {
		return Value{
			Raw:       nil,
			Type:      dataType,
			IsNull:    true,
			Formatted: "",
		}
	}

	return Value{
		Raw:       raw,
		Type:      dataType,
		IsNull:    false,
		Formatted: formatValue(raw, dataType),
	}
}

// NewNullValue creates a null value of the specified type.
func NewNullValue(dataType DataType) Value {
	return Value{
		Raw:       nil,
		Type:      dataType,
		IsNull:    true,
		Formatted: "",
	}
}

// formatValue converts a raw value to a formatted string.
func formatValue(raw interface{}, dataType DataType) string {
	if raw == nil {
		return ""
	}

	// Use default string formatting for now
	// This can be enhanced with type-specific formatting
	return fmt.Sprintf("%v", raw)
}

// Metadata holds optional metadata about a data source.
type Metadata map[string]interface{}

// SortDirection specifies the direction of sorting.
type SortDirection int

const (
	// SortNone indicates no sorting.
	SortNone SortDirection = iota
	// SortAscending indicates ascending sort order.
	SortAscending
	// SortDescending indicates descending sort order.
	SortDescending
)

// String returns the string representation of a SortDirection.
func (sd SortDirection) String() string {
	switch sd {
	case SortNone:
		return "None"
	case SortAscending:
		return "Ascending"
	case SortDescending:
		return "Descending"
	default:
		return fmt.Sprintf("Unknown(%d)", sd)
	}
}

// SortState represents the current sorting configuration.
type SortState struct {
	// Column is the index of the sorted column (-1 if unsorted).
	Column int
	// Direction is the sort direction.
	Direction SortDirection
}

// IsSorted returns true if this state represents an active sort.
func (s SortState) IsSorted() bool {
	return s.Column >= 0 && s.Direction != SortNone
}
