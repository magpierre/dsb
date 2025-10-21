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

package memory

import (
	"errors"
	"testing"

	"github.com/magpierre/fyne-datatable/datatable"
)

func TestNewDataSource(t *testing.T) {
	tests := []struct {
		name        string
		data        [][]string
		columnNames []string
		wantErr     bool
		wantRows    int
		wantCols    int
	}{
		{
			name: "Valid data",
			data: [][]string{
				{"Alice", "30", "Engineer"},
				{"Bob", "25", "Designer"},
			},
			columnNames: []string{"Name", "Age", "Role"},
			wantErr:     false,
			wantRows:    2,
			wantCols:    3,
		},
		{
			name:        "Empty data",
			data:        [][]string{},
			columnNames: []string{"Name", "Age"},
			wantErr:     false,
			wantRows:    0,
			wantCols:    2,
		},
		{
			name:        "No columns",
			data:        [][]string{{"test"}},
			columnNames: []string{},
			wantErr:     true,
		},
		{
			name: "Mismatched columns",
			data: [][]string{
				{"Alice", "30"},
				{"Bob", "25", "Designer"},
			},
			columnNames: []string{"Name", "Age", "Role"},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDataSource(tt.data, tt.columnNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDataSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.RowCount() != tt.wantRows {
					t.Errorf("RowCount() = %v, want %v", got.RowCount(), tt.wantRows)
				}
				if got.ColumnCount() != tt.wantCols {
					t.Errorf("ColumnCount() = %v, want %v", got.ColumnCount(), tt.wantCols)
				}
			}
		})
	}
}

func TestMemoryDataSource_ColumnName(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{{"Alice", "30"}},
		[]string{"Name", "Age"},
	)

	tests := []struct {
		name    string
		col     int
		want    string
		wantErr bool
	}{
		{"First column", 0, "Name", false},
		{"Second column", 1, "Age", false},
		{"Negative index", -1, "", true},
		{"Out of range", 2, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ds.ColumnName(tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("ColumnName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ColumnName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryDataSource_Cell(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{
			{"Alice", "30"},
			{"Bob", "25"},
		},
		[]string{"Name", "Age"},
	)

	tests := []struct {
		name    string
		row     int
		col     int
		want    string
		wantErr bool
	}{
		{"Valid cell (0,0)", 0, 0, "Alice", false},
		{"Valid cell (1,1)", 1, 1, "25", false},
		{"Invalid row", -1, 0, "", true},
		{"Invalid col", 0, -1, "", true},
		{"Row out of range", 2, 0, "", true},
		{"Col out of range", 0, 2, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ds.Cell(tt.row, tt.col)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Formatted != tt.want {
				t.Errorf("Cell() = %v, want %v", got.Formatted, tt.want)
			}
		})
	}
}

func TestMemoryDataSource_Row(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{
			{"Alice", "30", "Engineer"},
			{"Bob", "25", "Designer"},
		},
		[]string{"Name", "Age", "Role"},
	)

	tests := []struct {
		name     string
		row      int
		wantLen  int
		wantErr  bool
		wantVals []string
	}{
		{"First row", 0, 3, false, []string{"Alice", "30", "Engineer"}},
		{"Second row", 1, 3, false, []string{"Bob", "25", "Designer"}},
		{"Invalid row", -1, 0, true, nil},
		{"Out of range", 2, 0, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ds.Row(tt.row)
			if (err != nil) != tt.wantErr {
				t.Errorf("Row() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.wantLen {
					t.Errorf("Row() length = %v, want %v", len(got), tt.wantLen)
				}
				for i, val := range got {
					if val.Formatted != tt.wantVals[i] {
						t.Errorf("Row()[%d] = %v, want %v", i, val.Formatted, tt.wantVals[i])
					}
				}
			}
		})
	}
}

func TestMemoryDataSource_Metadata(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{{"Alice"}},
		[]string{"Name"},
	)

	// Initially empty
	meta := ds.Metadata()
	if len(meta) != 0 {
		t.Error("Initial metadata should be empty")
	}

	// Set metadata
	ds.SetMetadata("source", "test")
	ds.SetMetadata("version", 1)

	meta = ds.Metadata()
	if meta["source"] != "test" {
		t.Errorf("Metadata[source] = %v, want test", meta["source"])
	}
	if meta["version"] != 1 {
		t.Errorf("Metadata[version] = %v, want 1", meta["version"])
	}

	// Verify returned metadata is a copy (modifying it shouldn't affect source)
	meta["newkey"] = "value"
	meta2 := ds.Metadata()
	if _, exists := meta2["newkey"]; exists {
		t.Error("Modifying returned metadata should not affect source")
	}
}

func TestMemoryDataSource_ThreadSafety(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{
			{"Alice", "30"},
			{"Bob", "25"},
		},
		[]string{"Name", "Age"},
	)

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = ds.RowCount()
				_ = ds.ColumnCount()
				_, _ = ds.Cell(0, 0)
				_, _ = ds.Row(0)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryDataSource_ErrorTypes(t *testing.T) {
	ds, _ := NewDataSource(
		[][]string{{"Alice"}},
		[]string{"Name"},
	)

	// Test that errors are of the correct type
	_, err := ds.Cell(10, 0)
	if !errors.Is(err, datatable.ErrInvalidRow) {
		t.Errorf("Expected ErrInvalidRow, got %v", err)
	}

	_, err = ds.Cell(0, 10)
	if !errors.Is(err, datatable.ErrInvalidColumn) {
		t.Errorf("Expected ErrInvalidColumn, got %v", err)
	}

	_, err = ds.ColumnName(10)
	if !errors.Is(err, datatable.ErrInvalidColumn) {
		t.Errorf("Expected ErrInvalidColumn, got %v", err)
	}

	_, err = ds.Row(10)
	if !errors.Is(err, datatable.ErrInvalidRow) {
		t.Errorf("Expected ErrInvalidRow, got %v", err)
	}
}
