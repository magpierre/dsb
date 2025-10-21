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

// Package arrow provides a DataSource adapter for Apache Arrow tables.
package arrow

import (
	"fmt"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/magpierre/fyne-datatable/datatable"
)

// ArrowDataSource implements datatable.DataSource for Apache Arrow tables.
type ArrowDataSource struct {
	table  arrow.Table
	schema *arrow.Schema
	reader *array.TableReader
	record arrow.Record
}

// NewFromArrowTable creates a DataSource from an Apache Arrow table.
// The Arrow table must remain valid for the lifetime of the DataSource.
// The caller is responsible for releasing the Arrow table when done.
func NewFromArrowTable(table arrow.Table) (*ArrowDataSource, error) {
	if table == nil {
		return nil, fmt.Errorf("arrow table cannot be nil")
	}

	if table.NumRows() == 0 {
		return nil, fmt.Errorf("arrow table must have at least one row")
	}

	if table.NumCols() == 0 {
		return nil, fmt.Errorf("arrow table must have at least one column")
	}

	// Create reader to access records
	reader := array.NewTableReader(table, table.NumRows())
	reader.Retain()

	// Read the first (and only) record containing all rows
	if !reader.Next() {
		reader.Release()
		return nil, fmt.Errorf("failed to read arrow table record")
	}

	record := reader.Record()
	record.Retain()

	return &ArrowDataSource{
		table:  table,
		schema: table.Schema(),
		reader: reader,
		record: record,
	}, nil
}

// Release releases the Arrow resources held by this DataSource.
// This should be called when the DataSource is no longer needed.
func (a *ArrowDataSource) Release() {
	if a.record != nil {
		a.record.Release()
	}
	if a.reader != nil {
		a.reader.Release()
	}
}

// ColumnCount returns the number of columns in the table.
func (a *ArrowDataSource) ColumnCount() int {
	return int(a.table.NumCols())
}

// RowCount returns the number of rows in the table.
func (a *ArrowDataSource) RowCount() int {
	return int(a.table.NumRows())
}

// ColumnName returns the name of the column at the given index.
func (a *ArrowDataSource) ColumnName(col int) (string, error) {
	if col < 0 || col >= int(a.table.NumCols()) {
		return "", fmt.Errorf("column index %d out of range [0, %d)", col, a.table.NumCols())
	}

	return a.schema.Field(col).Name, nil
}

// ColumnType returns the datatable type of the column at the given index.
func (a *ArrowDataSource) ColumnType(col int) (datatable.DataType, error) {
	if col < 0 || col >= int(a.table.NumCols()) {
		return datatable.TypeString, fmt.Errorf("column index %d out of range [0, %d)", col, a.table.NumCols())
	}

	arrowType := a.schema.Field(col).Type
	return mapArrowTypeToDataType(arrowType), nil
}

// Cell returns the value of the cell at the given row and column.
func (a *ArrowDataSource) Cell(row, col int) (datatable.Value, error) {
	if row < 0 || row >= int(a.table.NumRows()) {
		return datatable.Value{}, fmt.Errorf("row index %d out of range [0, %d)", row, a.table.NumRows())
	}

	if col < 0 || col >= int(a.table.NumCols()) {
		return datatable.Value{}, fmt.Errorf("column index %d out of range [0, %d)", col, a.table.NumCols())
	}

	column := a.record.Column(col)
	return extractArrowValue(column, row)
}

// Row returns all values in the given row.
func (a *ArrowDataSource) Row(row int) ([]datatable.Value, error) {
	if row < 0 || row >= int(a.table.NumRows()) {
		return nil, fmt.Errorf("row index %d out of range [0, %d)", row, a.table.NumRows())
	}

	values := make([]datatable.Value, a.table.NumCols())
	for col := 0; col < int(a.table.NumCols()); col++ {
		column := a.record.Column(col)
		value, err := extractArrowValue(column, row)
		if err != nil {
			return nil, fmt.Errorf("failed to extract value at row %d, col %d: %w", row, col, err)
		}
		values[col] = value
	}

	return values, nil
}

// Metadata returns metadata for the data source (optional for Arrow).
// Arrow tables don't have application-specific metadata, so this returns an empty Metadata.
func (a *ArrowDataSource) Metadata() datatable.Metadata {
	return datatable.Metadata{}
}

// mapArrowTypeToDataType maps Arrow types to datatable types.
func mapArrowTypeToDataType(arrowType arrow.DataType) datatable.DataType {
	switch arrowType.ID() {
	case arrow.STRING, arrow.LARGE_STRING:
		return datatable.TypeString

	case arrow.INT8, arrow.INT16, arrow.INT32, arrow.INT64,
		arrow.UINT8, arrow.UINT16, arrow.UINT32, arrow.UINT64:
		return datatable.TypeInt

	case arrow.FLOAT16, arrow.FLOAT32, arrow.FLOAT64:
		return datatable.TypeFloat

	case arrow.BOOL:
		return datatable.TypeBool

	case arrow.DATE32, arrow.DATE64:
		return datatable.TypeDate

	case arrow.TIMESTAMP, arrow.TIME32, arrow.TIME64:
		return datatable.TypeTimestamp

	case arrow.DECIMAL128, arrow.DECIMAL256:
		return datatable.TypeDecimal

	case arrow.BINARY, arrow.LARGE_BINARY, arrow.FIXED_SIZE_BINARY:
		return datatable.TypeBinary

	case arrow.STRUCT:
		return datatable.TypeStruct

	case arrow.LIST, arrow.LARGE_LIST, arrow.FIXED_SIZE_LIST:
		return datatable.TypeList

	default:
		// For unsupported types, default to string
		return datatable.TypeString
	}
}

