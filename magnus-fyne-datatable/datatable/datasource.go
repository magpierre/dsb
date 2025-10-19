package datatable

// DataSource provides read-only access to tabular data.
// Implementations must be thread-safe for concurrent reads.
// All methods should return errors rather than panic.
type DataSource interface {
	// RowCount returns the total number of rows in the data source.
	RowCount() int

	// ColumnCount returns the total number of columns in the data source.
	ColumnCount() int

	// ColumnName returns the name of the column at the given index.
	// Returns ErrInvalidColumn if col is out of range.
	ColumnName(col int) (string, error)

	// ColumnType returns the data type of the column at the given index.
	// Returns ErrInvalidColumn if col is out of range.
	ColumnType(col int) (DataType, error)

	// Cell returns the value at the specified row and column.
	// Returns ErrInvalidRow if row is out of range.
	// Returns ErrInvalidColumn if col is out of range.
	Cell(row, col int) (Value, error)

	// Row returns all values for the specified row.
	// Returns ErrInvalidRow if row is out of range.
	Row(row int) ([]Value, error)

	// Metadata returns optional metadata about the data source.
	// Returns an empty Metadata map if no metadata is available.
	Metadata() Metadata
}
