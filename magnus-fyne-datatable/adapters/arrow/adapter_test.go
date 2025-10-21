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

package arrow

import (
	"testing"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/magpierre/fyne-datatable/datatable"
)

// Helper function to create a test Arrow table with various types
func createTestArrowTable() arrow.Table {
	pool := memory.NewGoAllocator()

	// Define schema
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "name", Type: arrow.BinaryTypes.String},
			{Name: "age", Type: arrow.PrimitiveTypes.Int32},
			{Name: "salary", Type: arrow.PrimitiveTypes.Float64},
			{Name: "active", Type: arrow.FixedWidthTypes.Boolean},
		},
		nil,
	)

	// Build arrays
	nameBuilder := array.NewStringBuilder(pool)
	nameBuilder.AppendValues([]string{"Alice", "Bob", "Charlie"}, nil)
	nameArray := nameBuilder.NewArray()
	defer nameArray.Release()

	ageBuilder := array.NewInt32Builder(pool)
	ageBuilder.AppendValues([]int32{30, 25, 35}, nil)
	ageArray := ageBuilder.NewArray()
	defer ageArray.Release()

	salaryBuilder := array.NewFloat64Builder(pool)
	salaryBuilder.AppendValues([]float64{75000.50, 65000.00, 85000.75}, nil)
	salaryArray := salaryBuilder.NewArray()
	defer salaryArray.Release()

	activeBuilder := array.NewBooleanBuilder(pool)
	activeBuilder.AppendValues([]bool{true, true, false}, nil)
	activeArray := activeBuilder.NewArray()
	defer activeArray.Release()

	// Create columns
	columns := []arrow.Column{
		*arrow.NewColumn(schema.Field(0), arrow.NewChunked(schema.Field(0).Type, []arrow.Array{nameArray})),
		*arrow.NewColumn(schema.Field(1), arrow.NewChunked(schema.Field(1).Type, []arrow.Array{ageArray})),
		*arrow.NewColumn(schema.Field(2), arrow.NewChunked(schema.Field(2).Type, []arrow.Array{salaryArray})),
		*arrow.NewColumn(schema.Field(3), arrow.NewChunked(schema.Field(3).Type, []arrow.Array{activeArray})),
	}

	// Create table
	return array.NewTable(schema, columns, 3)
}

// Helper function to create an Arrow table with null values
func createNullableArrowTable() arrow.Table {
	pool := memory.NewGoAllocator()

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
			{Name: "age", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
		},
		nil,
	)

	nameBuilder := array.NewStringBuilder(pool)
	nameBuilder.AppendValues([]string{"Alice", "Bob"}, []bool{true, false})
	nameArray := nameBuilder.NewArray()
	defer nameArray.Release()

	ageBuilder := array.NewInt32Builder(pool)
	ageBuilder.AppendValues([]int32{30, 0}, []bool{false, true})
	ageArray := ageBuilder.NewArray()
	defer ageArray.Release()

	columns := []arrow.Column{
		*arrow.NewColumn(schema.Field(0), arrow.NewChunked(schema.Field(0).Type, []arrow.Array{nameArray})),
		*arrow.NewColumn(schema.Field(1), arrow.NewChunked(schema.Field(1).Type, []arrow.Array{ageArray})),
	}

	return array.NewTable(schema, columns, 2)
}

// Helper function to create an Arrow table with date/timestamp types
func createDateTimeArrowTable() arrow.Table {
	pool := memory.NewGoAllocator()

	// Create timestamp type
	tsType := &arrow.TimestampType{Unit: arrow.Nanosecond}

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "date32", Type: arrow.FixedWidthTypes.Date32},
			{Name: "date64", Type: arrow.FixedWidthTypes.Date64},
			{Name: "timestamp", Type: tsType},
		},
		nil,
	)

	// Create date32 column
	date32Builder := array.NewDate32Builder(pool)
	date32Builder.AppendValues([]arrow.Date32{arrow.Date32FromTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}, nil)
	date32Array := date32Builder.NewArray()
	defer date32Array.Release()

	// Create date64 column
	date64Builder := array.NewDate64Builder(pool)
	date64Builder.AppendValues([]arrow.Date64{arrow.Date64FromTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))}, nil)
	date64Array := date64Builder.NewArray()
	defer date64Array.Release()

	// Create timestamp column (without timezone to match schema)
	tsBuilder := array.NewTimestampBuilder(pool, tsType)
	tsBuilder.AppendValues([]arrow.Timestamp{arrow.Timestamp(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC).UnixNano())}, nil)
	tsArray := tsBuilder.NewArray()
	defer tsArray.Release()

	columns := []arrow.Column{
		*arrow.NewColumn(schema.Field(0), arrow.NewChunked(schema.Field(0).Type, []arrow.Array{date32Array})),
		*arrow.NewColumn(schema.Field(1), arrow.NewChunked(schema.Field(1).Type, []arrow.Array{date64Array})),
		*arrow.NewColumn(schema.Field(2), arrow.NewChunked(schema.Field(2).Type, []arrow.Array{tsArray})),
	}

	return array.NewTable(schema, columns, 1)
}

