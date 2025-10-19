package slice

import (
	"testing"

	"github.com/magpierre/fyne-datatable/datatable"
)

func TestNewFromInterfaces_Basic(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30, true},
		{"Bob", 25, false},
		{"Charlie", 35, true},
	}
	headers := []string{"Name", "Age", "Active"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	// Check row count
	if source.RowCount() != 3 {
		t.Errorf("Expected 3 rows, got %d", source.RowCount())
	}

	// Check column count
	if source.ColumnCount() != 3 {
		t.Errorf("Expected 3 columns, got %d", source.ColumnCount())
	}

	// Check column names
	for i, expected := range headers {
		name, err := source.ColumnName(i)
		if err != nil {
			t.Errorf("ColumnName(%d) error: %v", i, err)
		}
		if name != expected {
			t.Errorf("ColumnName(%d) = %s, want %s", i, name, expected)
		}
	}
}

func TestNewFromInterfaces_TypeInference(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30, 50000.50, true},
		{"Bob", 25, 45000.00, false},
	}
	headers := []string{"Name", "Age", "Salary", "Active"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	// Check inferred types
	tests := []struct {
		col          int
		expectedType datatable.DataType
	}{
		{0, datatable.TypeString},
		{1, datatable.TypeInt},
		{2, datatable.TypeFloat},
		{3, datatable.TypeBool},
	}

	for _, tt := range tests {
		colType, err := source.ColumnType(tt.col)
		if err != nil {
			t.Errorf("ColumnType(%d) error: %v", tt.col, err)
		}
		if colType != tt.expectedType {
			t.Errorf("ColumnType(%d) = %v, want %v", tt.col, colType, tt.expectedType)
		}
	}
}

func TestNewFromInterfaces_NilValues(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30, nil},
		{nil, 25, "Engineer"},
	}
	headers := []string{"Name", "Age", "Role"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	// Check nil value
	cell, err := source.Cell(0, 2)
	if err != nil {
		t.Errorf("Cell(0,2) error: %v", err)
	}
	if !cell.IsNull {
		t.Errorf("Cell(0,2) should be null")
	}
}

func TestNewFromInterfaces_InconsistentColumns(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30, "Engineer"},
		{"Bob", 25}, // Missing column
	}
	headers := []string{"Name", "Age", "Role"}

	_, err := NewFromInterfaces(data, headers)
	if err == nil {
		t.Error("Expected error for inconsistent column count")
	}
}

func TestNewFromInterfaces_EmptyHeaders(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30},
	}

	_, err := NewFromInterfaces(data, []string{})
	if err == nil {
		t.Error("Expected error for empty headers")
	}
}

func TestNewFromInterfaces_NilData(t *testing.T) {
	_, err := NewFromInterfaces(nil, []string{"Name"})
	if err == nil {
		t.Error("Expected error for nil data")
	}
}

func TestNewFromStrings(t *testing.T) {
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
	}
	headers := []string{"Name", "Age", "Role"}

	source, err := NewFromStrings(data, headers)
	if err != nil {
		t.Fatalf("NewFromStrings failed: %v", err)
	}

	if source.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", source.RowCount())
	}

	cell, _ := source.Cell(0, 0)
	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0,0) = %s, want Alice", cell.Formatted)
	}
}

func TestNewFromMaps(t *testing.T) {
	data := []map[string]interface{}{
		{"Name": "Alice", "Age": 30, "Role": "Engineer"},
		{"Name": "Bob", "Age": 25, "Role": "Designer"},
	}

	source, err := NewFromMaps(data)
	if err != nil {
		t.Fatalf("NewFromMaps failed: %v", err)
	}

	if source.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", source.RowCount())
	}

	if source.ColumnCount() != 3 {
		t.Errorf("Expected 3 columns, got %d", source.ColumnCount())
	}

	// Note: column order from map is not guaranteed, so we just check count
}

