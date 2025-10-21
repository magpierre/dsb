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

package csv

import (
	"strings"
	"testing"

	"github.com/magpierre/fyne-datatable/datatable"
)

func TestNewFromReader_WithHeaders(t *testing.T) {
	csvData := `Name,Age,Role
Alice,30,Engineer
Bob,25,Designer
Charlie,35,Manager`

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, DefaultConfig())
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
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
	expectedNames := []string{"Name", "Age", "Role"}
	for i, expected := range expectedNames {
		name, err := source.ColumnName(i)
		if err != nil {
			t.Errorf("ColumnName(%d) error: %v", i, err)
		}
		if name != expected {
			t.Errorf("ColumnName(%d) = %s, want %s", i, name, expected)
		}
	}

	// Check data
	cell, err := source.Cell(0, 0)
	if err != nil {
		t.Errorf("Cell(0,0) error: %v", err)
	}
	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0,0) = %s, want Alice", cell.Formatted)
	}
}

func TestNewFromReader_WithoutHeaders(t *testing.T) {
	csvData := `Alice,30,Engineer
Bob,25,Designer`

	config := DefaultConfig()
	config.HasHeaders = false

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, config)
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
	}

	// Check row count
	if source.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", source.RowCount())
	}

	// Check generated column names
	expectedNames := []string{"Col1", "Col2", "Col3"}
	for i, expected := range expectedNames {
		name, err := source.ColumnName(i)
		if err != nil {
			t.Errorf("ColumnName(%d) error: %v", i, err)
		}
		if name != expected {
			t.Errorf("ColumnName(%d) = %s, want %s", i, name, expected)
		}
	}
}

func TestNewFromReader_TypeInference(t *testing.T) {
	csvData := `Name,Age,Salary,Active
Alice,30,50000.50,true
Bob,25,45000.00,false
Charlie,35,60000.75,true`

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, DefaultConfig())
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
	}

	// Check inferred types
	tests := []struct {
		col          int
		expectedType datatable.DataType
	}{
		{0, datatable.TypeString}, // Name
		{1, datatable.TypeInt},    // Age
		{2, datatable.TypeFloat},  // Salary
		{3, datatable.TypeBool},   // Active
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

func TestNewFromReader_CustomDelimiter(t *testing.T) {
	tsvData := "Name\tAge\tRole\nAlice\t30\tEngineer\nBob\t25\tDesigner"

	config := DefaultConfig()
	config.Delimiter = '\t'

	reader := strings.NewReader(tsvData)
	source, err := NewFromReader(reader, config)
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
	}

	if source.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", source.RowCount())
	}

	cell, _ := source.Cell(0, 0)
	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0,0) = %s, want Alice", cell.Formatted)
	}
}

func TestNewFromReader_EmptyFile(t *testing.T) {
	reader := strings.NewReader("")
	_, err := NewFromReader(reader, DefaultConfig())
	if err == nil {
		t.Error("Expected error for empty CSV")
	}
}

func TestNewFromReader_InconsistentColumns(t *testing.T) {
	csvData := `Name,Age,Role
Alice,30,Engineer
Bob,25
Charlie,35,Manager,Extra`

	reader := strings.NewReader(csvData)
	_, err := NewFromReader(reader, DefaultConfig())
	if err == nil {
		t.Error("Expected error for inconsistent column count")
	}
}

func TestNewFromReader_TrimSpace(t *testing.T) {
	csvData := `Name,Age,Role
 Alice ,  30  , Engineer
Bob,25,Designer`

	config := DefaultConfig()
	config.TrimSpace = true

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, config)
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
	}

	cell, _ := source.Cell(0, 0)
	if cell.Formatted != "Alice" {
		t.Errorf("Cell(0,0) = '%s', want 'Alice' (spaces should be trimmed)", cell.Formatted)
	}
}

func TestCSVDataSource_Row(t *testing.T) {
	csvData := `Name,Age,Role
Alice,30,Engineer`

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, DefaultConfig())
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
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

func TestCSVDataSource_Metadata(t *testing.T) {
	csvData := `Name,Age
Alice,30`

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, DefaultConfig())
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
	}

	metadata := source.Metadata()
	if metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestCSVDataSource_ErrorCases(t *testing.T) {
	csvData := `Name,Age
Alice,30`

	reader := strings.NewReader(csvData)
	source, err := NewFromReader(reader, DefaultConfig())
	if err != nil {
		t.Fatalf("NewFromReader failed: %v", err)
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