func TestNewFromArrowTable(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, err := NewFromArrowTable(table)
	if err != nil {
		t.Fatalf("Failed to create ArrowDataSource: %v", err)
	}
	defer source.Release()

	if source == nil {
		t.Fatal("Expected non-nil ArrowDataSource")
	}

	if source.table == nil {
		t.Error("Expected table to be set")
	}

	if source.schema == nil {
		t.Error("Expected schema to be set")
	}
}

func TestNewFromArrowTable_NilTable(t *testing.T) {
	_, err := NewFromArrowTable(nil)
	if err == nil {
		t.Error("Expected error for nil table")
	}
}

func TestColumnCount(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	count := source.ColumnCount()
	expected := 4
	if count != expected {
		t.Errorf("Expected %d columns, got %d", expected, count)
	}
}

func TestRowCount(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	count := source.RowCount()
	expected := 3
	if count != expected {
		t.Errorf("Expected %d rows, got %d", expected, count)
	}
}

func TestColumnName(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	tests := []struct {
		col      int
		expected string
	}{
		{0, "name"},
		{1, "age"},
		{2, "salary"},
		{3, "active"},
	}

	for _, tt := range tests {
		name, err := source.ColumnName(tt.col)
		if err != nil {
			t.Errorf("ColumnName(%d) returned error: %v", tt.col, err)
		}
		if name != tt.expected {
			t.Errorf("ColumnName(%d) = %q, expected %q", tt.col, name, tt.expected)
		}
	}
}

func TestColumnName_OutOfRange(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	_, err := source.ColumnName(-1)
	if err == nil {
		t.Error("Expected error for negative column index")
	}

	_, err = source.ColumnName(10)
	if err == nil {
		t.Error("Expected error for out of range column index")
	}
}

func TestColumnType(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	tests := []struct {
		col      int
		expected datatable.DataType
	}{
		{0, datatable.TypeString},
		{1, datatable.TypeInt},
		{2, datatable.TypeFloat},
		{3, datatable.TypeBool},
	}

	for _, tt := range tests {
		colType, err := source.ColumnType(tt.col)
		if err != nil {
			t.Errorf("ColumnType(%d) returned error: %v", tt.col, err)
		}
		if colType != tt.expected {
			t.Errorf("ColumnType(%d) = %v, expected %v", tt.col, colType, tt.expected)
		}
	}
}

func TestCell(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	tests := []struct {
		row       int
		col       int
		formatted string
	}{
		{0, 0, "Alice"},
		{1, 0, "Bob"},
		{2, 0, "Charlie"},
		{0, 1, "30"},
		{1, 1, "25"},
		{2, 1, "35"},
		{0, 2, "75000.50"},
		{1, 2, "65000.00"},
		{2, 2, "85000.75"},
		{0, 3, "true"},
		{1, 3, "true"},
		{2, 3, "false"},
	}

	for _, tt := range tests {
		cell, err := source.Cell(tt.row, tt.col)
		if err != nil {
			t.Errorf("Cell(%d, %d) returned error: %v", tt.row, tt.col, err)
		}
		if cell.IsNull {
			t.Errorf("Cell(%d, %d) is null, expected non-null", tt.row, tt.col)
		}
		if cell.Formatted != tt.formatted {
			t.Errorf("Cell(%d, %d).Formatted = %q, expected %q", tt.row, tt.col, cell.Formatted, tt.formatted)
		}
	}
}

func TestCell_OutOfRange(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	_, err := source.Cell(-1, 0)
	if err == nil {
		t.Error("Expected error for negative row index")
	}

	_, err = source.Cell(10, 0)
	if err == nil {
		t.Error("Expected error for out of range row index")
	}

	_, err = source.Cell(0, -1)
	if err == nil {
		t.Error("Expected error for negative column index")
	}

	_, err = source.Cell(0, 10)
	if err == nil {
		t.Error("Expected error for out of range column index")
	}
}

func TestRow(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	row, err := source.Row(0)
	if err != nil {
		t.Fatalf("Row(0) returned error: %v", err)
	}

	if len(row) != 4 {
		t.Errorf("Expected 4 values in row, got %d", len(row))
	}

	expectedValues := []string{"Alice", "30", "75000.50", "true"}
	for i, expected := range expectedValues {
		if row[i].Formatted != expected {
			t.Errorf("Row(0)[%d].Formatted = %q, expected %q", i, row[i].Formatted, expected)
		}
	}
}