func TestNewFromMaps_EmptyData(t *testing.T) {
	_, err := NewFromMaps([]map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestNewFromMaps_NilData(t *testing.T) {
	_, err := NewFromMaps(nil)
	if err == nil {
		t.Error("Expected error for nil data")
	}
}

func TestSliceDataSource_Cell(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30},
		{"Bob", 25},
	}
	headers := []string{"Name", "Age"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	cell, err := source.Cell(0, 0)
	if err != nil {
		t.Fatalf("Cell(0,0) error: %v", err)
	}

	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0,0) = %s, want Alice", cell.Formatted)
	}
}

func TestSliceDataSource_Row(t *testing.T) {
	data := [][]interface{}{
		{"Alice", 30, "Engineer"},
	}
	headers := []string{"Name", "Age", "Role"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	row, err := source.Row(0)
	if err != nil {
		t.Fatalf("Row(0) error: %v", err)
	}

	if len(row) != 3 {
		t.Errorf("Row length = %d, want 3", len(row))
	}

	if row[0].Formatted != "Alice" {
		t.Errorf("Row[0] = %s, want Alice", row[0].Formatted)
	}
}

func TestSliceDataSource_Metadata(t *testing.T) {
	data := [][]interface{}{{"Alice", 30}}
	headers := []string{"Name", "Age"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	metadata := source.Metadata()
	if metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestSliceDataSource_ErrorCases(t *testing.T) {
	data := [][]interface{}{{"Alice", 30}}
	headers := []string{"Name", "Age"}

	source, err := NewFromInterfaces(data, headers)
	if err != nil {
		t.Fatalf("NewFromInterfaces failed: %v", err)
	}

	// Test invalid row
	_, err = source.Cell(10, 0)
	if err != datatable.ErrInvalidRow {
		t.Errorf("Expected ErrInvalidRow, got %v", err)
	}

	// Test invalid column
	_, err = source.Cell(0, 10)
	if err != datatable.ErrInvalidColumn {
		t.Errorf("Expected ErrInvalidColumn, got %v", err)
	}

	// Test invalid column name
	_, err = source.ColumnName(10)
	if err != datatable.ErrInvalidColumn {
		t.Errorf("Expected ErrInvalidColumn for ColumnName, got %v", err)
	}

	// Test invalid column type
	_, err = source.ColumnType(10)
	if err != datatable.ErrInvalidColumn {
		t.Errorf("Expected ErrInvalidColumn for ColumnType, got %v", err)
	}

	// Test invalid row for Row()
	_, err = source.Row(10)
	if err != datatable.ErrInvalidRow {
		t.Errorf("Expected ErrInvalidRow for Row(), got %v", err)
	}
}

func TestConvertToValue_IntTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected datatable.DataType
	}{
		{"int", int(42), datatable.TypeInt},
		{"int32", int32(42), datatable.TypeInt},
		{"int64", int64(42), datatable.TypeInt},
		{"uint", uint(42), datatable.TypeInt},
		{"uint32", uint32(42), datatable.TypeInt},
		{"uint64", uint64(42), datatable.TypeInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := convertToValue(tt.value)
			if val.Type != tt.expected {
				t.Errorf("convertToValue(%v) type = %v, want %v", tt.value, val.Type, tt.expected)
			}
		})
	}
}

func TestConvertToValue_FloatTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected datatable.DataType
	}{
		{"float32", float32(3.14), datatable.TypeFloat},
		{"float64", float64(3.14), datatable.TypeFloat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := convertToValue(tt.value)
			if val.Type != tt.expected {
				t.Errorf("convertToValue(%v) type = %v, want %v", tt.value, val.Type, tt.expected)
			}
		})
	}
}

func TestConvertToValue_Bool(t *testing.T) {
	val := convertToValue(true)
	if val.Type != datatable.TypeBool {
		t.Errorf("convertToValue(true) type = %v, want TypeBool", val.Type)
	}
}

func TestConvertToValue_String(t *testing.T) {
	val := convertToValue("hello")
	if val.Type != datatable.TypeString {
		t.Errorf("convertToValue(\"hello\") type = %v, want TypeString", val.Type)
	}
}

func TestConvertToValue_Nil(t *testing.T) {
	val := convertToValue(nil)
	if !val.IsNull {
		t.Error("convertToValue(nil) should create null value")
	}
}
