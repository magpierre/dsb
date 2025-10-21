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
	"testing"
)

func TestDataType_String(t *testing.T) {
	tests := []struct {
		name string
		dt   DataType
		want string
	}{
		{"TypeString", TypeString, "String"},
		{"TypeInt", TypeInt, "Int"},
		{"TypeFloat", TypeFloat, "Float"},
		{"TypeBool", TypeBool, "Bool"},
		{"TypeDate", TypeDate, "Date"},
		{"TypeTimestamp", TypeTimestamp, "Timestamp"},
		{"TypeBinary", TypeBinary, "Binary"},
		{"TypeDecimal", TypeDecimal, "Decimal"},
		{"TypeStruct", TypeStruct, "Struct"},
		{"TypeList", TypeList, "List"},
		{"Unknown", DataType(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.want {
				t.Errorf("DataType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      interface{}
		dataType DataType
		wantNull bool
	}{
		{"String value", "hello", TypeString, false},
		{"Int value", 42, TypeInt, false},
		{"Nil value", nil, TypeString, true},
		{"Empty string", "", TypeString, false},
		{"Zero int", 0, TypeInt, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValue(tt.raw, tt.dataType)
			if got.IsNull != tt.wantNull {
				t.Errorf("NewValue().IsNull = %v, want %v", got.IsNull, tt.wantNull)
			}
			if got.Type != tt.dataType {
				t.Errorf("NewValue().Type = %v, want %v", got.Type, tt.dataType)
			}
			if !tt.wantNull && got.Raw != tt.raw {
				t.Errorf("NewValue().Raw = %v, want %v", got.Raw, tt.raw)
			}
		})
	}
}

func TestNewNullValue(t *testing.T) {
	v := NewNullValue(TypeString)
	if !v.IsNull {
		t.Error("NewNullValue() should create null value")
	}
	if v.Raw != nil {
		t.Error("NewNullValue().Raw should be nil")
	}
	if v.Formatted != "" {
		t.Error("NewNullValue().Formatted should be empty")
	}
}

func TestSortDirection_String(t *testing.T) {
	tests := []struct {
		name string
		sd   SortDirection
		want string
	}{
		{"SortNone", SortNone, "None"},
		{"SortAscending", SortAscending, "Ascending"},
		{"SortDescending", SortDescending, "Descending"},
		{"Unknown", SortDirection(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sd.String(); got != tt.want {
				t.Errorf("SortDirection.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortState_IsSorted(t *testing.T) {
	tests := []struct {
		name  string
		state SortState
		want  bool
	}{
		{"Not sorted", SortState{Column: -1, Direction: SortNone}, false},
		{"Ascending", SortState{Column: 0, Direction: SortAscending}, true},
		{"Descending", SortState{Column: 2, Direction: SortDescending}, true},
		{"Column set but direction none", SortState{Column: 1, Direction: SortNone}, false},
		{"Direction set but column invalid", SortState{Column: -1, Direction: SortAscending}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsSorted(); got != tt.want {
				t.Errorf("SortState.IsSorted() = %v, want %v", got, tt.want)
			}
		})
	}
}
