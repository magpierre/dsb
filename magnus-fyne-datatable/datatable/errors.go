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