// extractArrowValue extracts a value from an Arrow column at the given index.
func extractArrowValue(col arrow.Array, index int) (datatable.Value, error) {
	// Check for null
	if col.IsNull(index) {
		return datatable.Value{
			IsNull: true,
			Raw:    nil,
		}, nil
	}

	switch col.DataType().ID() {
	case arrow.STRING:
		s := col.(*array.String)
		val := s.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: val,
		}, nil

	case arrow.LARGE_STRING:
		s := col.(*array.LargeString)
		val := s.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: val,
		}, nil

	case arrow.INT8:
		i := col.(*array.Int8)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.INT16:
		i := col.(*array.Int16)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.INT32:
		i := col.(*array.Int32)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.INT64:
		i := col.(*array.Int64)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.UINT8:
		i := col.(*array.Uint8)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.UINT16:
		i := col.(*array.Uint16)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.UINT32:
		i := col.(*array.Uint32)
		val := i.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.UINT64:
		i := col.(*array.Uint64)
		val := i.Value(index)
		// Note: uint64 might overflow int64, but we'll cast for simplicity
		return datatable.Value{
			IsNull:    false,
			Raw:       int64(val),
			Formatted: fmt.Sprintf("%d", val),
		}, nil

	case arrow.FLOAT16:
		f := col.(*array.Float16)
		val := f.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       float64(val.Float32()),
			Formatted: fmt.Sprintf("%.2f", val.Float32()),
		}, nil

	case arrow.FLOAT32:
		f := col.(*array.Float32)
		val := f.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       float64(val),
			Formatted: fmt.Sprintf("%.2f", val),
		}, nil

	case arrow.FLOAT64:
		f := col.(*array.Float64)
		val := f.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: fmt.Sprintf("%.2f", val),
		}, nil

	case arrow.BOOL:
		b := col.(*array.Boolean)
		val := b.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: fmt.Sprintf("%v", val),
		}, nil

	case arrow.DATE32:
		d := col.(*array.Date32)
		val := d.Value(index)
		t := val.ToTime()
		return datatable.Value{
			IsNull:    false,
			Raw:       t,
			Formatted: t.Format("2006-01-02"),
		}, nil

	case arrow.DATE64:
		d := col.(*array.Date64)
		val := d.Value(index)
		t := val.ToTime()
		return datatable.Value{
			IsNull:    false,
			Raw:       t,
			Formatted: t.Format("2006-01-02"),
		}, nil

	case arrow.TIMESTAMP:
		ts := col.(*array.Timestamp)
		val := ts.Value(index)
		t := val.ToTime(arrow.Nanosecond)
		return datatable.Value{
			IsNull:    false,
			Raw:       t,
			Formatted: t.Format("2006-01-02 15:04:05"),
		}, nil

	case arrow.DECIMAL128:
		d := col.(*array.Decimal128)
		val := d.Value(index)
		str := val.BigInt().String()
		return datatable.Value{
			IsNull:    false,
			Raw:       str,
			Formatted: str,
		}, nil

	case arrow.BINARY:
		b := col.(*array.Binary)
		val := b.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: string(val),
		}, nil

	case arrow.LARGE_BINARY:
		b := col.(*array.LargeBinary)
		val := b.Value(index)
		return datatable.Value{
			IsNull:    false,
			Raw:       val,
			Formatted: string(val),
		}, nil

	case arrow.STRUCT:
		s := col.(*array.Struct)
		b, err := s.MarshalJSON()
		if err != nil {
			return datatable.Value{}, fmt.Errorf("failed to marshal struct: %w", err)
		}
		return datatable.Value{
			IsNull:    false,
			Raw:       string(b),
			Formatted: string(b),
		}, nil

	case arrow.LIST:
		l := col.(*array.List)
		// Create a slice of the list for this row
		offsets := l.Offsets()
		start := int(offsets[index])
		end := int(offsets[index+1])
		length := end - start

		str := fmt.Sprintf("[%d items]", length)
		if length > 0 && length <= 10 {
			// Show first few items for small lists
			values := l.ListValues()
			items := make([]string, 0, length)
			for i := start; i < end && i < start+10; i++ {
				val, err := extractArrowValue(values, i)
				if err == nil {
					items = append(items, val.Formatted)
				}
			}
			if len(items) > 0 {
				str = fmt.Sprintf("[%s]", fmt.Sprint(items))
			}
		}

		return datatable.Value{
			IsNull:    false,
			Raw:       str,
			Formatted: str,
		}, nil

	default:
		// For unsupported types, return string representation
		str := fmt.Sprintf("%v", col)
		return datatable.Value{
			IsNull:    false,
			Raw:       str,
			Formatted: str,
		}, nil
	}
}
