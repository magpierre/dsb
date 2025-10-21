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
	"errors"
	"sync"
	"testing"
)

// mockDataSource is a simple mock for testing
type mockDataSource struct {
	rows        int
	cols        int
	columnNames []string
	columnTypes []DataType
	data        [][]Value
}

func newMockDataSource(rows, cols int) *mockDataSource {
	columnNames := make([]string, cols)
	columnTypes := make([]DataType, cols)
	data := make([][]Value, rows)

	for i := 0; i < cols; i++ {
		columnNames[i] = string(rune('A' + i))
		columnTypes[i] = TypeString
	}

	for i := 0; i < rows; i++ {
		data[i] = make([]Value, cols)
		for j := 0; j < cols; j++ {
			data[i][j] = NewValue(string(rune('A'+j))+string(rune('0'+i)), TypeString)
		}
	}

	return &mockDataSource{
		rows:        rows,
		cols:        cols,
		columnNames: columnNames,
		columnTypes: columnTypes,
		data:        data,
	}
}

func (m *mockDataSource) RowCount() int {
	return m.rows
}

func (m *mockDataSource) ColumnCount() int {
	return m.cols
}

func (m *mockDataSource) ColumnName(col int) (string, error) {
	if col < 0 || col >= m.cols {
		return "", ErrInvalidColumn
	}
	return m.columnNames[col], nil
}

func (m *mockDataSource) ColumnType(col int) (DataType, error) {
	if col < 0 || col >= m.cols {
		return TypeString, ErrInvalidColumn
	}
	return m.columnTypes[col], nil
}

func (m *mockDataSource) Cell(row, col int) (Value, error) {
	if row < 0 || row >= m.rows {
		return Value{}, ErrInvalidRow
	}
	if col < 0 || col >= m.cols {
		return Value{}, ErrInvalidColumn
	}
	return m.data[row][col], nil
}

func (m *mockDataSource) Row(row int) ([]Value, error) {
	if row < 0 || row >= m.rows {
		return nil, ErrInvalidRow
	}
	result := make([]Value, len(m.data[row]))
	copy(result, m.data[row])
	return result, nil
}

func (m *mockDataSource) Metadata() Metadata {
	return make(Metadata)
}

func TestNewTableModel(t *testing.T) {
	tests := []struct {
		name    string
		source  DataSource
		wantErr bool
	}{
		{
			name:    "Valid source",
			source:  newMockDataSource(10, 3),
			wantErr: false,
		},
		{
			name:    "Nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name:    "Empty source",
			source:  newMockDataSource(0, 0),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := NewTableModel(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTableModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && model == nil {
				t.Error("NewTableModel() returned nil model")
			}
		})
	}
}

func TestTableModel_InitialState(t *testing.T) {
	source := newMockDataSource(10, 3)
	model, err := NewTableModel(source)
	if err != nil {
		t.Fatalf("NewTableModel() failed: %v", err)
	}

	// Check initial visible counts match original
	if model.VisibleRowCount() != source.RowCount() {
		t.Errorf("VisibleRowCount() = %d, want %d", model.VisibleRowCount(), source.RowCount())
	}

	if model.VisibleColumnCount() != source.ColumnCount() {
		t.Errorf("VisibleColumnCount() = %d, want %d", model.VisibleColumnCount(), source.ColumnCount())
	}

	// Check not sorted initially
	if model.IsSorted() {
		t.Error("Model should not be sorted initially")
	}

	// Check not filtered initially
	if model.IsFiltered() {
		t.Error("Model should not be filtered initially")
	}
}

func TestTableModel_VisibleCell(t *testing.T) {
	source := newMockDataSource(5, 3)
	model, _ := NewTableModel(source)

	tests := []struct {
		name    string
		row     int
		col     int
		want    string
		wantErr bool
	}{
		{"Valid cell (0,0)", 0, 0, "A0", false},
		{"Valid cell (2,1)", 2, 1, "B2", false},
		{"Invalid row -1", -1, 0, "", true},
		{"Invalid row out of range", 5, 0, "", true},
		{"Invalid col -1", 0, -1, "", true},
		{"Invalid col out of range", 0, 3, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.VisibleCell(tt.row, tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("VisibleCell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Formatted != tt.want {
				t.Errorf("VisibleCell() = %v, want %v", got.Formatted, tt.want)
			}
		})
	}
}

func TestTableModel_VisibleRow(t *testing.T) {
	source := newMockDataSource(3, 3)
	model, _ := NewTableModel(source)

	tests := []struct {
		name     string
		row      int
		wantLen  int
		wantErr  bool
		wantVals []string
	}{
		{"First row", 0, 3, false, []string{"A0", "B0", "C0"}},
		{"Second row", 1, 3, false, []string{"A1", "B1", "C1"}},
		{"Invalid row", -1, 0, true, nil},
		{"Out of range", 3, 0, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.VisibleRow(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("VisibleRow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.wantLen {
					t.Errorf("VisibleRow() length = %d, want %d", len(got), tt.wantLen)
				}
				for i, val := range got {
					if val.Formatted != tt.wantVals[i] {
						t.Errorf("VisibleRow()[%d] = %v, want %v", i, val.Formatted, tt.wantVals[i])
					}
				}
			}
		})
	}
}

