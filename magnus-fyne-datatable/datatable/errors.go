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

import "errors"

// Common errors returned by the datatable package.
var (
	// ErrInvalidColumn is returned when a column index is out of range.
	ErrInvalidColumn = errors.New("invalid column index")

	// ErrInvalidRow is returned when a row index is out of range.
	ErrInvalidRow = errors.New("invalid row index")

	// ErrInvalidFilter is returned when a filter expression is invalid.
	ErrInvalidFilter = errors.New("invalid filter expression")

	// ErrTypeMismatch is returned when a type comparison is invalid.
	ErrTypeMismatch = errors.New("type mismatch in comparison")

	// ErrNoDataSource is returned when a required data source is nil.
	ErrNoDataSource = errors.New("data source is nil")

	// ErrEmptyData is returned when data is empty where it shouldn't be.
	ErrEmptyData = errors.New("data is empty")

	// ErrColumnNotFound is returned when a column name is not found.
	ErrColumnNotFound = errors.New("column not found")

	// ErrInvalidSortColumn is returned when trying to sort by an invalid column.
	ErrInvalidSortColumn = errors.New("invalid sort column")

	// ErrExportFailed is returned when export operation fails.
	ErrExportFailed = errors.New("export failed")
)
