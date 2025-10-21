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

package sort

import (
	"errors"
	"testing"

	"github.com/magpierre/fyne-datatable/datatable"
)

// mockDataSource for testing
type mockDataSource struct {
	rows        [][]datatable.Value
	columnNames []string
	columnTypes []datatable.DataType
}

func newMockSource() *mockDataSource {
	return &mockDataSource{
		rows: [][]datatable.Value{
			{
				datatable.NewValue("Alice", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
				datatable.NewValue("Engineer", datatable.TypeString),
				datatable.NewValue("2024-01-15", datatable.TypeDate),
			},
			{
				datatable.NewValue("Bob", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
				datatable.NewValue("Designer", datatable.TypeString),
				datatable.NewValue("2024-03-20", datatable.TypeDate),
			},
			{
				datatable.NewValue("Charlie", datatable.TypeString),
				datatable.NewValue("35", datatable.TypeInt),
				datatable.NewValue("Manager", datatable.TypeString),
				datatable.NewValue("2024-02-10", datatable.TypeDate),
			},
			{
				datatable.NewValue("Diana", datatable.TypeString),
				datatable.NewValue("28", datatable.TypeInt),
				datatable.NewValue("Developer", datatable.TypeString),
				datatable.NewValue("2024-01-05", datatable.TypeDate),
			},
		},
		columnNames: []string{"Name", "Age", "Role", "JoinDate"},
		columnTypes: []datatable.DataType{
			datatable.TypeString,
			datatable.TypeInt,
			datatable.TypeString,
			datatable.TypeDate,
		},
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
	if col < 0 || col >= len(m.columnTypes) {
		return datatable.TypeString, datatable.ErrInvalidColumn
	}
	return m.columnTypes[col], nil
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

// TestEngine_Sort_Ascending tests ascending sort
func TestEngine_Sort_Ascending(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()
	indices := []int{0, 1, 2, 3} // All rows

	tests := []struct {
		name     string
		column   int
		dataType datatable.DataType
		want     []int
	}{
		{
			name:     "Sort by Name (ascending)",
			column:   0,
			dataType: datatable.TypeString,
			want:     []int{0, 1, 2, 3}, // Alice, Bob, Charlie, Diana
		},
		{
			name:     "Sort by Age (ascending)",
			column:   1,
			dataType: datatable.TypeInt,
			want:     []int{1, 3, 0, 2}, // 25, 28, 30, 35
		},
		{
			name:     "Sort by Role (ascending)",
			column:   2,
			dataType: datatable.TypeString,
			want:     []int{1, 3, 0, 2}, // Designer, Developer, Engineer, Manager
		},
		{
			name:     "Sort by JoinDate (ascending)",
			column:   3,
			dataType: datatable.TypeDate,
			want:     []int{3, 0, 2, 1}, // 2024-01-05, 2024-01-15, 2024-02-10, 2024-03-20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := SortSpec{
				Column:    tt.column,
				Direction: datatable.SortAscending,
				DataType:  tt.dataType,
			}

			got, err := engine.Sort(source, indices, spec)
			if err != nil {
				t.Errorf("Sort() error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Sort() got %d indices, want %d", len(got), len(tt.want))
				return
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Sort() indices[%d] = %d, want %d (full result: %v, want %v)",
						i, got[i], tt.want[i], got, tt.want)
					break
				}
			}
		})
	}
}

// TestEngine_Sort_Descending tests descending sort
func TestEngine_Sort_Descending(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()
	indices := []int{0, 1, 2, 3}

	tests := []struct {
		name     string
		column   int
		dataType datatable.DataType
		want     []int
	}{
		{
			name:     "Sort by Name (descending)",
			column:   0,
			dataType: datatable.TypeString,
			want:     []int{3, 2, 1, 0}, // Diana, Charlie, Bob, Alice
		},
		{
			name:     "Sort by Age (descending)",
			column:   1,
			dataType: datatable.TypeInt,
			want:     []int{2, 0, 3, 1}, // 35, 30, 28, 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := SortSpec{
				Column:    tt.column,
				Direction: datatable.SortDescending,
				DataType:  tt.dataType,
			}

			got, err := engine.Sort(source, indices, spec)
			if err != nil {
				t.Errorf("Sort() error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Sort() got %d indices, want %d", len(got), len(tt.want))
				return
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Sort() indices = %v, want %v", got, tt.want)
					break
				}
			}
		})
	}
}

// TestEngine_Sort_SortNone tests that SortNone returns unchanged indices
func TestEngine_Sort_SortNone(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()
	indices := []int{3, 1, 2, 0} // Deliberately out of order

	spec := SortSpec{
		Column:    0,
		Direction: datatable.SortNone,
		DataType:  datatable.TypeString,
	}

	got, err := engine.Sort(source, indices, spec)
	if err != nil {
		t.Fatalf("Sort() error = %v", err)
	}

	// Should return copy of original indices unchanged
	if len(got) != len(indices) {
		t.Errorf("Sort() got %d indices, want %d", len(got), len(indices))
	}

	for i := range indices {
		if got[i] != indices[i] {
			t.Errorf("Sort() indices = %v, want %v (SortNone should preserve order)", got, indices)
			break
		}
	}
}