func TestRow_OutOfRange(t *testing.T) {
	table := createTestArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	_, err := source.Row(-1)
	if err == nil {
		t.Error("Expected error for negative row index")
	}

	_, err = source.Row(10)
	if err == nil {
		t.Error("Expected error for out of range row index")
	}
}

func TestNullValues(t *testing.T) {
	table := createNullableArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	// Row 0, col 0: "Alice" (not null)
	cell, err := source.Cell(0, 0)
	if err != nil {
		t.Errorf("Cell(0, 0) returned error: %v", err)
	}
	if cell.IsNull {
		t.Error("Cell(0, 0) should not be null")
	}
	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0, 0).Formatted = %q, expected \"Alice\"", cell.Formatted)
	}

	// Row 1, col 0: null
	cell, err = source.Cell(1, 0)
	if err != nil {
		t.Errorf("Cell(1, 0) returned error: %v", err)
	}
	if !cell.IsNull {
		t.Error("Cell(1, 0) should be null")
	}

	// Row 0, col 1: null
	cell, err = source.Cell(0, 1)
	if err != nil {
		t.Errorf("Cell(0, 1) returned error: %v", err)
	}
	if !cell.IsNull {
		t.Error("Cell(0, 1) should be null")
	}

	// Row 1, col 1: not null
	cell, err = source.Cell(1, 1)
	if err != nil {
		t.Errorf("Cell(1, 1) returned error: %v", err)
	}
	if cell.IsNull {
		t.Error("Cell(1, 1) should not be null")
	}
}

func TestDateTimeTypes(t *testing.T) {
	table := createDateTimeArrowTable()
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	// Test column types
	date32Type, _ := source.ColumnType(0)
	if date32Type != datatable.TypeDate {
		t.Errorf("Expected TypeDate for date32, got %v", date32Type)
	}

	date64Type, _ := source.ColumnType(1)
	if date64Type != datatable.TypeDate {
		t.Errorf("Expected TypeDate for date64, got %v", date64Type)
	}

	timestampType, _ := source.ColumnType(2)
	if timestampType != datatable.TypeTimestamp {
		t.Errorf("Expected TypeTimestamp for timestamp, got %v", timestampType)
	}

	// Test date32 value
	cell, err := source.Cell(0, 0)
	if err != nil {
		t.Errorf("Cell(0, 0) returned error: %v", err)
	}
	if cell.Formatted != "2023-01-01" {
		t.Errorf("Date32 formatted value = %q, expected \"2023-01-01\"", cell.Formatted)
	}

	// Test date64 value
	cell, err = source.Cell(0, 1)
	if err != nil {
		t.Errorf("Cell(0, 1) returned error: %v", err)
	}
	if cell.Formatted != "2023-01-01" {
		t.Errorf("Date64 formatted value = %q, expected \"2023-01-01\"", cell.Formatted)
	}

	// Test timestamp value
	cell, err = source.Cell(0, 2)
	if err != nil {
		t.Errorf("Cell(0, 2) returned error: %v", err)
	}
	// Timestamp should include time
	if cell.Formatted != "2023-01-01 12:00:00" {
		t.Errorf("Timestamp formatted value = %q, expected \"2023-01-01 12:00:00\"", cell.Formatted)
	}
}

