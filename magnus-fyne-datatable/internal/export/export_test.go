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

package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/magpierre/fyne-datatable/adapters/memory"
	"github.com/magpierre/fyne-datatable/datatable"
)

// Helper function to create test data
func createTestData() (datatable.DataSource, error) {
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
		{"Charlie", "35", "Manager"},
	}
	headers := []string{"Name", "Age", "Role"}
	return memory.NewDataSource(data, headers)
}

// TestCSVExport_Basic tests basic CSV export
func TestCSVExport_Basic(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewCSVExporter()
	var buf bytes.Buffer

	rowCount, err := exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if rowCount != 3 {
		t.Errorf("Expected 3 rows exported, got %d", rowCount)
	}

	output := buf.String()

	// Check headers
	if !strings.Contains(output, "Name,Age,Role") {
		t.Errorf("Expected headers in output, got: %s", output)
	}

	// Check data rows
	if !strings.Contains(output, "Alice,30,Engineer") {
		t.Errorf("Expected Alice row in output, got: %s", output)
	}
}

// TestCSVExport_NoHeaders tests CSV export without headers
func TestCSVExport_NoHeaders(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	config := DefaultCSVConfig()
	config.IncludeHeaders = false

	exporter := NewCSVExporterWithConfig(config)
	var buf bytes.Buffer

	_, err = exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()

	// Should NOT contain headers
	if strings.Contains(output, "Name,Age,Role") {
		t.Errorf("Should not have headers in output, got: %s", output)
	}

	// Should contain data
	if !strings.Contains(output, "Alice,30,Engineer") {
		t.Errorf("Expected data row in output, got: %s", output)
	}
}

// TestCSVExport_CustomDelimiter tests CSV export with custom delimiter
func TestCSVExport_CustomDelimiter(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	config := DefaultCSVConfig()
	config.Delimiter = '\t' // Tab-separated

	exporter := NewCSVExporterWithConfig(config)
	var buf bytes.Buffer

	_, err = exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()

	// Check for tab-separated values
	if !strings.Contains(output, "Name\tAge\tRole") {
		t.Errorf("Expected tab-separated headers, got: %s", output)
	}

	if !strings.Contains(output, "Alice\t30\tEngineer") {
		t.Errorf("Expected tab-separated data, got: %s", output)
	}
}

// TestCSVExport_SpecialCharacters tests CSV export with special characters
func TestCSVExport_SpecialCharacters(t *testing.T) {
	data := [][]string{
		{"O'Neill, John", "42", "Manager, Sales"},
		{"Smith \"The Boss\"", "35", "CEO"},
	}
	headers := []string{"Name", "Age", "Title"}

	source, err := memory.NewDataSource(data, headers)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewCSVExporter()
	var buf bytes.Buffer

	_, err = exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()

	// CSV library should handle quoting automatically
	// Commas in data should cause field to be quoted
	if !strings.Contains(output, "O'Neill, John") {
		t.Errorf("Expected quoted field with comma, got: %s", output)
	}
}

// TestJSONExport_Basic tests basic JSON export
func TestJSONExport_Basic(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewJSONExporter()
	var buf bytes.Buffer

	rowCount, err := exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if rowCount != 3 {
		t.Errorf("Expected 3 rows exported, got %d", rowCount)
	}

	// Parse JSON to verify validity
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v\nOutput: %s", err, buf.String())
	}

	// Check structure
	if len(result) != 3 {
		t.Errorf("Expected 3 objects in JSON array, got %d", len(result))
	}

	// Check first object
	if result[0]["Name"] != "Alice" {
		t.Errorf("Expected Name='Alice', got %v", result[0]["Name"])
	}

	if result[0]["Age"] != "30" {
		t.Errorf("Expected Age='30', got %v", result[0]["Age"])
	}
}

// TestJSONExport_PrettyPrint tests JSON export with pretty printing
func TestJSONExport_PrettyPrint(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	config := DefaultJSONConfig()
	config.PrettyPrint = true

	exporter := NewJSONExporterWithConfig(config)
	var buf bytes.Buffer

	_, err = exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	output := buf.String()

	// Pretty print should have newlines
	if !strings.Contains(output, "\n") {
		t.Errorf("Expected newlines in pretty-printed output")
	}

	// Should still be valid JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
}

// TestJSONExport_EmptyData tests JSON export with no data
func TestJSONExport_EmptyData(t *testing.T) {
	data := [][]string{}
	headers := []string{"Name", "Age", "Role"}

	source, err := memory.NewDataSource(data, headers)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewJSONExporter()
	var buf bytes.Buffer

	rowCount, err := exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if rowCount != 0 {
		t.Errorf("Expected 0 rows exported, got %d", rowCount)
	}

	// Should be empty array
	output := buf.String()
	if output != "[]" {
		t.Errorf("Expected '[]', got: %s", output)
	}
}