// TestEngine_Sort_SubsetOfRows tests sorting a subset of rows
func TestEngine_Sort_SubsetOfRows(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()

	// Only sort first 3 rows: Alice(30), Bob(25), Charlie(35)
	indices := []int{0, 1, 2}

	spec := SortSpec{
		Column:    1, // Age
		Direction: datatable.SortAscending,
		DataType:  datatable.TypeInt,
	}

	got, err := engine.Sort(source, indices, spec)
	if err != nil {
		t.Fatalf("Sort() error = %v", err)
	}

	// Should be: Bob(25), Alice(30), Charlie(35)
	want := []int{1, 0, 2}

	if len(got) != len(want) {
		t.Errorf("Sort() got %d indices, want %d", len(got), len(want))
		return
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Sort() indices = %v, want %v", got, want)
			break
		}
	}
}

// TestEngine_MultiSort tests multi-column sorting
func TestEngine_MultiSort(t *testing.T) {
	engine := NewEngine()

	// Create source with duplicate ages
	source := &mockDataSource{
		rows: [][]datatable.Value{
			{
				datatable.NewValue("Alice", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
			},
			{
				datatable.NewValue("Bob", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
			},
			{
				datatable.NewValue("Charlie", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
			},
			{
				datatable.NewValue("Diana", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
			},
		},
		columnNames: []string{"Name", "Age"},
		columnTypes: []datatable.DataType{
			datatable.TypeString,
			datatable.TypeInt,
		},
	}

	indices := []int{0, 1, 2, 3}

	// Sort by Age ascending, then Name ascending
	specs := []SortSpec{
		{Column: 1, Direction: datatable.SortAscending, DataType: datatable.TypeInt},
		{Column: 0, Direction: datatable.SortAscending, DataType: datatable.TypeString},
	}

	got, err := engine.MultiSort(source, indices, specs)
	if err != nil {
		t.Fatalf("MultiSort() error = %v", err)
	}

	// Expected order:
	// 1. Charlie(25)
	// 2. Diana(25)
	// 3. Alice(30)
	// 4. Bob(30)
	want := []int{2, 3, 0, 1}

	if len(got) != len(want) {
		t.Errorf("MultiSort() got %d indices, want %d", len(got), len(want))
		return
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("MultiSort() indices = %v, want %v", got, want)
			break
		}
	}
}

// TestEngine_MultiSort_MixedDirections tests multi-column sort with different directions
func TestEngine_MultiSort_MixedDirections(t *testing.T) {
	engine := NewEngine()

	source := &mockDataSource{
		rows: [][]datatable.Value{
			{
				datatable.NewValue("Alice", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
			},
			{
				datatable.NewValue("Bob", datatable.TypeString),
				datatable.NewValue("30", datatable.TypeInt),
			},
			{
				datatable.NewValue("Charlie", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
			},
			{
				datatable.NewValue("Diana", datatable.TypeString),
				datatable.NewValue("25", datatable.TypeInt),
			},
		},
		columnNames: []string{"Name", "Age"},
		columnTypes: []datatable.DataType{
			datatable.TypeString,
			datatable.TypeInt,
		},
	}

	indices := []int{0, 1, 2, 3}

	// Sort by Age ascending, then Name descending
	specs := []SortSpec{
		{Column: 1, Direction: datatable.SortAscending, DataType: datatable.TypeInt},
		{Column: 0, Direction: datatable.SortDescending, DataType: datatable.TypeString},
	}

	got, err := engine.MultiSort(source, indices, specs)
	if err != nil {
		t.Fatalf("MultiSort() error = %v", err)
	}

	// Expected order:
	// 1. Diana(25)  - age 25, name D
	// 2. Charlie(25) - age 25, name C
	// 3. Bob(30)     - age 30, name B
	// 4. Alice(30)   - age 30, name A
	want := []int{3, 2, 1, 0}

	if len(got) != len(want) {
		t.Errorf("MultiSort() got %d indices, want %d", len(got), len(want))
		return
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("MultiSort() indices = %v, want %v", got, want)
			break
		}
	}
}

// TestEngine_Sort_ErrorHandling tests error cases
func TestEngine_Sort_ErrorHandling(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		source    datatable.DataSource
		indices   []int
		spec      SortSpec
		wantError error
	}{
		{
			name:      "Nil source",
			source:    nil,
			indices:   []int{0, 1},
			spec:      SortSpec{Column: 0, Direction: datatable.SortAscending},
			wantError: datatable.ErrNoDataSource,
		},
		{
			name:      "Invalid column (negative)",
			source:    newMockSource(),
			indices:   []int{0, 1},
			spec:      SortSpec{Column: -1, Direction: datatable.SortAscending},
			wantError: datatable.ErrInvalidSortColumn,
		},
		{
			name:      "Invalid column (too large)",
			source:    newMockSource(),
			indices:   []int{0, 1},
			spec:      SortSpec{Column: 999, Direction: datatable.SortAscending},
			wantError: datatable.ErrInvalidSortColumn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Sort(tt.source, tt.indices, tt.spec)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("Sort() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

// TestEngine_MultiSort_ErrorHandling tests error cases for multi-sort
func TestEngine_MultiSort_ErrorHandling(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		source    datatable.DataSource
		indices   []int
		specs     []SortSpec
		wantError error
	}{
		{
			name:      "Nil source",
			source:    nil,
			indices:   []int{0, 1},
			specs:     []SortSpec{{Column: 0, Direction: datatable.SortAscending}},
			wantError: datatable.ErrNoDataSource,
		},
		{
			name:    "Invalid column in spec",
			source:  newMockSource(),
			indices: []int{0, 1},
			specs: []SortSpec{
				{Column: 0, Direction: datatable.SortAscending},
				{Column: 999, Direction: datatable.SortAscending},
			},
			wantError: datatable.ErrInvalidSortColumn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.MultiSort(tt.source, tt.indices, tt.specs)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("MultiSort() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

// TestEngine_MultiSort_EmptySpecs tests that empty specs returns unchanged indices
func TestEngine_MultiSort_EmptySpecs(t *testing.T) {
	engine := NewEngine()
	source := newMockSource()
	indices := []int{3, 1, 2, 0}

	got, err := engine.MultiSort(source, indices, []SortSpec{})
	if err != nil {
		t.Fatalf("MultiSort() error = %v", err)
	}

	// Should return copy of original indices
	if len(got) != len(indices) {
		t.Errorf("MultiSort() got %d indices, want %d", len(got), len(indices))
	}

	for i := range indices {
		if got[i] != indices[i] {
			t.Errorf("MultiSort() indices = %v, want %v", got, indices)
			break
		}
	}
}

// TestEngine_Sort_NullHandling tests that nulls sort to end
func TestEngine_Sort_NullHandling(t *testing.T) {
	engine := NewEngine()

	source := &mockDataSource{
		rows: [][]datatable.Value{
			{datatable.NewValue("Alice", datatable.TypeString), datatable.NewValue("30", datatable.TypeInt)},
			{datatable.NewValue("Bob", datatable.TypeString), datatable.Value{IsNull: true, Type: datatable.TypeInt}},
			{datatable.NewValue("Charlie", datatable.TypeString), datatable.NewValue("25", datatable.TypeInt)},
			{datatable.NewValue("Diana", datatable.TypeString), datatable.Value{IsNull: true, Type: datatable.TypeInt}},
		},
		columnNames: []string{"Name", "Age"},
		columnTypes: []datatable.DataType{datatable.TypeString, datatable.TypeInt},
	}

	indices := []int{0, 1, 2, 3}

	spec := SortSpec{
		Column:    1, // Age column with nulls
		Direction: datatable.SortAscending,
		DataType:  datatable.TypeInt,
	}

	got, err := engine.Sort(source, indices, spec)
	if err != nil {
		t.Fatalf("Sort() error = %v", err)
	}

	// Non-null values should come first (25, 30), then nulls
	// Expected: Charlie(25), Alice(30), Bob(null), Diana(null)
	want := []int{2, 0, 1, 3}

	if len(got) != len(want) {
		t.Errorf("Sort() got %d indices, want %d", len(got), len(want))
		return
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Sort() indices = %v, want %v (nulls should sort to end)", got, want)
			break
		}
	}
}

// TestCompareNumeric tests numeric comparison
func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"10 > 9", "10", "9", 1},
		{"9 < 10", "9", "10", -1},
		{"5 = 5", "5", "5", 0},
		{"negative numbers", "-5", "3", -1},
		{"floats", "3.14", "2.71", 1},
		{"invalid falls back to string", "abc", "def", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareNumeric(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareNumeric(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestCompareString tests string comparison
func TestCompareString(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"a < b", "alice", "bob", -1},
		{"b > a", "bob", "alice", 1},
		{"equal", "alice", "alice", 0},
		{"case insensitive", "Alice", "alice", 0},
		{"case insensitive ordering", "alice", "Bob", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareString(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareString(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestCompareBool tests boolean comparison
func TestCompareBool(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"false < true", "false", "true", -1},
		{"true > false", "true", "false", 1},
		{"true = true", "true", "true", 0},
		{"false = false", "false", "false", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareBool(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareBool(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
