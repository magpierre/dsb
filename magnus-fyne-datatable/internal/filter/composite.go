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
	"fmt"
	"strings"

	"github.com/magpierre/fyne-datatable/datatable"
)

// LogicOp represents a logical operator for combining filters.
type LogicOp int

const (
	// LogicAND requires all filters to pass.
	LogicAND LogicOp = iota
	// LogicOR requires at least one filter to pass.
	LogicOR
)

// String returns the string representation of a LogicOp.
func (op LogicOp) String() string {
	switch op {
	case LogicAND:
		return "AND"
	case LogicOR:
		return "OR"
	default:
		return fmt.Sprintf("unknown(%d)", op)
	}
}

// CompositeFilter combines multiple filters with AND or OR logic.
type CompositeFilter struct {
	// Filters is the list of filters to combine.
	Filters []datatable.Filter

	// Logic specifies how to combine the filters (AND or OR).
	Logic LogicOp
}

// Evaluate implements the Filter interface.
func (f *CompositeFilter) Evaluate(row []datatable.Value, columnNames []string) (bool, error) {
	if len(f.Filters) == 0 {
		return true, nil // Empty filter passes all rows
	}

	switch f.Logic {
	case LogicAND:
		// All filters must pass
		for _, filter := range f.Filters {
			passes, err := filter.Evaluate(row, columnNames)
			if err != nil {
				return false, err
			}
			if !passes {
				return false, nil // Short-circuit on first failure
			}
		}
		return true, nil

	case LogicOR:
		// At least one filter must pass
		for _, filter := range f.Filters {
			passes, err := filter.Evaluate(row, columnNames)
			if err != nil {
				return false, err
			}
			if passes {
				return true, nil // Short-circuit on first success
			}
		}
		return false, nil

	default:
		return false, fmt.Errorf("%w: unknown logic operator %d", datatable.ErrInvalidFilter, f.Logic)
	}
}

// Description implements the Filter interface.
func (f *CompositeFilter) Description() string {
	if len(f.Filters) == 0 {
		return "empty filter"
	}

	descriptions := make([]string, len(f.Filters))
	for i, filter := range f.Filters {
		descriptions[i] = filter.Description()
	}

	logicStr := f.Logic.String()
	return "(" + strings.Join(descriptions, " "+logicStr+" ") + ")"
}
