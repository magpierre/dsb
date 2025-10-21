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

package windows

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

// ExportFormat represents the supported export formats
type ExportFormat int

const (
	FormatParquet ExportFormat = iota
	FormatCSV
	FormatJSON
)

// ExportToParquet exports the Arrow table to a Parquet file
func ExportToParquet(table arrow.Table, filePath string) error {
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer file.Close()

	// Create Parquet writer properties
	props := parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy))
	arrowProps := pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema())

	// Create a Parquet file writer
	writer, err := pqarrow.NewFileWriter(table.Schema(), file, props, arrowProps)
	if err != nil {
		return fmt.Errorf("failed to create parquet writer: %w", err)
	}
	defer writer.Close()

	// Write the table
	err = writer.WriteTable(table, table.NumRows())
	if err != nil {
		return fmt.Errorf("failed to write table to parquet: %w", err)
	}

	return nil
}

// ExportToCSV exports the Arrow table to a CSV file
func ExportToCSV(table arrow.Table, filePath string) error {
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	schema := table.Schema()
	headers := make([]string, schema.NumFields())
	for i, field := range schema.Fields() {
		headers[i] = field.Name
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Read records from table
	tr := array.NewTableReader(table, table.NumRows())
	defer tr.Release()

	// Process each record
	for tr.Next() {
		rec := tr.Record()
		numRows := rec.NumRows()

		// Process each row
		for rowIdx := int64(0); rowIdx < numRows; rowIdx++ {
			row := make([]string, rec.NumCols())

			// Process each column
			for colIdx, col := range rec.Columns() {
				row[colIdx] = formatValue(col, int(rowIdx))
			}

			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	if tr.Err() != nil {
		return fmt.Errorf("error reading table: %w", tr.Err())
	}

	return nil
}

// ExportToJSON exports the Arrow table to a JSON file
func ExportToJSON(table arrow.Table, filePath string) error {
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Read records from table
	tr := array.NewTableReader(table, table.NumRows())
	defer tr.Release()

	// Collect all records into a slice of maps
	var records []map[string]interface{}
	schema := table.Schema()

	for tr.Next() {
		rec := tr.Record()
		numRows := rec.NumRows()

		// Process each row
		for rowIdx := int64(0); rowIdx < numRows; rowIdx++ {
			record := make(map[string]interface{})

			// Process each column
			for colIdx, col := range rec.Columns() {
				fieldName := schema.Field(colIdx).Name
				record[fieldName] = getTypedValue(col, int(rowIdx))
			}

			records = append(records, record)
		}
	}

	if tr.Err() != nil {
		return fmt.Errorf("error reading table: %w", tr.Err())
	}

	// Encode to JSON with indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// formatValue converts an Arrow column value at a specific position to a string
func formatValue(col arrow.Array, pos int) string {
	if col.IsNull(pos) {
		return ""
	}

	switch col.DataType().ID() {
	case arrow.STRUCT:
		s := col.(*array.Struct)
		b, _ := s.MarshalJSON()
		return string(b)

	case arrow.LIST:
		as := array.NewSlice(col, int64(pos), int64(pos+1))
		return fmt.Sprintf("%v", as)

	case arrow.STRING:
		s := col.(*array.String)
		return s.Value(pos)

	case arrow.BINARY:
		b := col.(*array.Binary)
		return string(b.Value(pos))

	case arrow.BOOL:
		b := col.(*array.Boolean)
		return fmt.Sprintf("%v", b.Value(pos))

	case arrow.DATE32:
		d32 := col.(*array.Date32)
		return d32.Value(pos).ToTime().Format("2006-01-02")

	case arrow.DATE64:
		d64 := col.(*array.Date64)
		return d64.Value(pos).ToTime().Format("2006-01-02")

	case arrow.DECIMAL128:
		d128 := col.(*array.Decimal128)
		return d128.Value(pos).BigInt().String()

	case arrow.INT8:
		i8 := col.(*array.Int8)
		return fmt.Sprintf("%d", i8.Value(pos))

	case arrow.INT16:
		i16 := col.(*array.Int16)
		return fmt.Sprintf("%d", i16.Value(pos))

	case arrow.INT32:
		i32 := col.(*array.Int32)
		return fmt.Sprintf("%d", i32.Value(pos))

	case arrow.INT64:
		i64 := col.(*array.Int64)
		return fmt.Sprintf("%d", i64.Value(pos))

	case arrow.UINT8:
		u8 := col.(*array.Uint8)
		return fmt.Sprintf("%d", u8.Value(pos))

	case arrow.UINT16:
		u16 := col.(*array.Uint16)
		return fmt.Sprintf("%d", u16.Value(pos))

	case arrow.UINT32:
		u32 := col.(*array.Uint32)
		return fmt.Sprintf("%d", u32.Value(pos))

	case arrow.UINT64:
		u64 := col.(*array.Uint64)
		return fmt.Sprintf("%d", u64.Value(pos))

	case arrow.FLOAT16:
		f16 := col.(*array.Float16)
		return f16.Value(pos).String()

	case arrow.FLOAT32:
		f32 := col.(*array.Float32)
		return fmt.Sprintf("%.6f", f32.Value(pos))

	case arrow.FLOAT64:
		f64 := col.(*array.Float64)
		return fmt.Sprintf("%.6f", f64.Value(pos))

	case arrow.TIMESTAMP:
		ts := col.(*array.Timestamp)
		return ts.Value(pos).ToTime(arrow.Nanosecond).Format("2006-01-02 15:04:05.999999999")

	case arrow.INTERVAL_MONTHS:
		intV := col.(*array.MonthInterval)
		return fmt.Sprintf("%v", intV.Value(pos))

	case arrow.INTERVAL_DAY_TIME:
		intV := col.(*array.DayTimeInterval)
		return fmt.Sprintf("%v", intV.Value(pos))

	default:
		return fmt.Sprintf("%v", col)
	}
}

// getTypedValue returns the typed value for JSON export (preserves types)
func getTypedValue(col arrow.Array, pos int) interface{} {
	if col.IsNull(pos) {
		return nil
	}

	switch col.DataType().ID() {
	case arrow.STRING:
		s := col.(*array.String)
		return s.Value(pos)

	case arrow.BINARY:
		b := col.(*array.Binary)
		return string(b.Value(pos))

	case arrow.BOOL:
		b := col.(*array.Boolean)
		return b.Value(pos)

	case arrow.INT8:
		i8 := col.(*array.Int8)
		return i8.Value(pos)

	case arrow.INT16:
		i16 := col.(*array.Int16)
		return i16.Value(pos)

	case arrow.INT32:
		i32 := col.(*array.Int32)
		return i32.Value(pos)

	case arrow.INT64:
		i64 := col.(*array.Int64)
		return i64.Value(pos)

	case arrow.UINT8:
		u8 := col.(*array.Uint8)
		return u8.Value(pos)

	case arrow.UINT16:
		u16 := col.(*array.Uint16)
		return u16.Value(pos)

	case arrow.UINT32:
		u32 := col.(*array.Uint32)
		return u32.Value(pos)

	case arrow.UINT64:
		u64 := col.(*array.Uint64)
		return u64.Value(pos)

	case arrow.FLOAT16:
		f16 := col.(*array.Float16)
		return f16.Value(pos).Float32()

	case arrow.FLOAT32:
		f32 := col.(*array.Float32)
		return f32.Value(pos)

	case arrow.FLOAT64:
		f64 := col.(*array.Float64)
		return f64.Value(pos)

	case arrow.DATE32:
		d32 := col.(*array.Date32)
		return d32.Value(pos).ToTime().Format("2006-01-02")

	case arrow.DATE64:
		d64 := col.(*array.Date64)
		return d64.Value(pos).ToTime().Format("2006-01-02")

	case arrow.TIMESTAMP:
		ts := col.(*array.Timestamp)
		return ts.Value(pos).ToTime(arrow.Nanosecond).Format("2006-01-02T15:04:05.999999999Z")

	case arrow.STRUCT:
		s := col.(*array.Struct)
		b, _ := s.MarshalJSON()
		var result interface{}
		json.Unmarshal(b, &result)
		return result

	case arrow.LIST:
		as := array.NewSlice(col, int64(pos), int64(pos+1))
		return fmt.Sprintf("%v", as)

	case arrow.DECIMAL128:
		d128 := col.(*array.Decimal128)
		return d128.Value(pos).BigInt().String()

	default:
		return formatValue(col, pos)
	}
}
