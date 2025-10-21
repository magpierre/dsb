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

package filter

import (
	"errors"
	"testing"

	"github.com/magpierre/fyne-datatable/datatable"
)

// mockDataSource for testing
type mockDataSource struct {
	rows        [][]datatable.Value
	columnNames []string
}

func newMockSource() *mockDataSource {
	return &mockDataSource{
		rows: [][]datatable.Value{
			{
				datatable.NewValue("Alice", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
				datatable.NewValue("Engineer", datatable.TypeString),
			},
			{
				datatable.NewValue("Bob", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
				datatable.NewValue("Designer", datatable.TypeString),
			},
			{
				datatable.NewValue("Charlie", datatable.TypeString),
				datatable.NewValue("35", datatable.TypeInt),
				datatable.NewValue("Manager", datatable.TypeString),
			},
			{
				datatable.NewValue("Diana", datatable.TypeString),
				datatable.NewValue("28", datatable.TypeInt),
				datatable.NewValue("Developer", datatable.TypeString),
			},
		},
		columnNames: []string{"Name", "Age", "Role"},
	}
}

func (m *mockDataSource) RowCount() int {
	return len(m.rows)
}

func (m *mockDataSource) ColumnCount() int {
	return len(m.columnNames)
}

func (m *mockDataSource) ColumnName(col int) (string, error) {
	if col < 0 || col >= len(m.columnNames) {
		return "", datatable.ErrInvalidColumn
	}
	return m.columnNames[col], nil
}

func (m *mockDataSource) ColumnType(col int) (datatable.DataType, error) {
	if col < 0 || col >= len(m.columnNames) {
		return datatable.TypeString, datatable.ErrInvalidColumn
	}
	return datatable.TypeString, nil
}

func (m *mockDataSource) Cell(row, col int) (datatable.Value, error) {
	if row < 0 || row >= len(m.rows) {
		return datatable.Value{}, datatable.ErrInvalidRow
	}
	if col < 0 || col >= len(m.columnNames) {
		return datatable.Value{}, datatable.ErrInvalidColumn
	}
	return m.rows[row][col], nil
}

func (m *mockDataSource) Row(row int) ([]datatable.Value, error) {
	if row < 0 || row >= len(m.rows) {
		return nil, datatable.ErrInvalidRow
	}
	result := make([]datatable.Value, len(m.rows[row]))
	copy(result, m.rows[row])
	return result, nil
}

func (m *mockDataSource) Metadata() datatable.Metadata {
	return make(datatable.Metadata)
}

func TestEngine_Apply_SimpleFilter(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	tests := []struct {
		name      string
		filter    *SimpleFilter
		wantCount int
		wantRows  []int
	}{
		{
			name:      "Age > 25",
			filter:    &SimpleFilter{Column: "Age", Operator: OpGreaterThan, Value: "25"},
			wantCount: 3,
			wantRows:  []int{0, 2, 3}, // Alice(30), Charlie(35), Diana(28)
		},
		{
			name:      "Age = 30",
			filter:    &SimpleFilter{Column: "Age", Operator: OpEqual, Value: "30"},
			wantCount: 1,
			wantRows:  []int{0}, // Alice
		},
		{
			name:      "Name contains 'i'",
			filter:    &SimpleFilter{Column: "Name", Operator: OpContains, Value: "i"},
			wantCount: 3,
			wantRows:  []int{0, 2, 3}, // Alice, Charlie, Diana
		},
		{
			name:      "Role = Engineer",
			filter:    &SimpleFilter{Column: "Role", Operator: OpEqual, Value: "Engineer"},
			wantCount: 1,
			wantRows:  []int{0}, // Alice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Apply(source, tt.filter)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("Apply() got %d rows, want %d", len(got), tt.wantCount)
			}
			for i, rowIdx := range tt.wantRows {
				if i >= len(got) || got[i] != rowIdx {
					t.Errorf("Apply() row indices = %v, want %v", got, tt.wantRows)
					break
				}
			}
		})
	}
}

func TestEngine_Apply_CompositeFilter(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	// Age > 25 AND Role contains 'er'
	filter := &CompositeFilter{
		Filters: []datatable.Filter{
			&SimpleFilter{Column: "Age", Operator: OpGreaterThan, Value: "25"},
			&SimpleFilter{Column: "Role", Operator: OpContains, Value: "er"},
		},
		Logic: LogicAND,
	}

	got, err := engine.Apply(source, filter)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	// Should match: Alice(30, Engineer), Charlie(35, Manager), Diana(28, Developer)
	want := 3
	if len(got) != want {
		t.Errorf("Apply() got %d rows, want %d", len(got), want)
	}
}

func TestEngine_ApplyMultiple(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	filters := []datatable.Filter{
		&SimpleFilter{Column: "Age", Operator: OpGreaterThan, Value: "25"},
		&SimpleFilter{Column: "Role", Operator: OpContains, Value: "er"},
	}

	got, err := engine.ApplyMultiple(source, filters)
	if err != nil {
		t.Fatalf("ApplyMultiple() error = %v", err)
	}

	// Should match all rows that pass both filters (AND logic)
	want := 3
	if len(got) != want {
		t.Errorf("ApplyMultiple() got %d rows, want %d", len(got), want)
	}
}

func TestEngine_ApplyToIndices(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	// Start with subset: rows 0, 2, 3 (Alice, Charlie, Diana)
	indices := []int{0, 2, 3}

	// Filter: Age < 32
	filter := &SimpleFilter{Column: "Age", Operator: OpLessThan, Value: "32"}

	got, err := engine.ApplyToIndices(source, filter, indices)
	if err != nil {
		t.Fatalf("ApplyToIndices() error = %v", err)
	}

	// Should match: Alice(30), Diana(28) from the subset
	want := []int{0, 3}
	if len(got) != len(want) {
		t.Errorf("ApplyToIndices() got %d rows, want %d", len(got), len(want))
	}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Errorf("ApplyToIndices() = %v, want %v", got, want)
			break
		}
	}
}

func TestEngine_Validate(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	tests := []struct {
		name    string
		filter  datatable.Filter
		wantErr bool
	}{
		{
			name:    "Valid filter",
			filter:  &SimpleFilter{Column: "Age", Operator: OpGreaterThan, Value: "25"},
			wantErr: false,
		},
		{
			name:    "Invalid column",
			filter:  &SimpleFilter{Column: "InvalidColumn", Operator: OpEqual, Value: "test"},
			wantErr: true,
		},
		{
			name:    "Nil filter",
			filter:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.Validate(tt.filter, source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_ErrorHandling(t *testing.T) {
	engine := NewEngine()

	// Test nil source
	_, err := engine.Apply(nil, &SimpleFilter{Column: "Age", Operator: OpEqual, Value: "25"})
	if !errors.Is(err, datatable.ErrNoDataSource) {
		t.Errorf("Expected ErrNoDataSource, got %v", err)
	}

	// Test nil filter
	source := newMockSource()
	_, err = engine.Apply(source, nil)
	if !errors.Is(err, datatable.ErrInvalidFilter) {
		t.Errorf("Expected ErrInvalidFilter, got %v", err)
	}
}

func TestEngine_EmptyFilters(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	// ApplyMultiple with no filters should return all rows
	got, err := engine.ApplyMultiple(source, []datatable.Filter{})
	if err != nil {
		t.Fatalf("ApplyMultiple() error = %v", err)
	}

	if len(got) != source.RowCount() {
		t.Errorf("ApplyMultiple() with no filters got %d rows, want %d", len(got), source.RowCount())
	}
}