// TestIterator_Subset tests iterating over a subset of rows
func TestIterator_Subset(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	// Only iterate over rows 0 and 2 (Alice and Charlie)
	visibleRows := []int{0, 2}

	iterator, err := NewModelIterator(source, visibleRows)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewJSONExporter()
	var buf bytes.Buffer

	rowCount, err := exporter.Export(&buf, iterator, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if rowCount != 2 {
		t.Errorf("Expected 2 rows exported, got %d", rowCount)
	}

	// Parse JSON
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Should have Alice and Charlie, but not Bob
	if len(result) != 2 {
		t.Fatalf("Expected 2 objects, got %d", len(result))
	}

	if result[0]["Name"] != "Alice" {
		t.Errorf("Expected first row Name='Alice', got %v", result[0]["Name"])
	}

	if result[1]["Name"] != "Charlie" {
		t.Errorf("Expected second row Name='Charlie', got %v", result[1]["Name"])
	}
}

// TestProgressCallback tests progress callback during export
func TestProgressCallback(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewCSVExporter()
	var buf bytes.Buffer

	callCount := 0
	lastCurrent := 0

	progress := func(current, total int) bool {
		callCount++
		lastCurrent = current

		// Total should be 3
		if total != 3 {
			t.Errorf("Expected total=3, got %d", total)
		}

		// Current should increase
		if current < 1 || current > total {
			t.Errorf("Invalid current value: %d (total=%d)", current, total)
		}

		return true // Continue export
	}

	_, err = exporter.Export(&buf, iterator, progress)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Progress should have been called 3 times (once per row)
	if callCount != 3 {
		t.Errorf("Expected 3 progress callbacks, got %d", callCount)
	}

	// Last current should be 3
	if lastCurrent != 3 {
		t.Errorf("Expected last current=3, got %d", lastCurrent)
	}
}

// TestProgressCallback_Cancel tests cancelling export via progress callback
func TestProgressCallback_Cancel(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	exporter := NewCSVExporter()
	var buf bytes.Buffer

	// Cancel after first row
	progress := func(current, total int) bool {
		return current < 1 // Cancel after first row
	}

	rowCount, err := exporter.Export(&buf, iterator, progress)

	// Should have error about cancellation
	if err == nil {
		t.Error("Expected error when export is cancelled")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected cancellation error, got: %v", err)
	}

	// Should have exported 1 row before cancellation
	if rowCount != 1 {
		t.Errorf("Expected 1 row exported before cancel, got %d", rowCount)
	}
}

// TestExporterMetadata tests exporter metadata methods
func TestExporterMetadata(t *testing.T) {
	csvExporter := NewCSVExporter()

	if csvExporter.FileExtension() != "csv" {
		t.Errorf("Expected extension 'csv', got %s", csvExporter.FileExtension())
	}

	if csvExporter.MimeType() != "text/csv" {
		t.Errorf("Expected MIME type 'text/csv', got %s", csvExporter.MimeType())
	}

	if csvExporter.Description() == "" {
		t.Error("Expected non-empty description")
	}

	jsonExporter := NewJSONExporter()

	if jsonExporter.FileExtension() != "json" {
		t.Errorf("Expected extension 'json', got %s", jsonExporter.FileExtension())
	}

	if jsonExporter.MimeType() != "application/json" {
		t.Errorf("Expected MIME type 'application/json', got %s", jsonExporter.MimeType())
	}
}

// TestEngine_Export tests the export engine
func TestEngine_Export(t *testing.T) {
	source, err := createTestData()
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	iterator, err := NewModelIterator(source, nil)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	engine := NewEngine()
	exporter := NewCSVExporter()
	var buf bytes.Buffer

	rowCount, err := engine.Export(&buf, iterator, exporter, nil)
	if err != nil {
		t.Fatalf("Engine export failed: %v", err)
	}

	if rowCount != 3 {
		t.Errorf("Expected 3 rows exported, got %d", rowCount)
	}
}

// TestEngine_ValidateExporter tests exporter validation
func TestEngine_ValidateExporter(t *testing.T) {
	engine := NewEngine()

	// Test with nil exporter
	err := engine.ValidateExporter(nil)
	if err == nil {
		t.Error("Expected error for nil exporter")
	}

	// Test with valid exporter
	exporter := NewCSVExporter()
	err = engine.ValidateExporter(exporter)
	if err != nil {
		t.Errorf("Expected no error for valid exporter, got: %v", err)
	}
}