func TestIntegerTypes(t *testing.T) {
	pool := memory.NewGoAllocator()

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "int8", Type: arrow.PrimitiveTypes.Int8},
			{Name: "int16", Type: arrow.PrimitiveTypes.Int16},
			{Name: "int32", Type: arrow.PrimitiveTypes.Int32},
			{Name: "int64", Type: arrow.PrimitiveTypes.Int64},
			{Name: "uint8", Type: arrow.PrimitiveTypes.Uint8},
			{Name: "uint16", Type: arrow.PrimitiveTypes.Uint16},
			{Name: "uint32", Type: arrow.PrimitiveTypes.Uint32},
			{Name: "uint64", Type: arrow.PrimitiveTypes.Uint64},
		},
		nil,
	)

	// Build arrays
	int8Builder := array.NewInt8Builder(pool)
	int8Builder.AppendValues([]int8{1}, nil)
	int8Array := int8Builder.NewArray()
	defer int8Array.Release()

	int16Builder := array.NewInt16Builder(pool)
	int16Builder.AppendValues([]int16{2}, nil)
	int16Array := int16Builder.NewArray()
	defer int16Array.Release()

	int32Builder := array.NewInt32Builder(pool)
	int32Builder.AppendValues([]int32{3}, nil)
	int32Array := int32Builder.NewArray()
	defer int32Array.Release()

	int64Builder := array.NewInt64Builder(pool)
	int64Builder.AppendValues([]int64{4}, nil)
	int64Array := int64Builder.NewArray()
	defer int64Array.Release()

	uint8Builder := array.NewUint8Builder(pool)
	uint8Builder.AppendValues([]uint8{5}, nil)
	uint8Array := uint8Builder.NewArray()
	defer uint8Array.Release()

	uint16Builder := array.NewUint16Builder(pool)
	uint16Builder.AppendValues([]uint16{6}, nil)
	uint16Array := uint16Builder.NewArray()
	defer uint16Array.Release()

	uint32Builder := array.NewUint32Builder(pool)
	uint32Builder.AppendValues([]uint32{7}, nil)
	uint32Array := uint32Builder.NewArray()
	defer uint32Array.Release()

	uint64Builder := array.NewUint64Builder(pool)
	uint64Builder.AppendValues([]uint64{8}, nil)
	uint64Array := uint64Builder.NewArray()
	defer uint64Array.Release()

	columns := []arrow.Column{
		*arrow.NewColumn(schema.Field(0), arrow.NewChunked(schema.Field(0).Type, []arrow.Array{int8Array})),
		*arrow.NewColumn(schema.Field(1), arrow.NewChunked(schema.Field(1).Type, []arrow.Array{int16Array})),
		*arrow.NewColumn(schema.Field(2), arrow.NewChunked(schema.Field(2).Type, []arrow.Array{int32Array})),
		*arrow.NewColumn(schema.Field(3), arrow.NewChunked(schema.Field(3).Type, []arrow.Array{int64Array})),
		*arrow.NewColumn(schema.Field(4), arrow.NewChunked(schema.Field(4).Type, []arrow.Array{uint8Array})),
		*arrow.NewColumn(schema.Field(5), arrow.NewChunked(schema.Field(5).Type, []arrow.Array{uint16Array})),
		*arrow.NewColumn(schema.Field(6), arrow.NewChunked(schema.Field(6).Type, []arrow.Array{uint32Array})),
		*arrow.NewColumn(schema.Field(7), arrow.NewChunked(schema.Field(7).Type, []arrow.Array{uint64Array})),
	}

	table := array.NewTable(schema, columns, 1)
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	// Test all integer types map to TypeInt
	for i := 0; i < 8; i++ {
		colType, _ := source.ColumnType(i)
		if colType != datatable.TypeInt {
			t.Errorf("Column %d: expected TypeInt, got %v", i, colType)
		}
	}

	// Test values
	expectedValues := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	for i, expected := range expectedValues {
		cell, err := source.Cell(0, i)
		if err != nil {
			t.Errorf("Cell(0, %d) returned error: %v", i, err)
		}
		if cell.Formatted != expected {
			t.Errorf("Cell(0, %d).Formatted = %q, expected %q", i, cell.Formatted, expected)
		}
	}
}

func TestFloatTypes(t *testing.T) {
	pool := memory.NewGoAllocator()

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "float32", Type: arrow.PrimitiveTypes.Float32},
			{Name: "float64", Type: arrow.PrimitiveTypes.Float64},
		},
		nil,
	)

	float32Builder := array.NewFloat32Builder(pool)
	float32Builder.AppendValues([]float32{1.5}, nil)
	float32Array := float32Builder.NewArray()
	defer float32Array.Release()

	float64Builder := array.NewFloat64Builder(pool)
	float64Builder.AppendValues([]float64{2.5}, nil)
	float64Array := float64Builder.NewArray()
	defer float64Array.Release()

	columns := []arrow.Column{
		*arrow.NewColumn(schema.Field(0), arrow.NewChunked(schema.Field(0).Type, []arrow.Array{float32Array})),
		*arrow.NewColumn(schema.Field(1), arrow.NewChunked(schema.Field(1).Type, []arrow.Array{float64Array})),
	}

	table := array.NewTable(schema, columns, 1)
	defer table.Release()

	source, _ := NewFromArrowTable(table)
	defer source.Release()

	// Test float types map to TypeFloat
	for i := 0; i < 2; i++ {
		colType, _ := source.ColumnType(i)
		if colType != datatable.TypeFloat {
			t.Errorf("Column %d: expected TypeFloat, got %v", i, colType)
		}
	}

	// Test float32 value
	cell, _ := source.Cell(0, 0)
	if cell.Formatted != "1.50" {
		t.Errorf("Float32 formatted value = %q, expected \"1.50\"", cell.Formatted)
	}

	// Test float64 value
	cell, _ = source.Cell(0, 1)
	if cell.Formatted != "2.50" {
		t.Errorf("Float64 formatted value = %q, expected \"2.50\"", cell.Formatted)
	}
}