func TestTableModel_SetVisibleColumns(t *testing.T) {
	source := newMockDataSource(5, 4)
	model, _ := NewTableModel(source)

	tests := []struct {
		name      string
		cols      []int
		wantErr   bool
		wantCount int
	}{
		{"Select subset", []int{0, 2}, false, 2},
		{"Select all", []int{0, 1, 2, 3}, false, 4},
		{"Select one", []int{1}, false, 1},
		{"Invalid column -1", []int{0, -1}, true, 4},
		{"Invalid column out of range", []int{0, 4}, true, 4},
		{"Duplicate columns", []int{0, 0}, true, 4},
		{"Empty selection", []int{}, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := model.SetVisibleColumns(tt.cols)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetVisibleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if model.VisibleColumnCount() != tt.wantCount {
					t.Errorf("VisibleColumnCount() = %d, want %d", model.VisibleColumnCount(), tt.wantCount)
				}
			}
		})
	}
}

func TestTableModel_ResetVisibleColumns(t *testing.T) {
	source := newMockDataSource(5, 4)
	model, _ := NewTableModel(source)

	// Hide some columns
	model.SetVisibleColumns([]int{0, 2})
	if model.VisibleColumnCount() != 2 {
		t.Errorf("After SetVisibleColumns, count = %d, want 2", model.VisibleColumnCount())
	}

	// Reset
	err := model.ResetVisibleColumns()
	if err != nil {
		t.Errorf("ResetVisibleColumns() error = %v", err)
	}

	// Should show all columns again
	if model.VisibleColumnCount() != source.ColumnCount() {
		t.Errorf("After reset, count = %d, want %d", model.VisibleColumnCount(), source.ColumnCount())
	}
}

func TestTableModel_VisibleColumnName(t *testing.T) {
	source := newMockDataSource(5, 4)
	model, _ := NewTableModel(source)

	// Hide column B (index 1)
	model.SetVisibleColumns([]int{0, 2, 3})

	tests := []struct {
		name    string
		col     int
		want    string
		wantErr bool
	}{
		{"First visible (A)", 0, "A", false},
		{"Second visible (C)", 1, "C", false},
		{"Third visible (D)", 2, "D", false},
		{"Invalid index", 3, "", true},
		{"Negative index", -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.VisibleColumnName(tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("VisibleColumnName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("VisibleColumnName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTableModel_ConcurrentReads(t *testing.T) {
	source := newMockDataSource(100, 10)
	model, _ := NewTableModel(source)

	// Spawn multiple goroutines doing concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = model.VisibleRowCount()
				_ = model.VisibleColumnCount()
				_, _ = model.VisibleCell(j%100, j%10)
				_ = model.IsSorted()
				_ = model.IsFiltered()
			}
		}()
	}

	wg.Wait()
}

func TestTableModel_ErrorTypes(t *testing.T) {
	source := newMockDataSource(5, 3)
	model, _ := NewTableModel(source)

	// Test that errors are of correct types
	_, err := model.VisibleCell(10, 0)
	if !errors.Is(err, ErrInvalidRow) {
		t.Errorf("Expected ErrInvalidRow, got %v", err)
	}

	_, err = model.VisibleCell(0, 10)
	if !errors.Is(err, ErrInvalidColumn) {
		t.Errorf("Expected ErrInvalidColumn, got %v", err)
	}

	err = model.SetVisibleColumns([]int{10})
	if !errors.Is(err, ErrInvalidColumn) {
		t.Errorf("Expected ErrInvalidColumn, got %v", err)
	}
}

func TestTableModel_GetDataSource(t *testing.T) {
	source := newMockDataSource(5, 3)
	model, _ := NewTableModel(source)

	got := model.GetDataSource()
	if got != source {
		t.Error("GetDataSource() did not return original source")
	}
}

func TestTableModel_GetVisibleColumnIndices(t *testing.T) {
	source := newMockDataSource(5, 4)
	model, _ := NewTableModel(source)

	// Set visible columns
	original := []int{0, 2, 3}
	model.SetVisibleColumns(original)

	// Get copy
	got := model.GetVisibleColumnIndices()

	// Verify values match
	if len(got) != len(original) {
		t.Fatalf("Length mismatch: got %d, want %d", len(got), len(original))
	}
	for i := range original {
		if got[i] != original[i] {
			t.Errorf("Index %d: got %d, want %d", i, got[i], original[i])
		}
	}

	// Verify it's a copy (modifying shouldn't affect model)
	got[0] = 999
	got2 := model.GetVisibleColumnIndices()
	if got2[0] == 999 {
		t.Error("GetVisibleColumnIndices() should return a copy, not original slice")
	}
}
