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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/magpierre/fyne-datatable/adapters/memory"
	"github.com/magpierre/fyne-datatable/datatable"
	"github.com/magpierre/fyne-datatable/internal/export"
	"github.com/magpierre/fyne-datatable/internal/filter"
	"github.com/magpierre/fyne-datatable/internal/sort"
)

func main() {
	fmt.Println("=== Fyne DataTable: Export Example ===")
	fmt.Println()

	// Create sample data
	data := [][]string{
		{"Alice", "30", "Engineer", "2024-01-15"},
		{"Bob", "25", "Designer", "2024-03-20"},
		{"Charlie", "35", "Manager", "2024-02-10"},
		{"Diana", "28", "Developer", "2024-01-05"},
		{"Eve", "32", "Engineer", "2024-02-20"},
		{"Frank", "27", "Designer", "2024-03-01"},
	}

	headers := []string{"Name", "Age", "Role", "JoinDate"}

	// Create data source and model
	source, err := memory.NewDataSource(data, headers)
	if err != nil {
		log.Fatal("Failed to create data source:", err)
	}

	model, err := datatable.NewTableModel(source)
	if err != nil {
		log.Fatal("Failed to create table model:", err)
	}

	// Example 1: Export all data to CSV
	fmt.Println("--- Example 1: Export All Data to CSV ---")
	if err := exportToCSV(source, nil, "output_all.csv"); err != nil {
		log.Printf("CSV export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported all data to output_all.csv")
	}

	// Example 2: Export all data to JSON
	fmt.Println("\n--- Example 2: Export All Data to JSON ---")
	if err := exportToJSON(source, nil, "output_all.json", false); err != nil {
		log.Printf("JSON export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported all data to output_all.json")
	}

	// Example 3: Export to pretty-printed JSON
	fmt.Println("\n--- Example 3: Export to Pretty-Printed JSON ---")
	if err := exportToJSON(source, nil, "output_pretty.json", true); err != nil {
		log.Printf("JSON export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported all data to output_pretty.json (formatted)")
	}

	// Example 4: Export filtered data
	fmt.Println("\n--- Example 4: Export Filtered Data (Engineers only) ---")

	// Apply filter to model
	engineerFilter := &filter.SimpleFilter{
		Column:   "Role",
		Operator: filter.OpEqual,
		Value:    "Engineer",
	}

	if err := model.SetFilter(engineerFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	fmt.Printf("Filtered to %d rows (Engineers only)\n", model.VisibleRowCount())

	// Export filtered data
	visibleRows := model.GetVisibleRowIndices()
	if err := exportToCSV(source, visibleRows, "output_engineers.csv"); err != nil {
		log.Printf("CSV export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported filtered data to output_engineers.csv")
	}

	// Example 5: Export filtered and sorted data
	fmt.Println("\n--- Example 5: Export Filtered + Sorted Data ---")

	// Clear previous filter
	model.SetFilter(nil)

	// Filter: Age > 27
	ageFilter := &filter.SimpleFilter{
		Column:   "Age",
		Operator: filter.OpGreaterThan,
		Value:    "27",
	}

	if err := model.SetFilter(ageFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	// Get filtered indices
	filterEngine := filter.NewEngine()
	filteredIndices, err := filterEngine.Apply(source, ageFilter)
	if err != nil {
		log.Fatal("Failed to apply filter:", err)
	}

	// Sort by Age descending
	sortEngine := sort.NewEngine()
	sortSpec := sort.SortSpec{
		Column:    1, // Age column
		Direction: datatable.SortDescending,
		DataType:  datatable.TypeInt,
	}

	sortedIndices, err := sortEngine.Sort(source, filteredIndices, sortSpec)
	if err != nil {
		log.Fatal("Failed to sort:", err)
	}

	fmt.Printf("Filtered to %d rows (Age > 27), sorted by Age descending\n", len(sortedIndices))

	if err := exportToCSV(source, sortedIndices, "output_filtered_sorted.csv"); err != nil {
		log.Printf("CSV export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported filtered + sorted data to output_filtered_sorted.csv")
	}

	// Example 6: Export with progress callback
	fmt.Println("\n--- Example 6: Export with Progress Tracking ---")

	model.SetFilter(nil) // Reset filter
	visibleRows = model.GetVisibleRowIndices()

	if err := exportWithProgress(source, visibleRows, "output_progress.csv"); err != nil {
		log.Printf("CSV export with progress failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported with progress tracking to output_progress.csv")
	}

	// Example 7: Export to TSV (Tab-Separated Values)
	fmt.Println("\n--- Example 7: Export to TSV ---")
	if err := exportToTSV(source, nil, "output.tsv"); err != nil {
		log.Printf("TSV export failed: %v\n", err)
	} else {
		fmt.Println("✓ Exported all data to output.tsv")
	}

	fmt.Println("\n=== Export Complete ===")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - output_all.csv (all data, CSV)")
	fmt.Println("  - output_all.json (all data, JSON compact)")
	fmt.Println("  - output_pretty.json (all data, JSON formatted)")
	fmt.Println("  - output_engineers.csv (filtered: Engineers)")
	fmt.Println("  - output_filtered_sorted.csv (filtered: Age>27, sorted)")
	fmt.Println("  - output_progress.csv (with progress tracking)")
	fmt.Println("  - output.tsv (tab-separated)")
}

// exportToCSV exports data to a CSV file
func exportToCSV(source datatable.DataSource, visibleRows []int, filename string) error {
	// Create iterator
	iterator, err := export.NewModelIterator(source, visibleRows)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}

	// Create exporter
	exporter := export.NewCSVExporter()

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Export
	engine := export.NewEngine()
	rowCount, err := engine.Export(file, iterator, exporter, nil)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Printf("  → Exported %d rows\n", rowCount)
	return nil
}

// exportToJSON exports data to a JSON file
func exportToJSON(source datatable.DataSource, visibleRows []int, filename string, pretty bool) error {
	// Create iterator
	iterator, err := export.NewModelIterator(source, visibleRows)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}

	// Create exporter with configuration
	config := export.DefaultJSONConfig()
	config.PrettyPrint = pretty
	exporter := export.NewJSONExporterWithConfig(config)

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Export
	engine := export.NewEngine()
	rowCount, err := engine.Export(file, iterator, exporter, nil)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Printf("  → Exported %d rows\n", rowCount)
	return nil
}

// exportToTSV exports data to a TSV file
func exportToTSV(source datatable.DataSource, visibleRows []int, filename string) error {
	// Create iterator
	iterator, err := export.NewModelIterator(source, visibleRows)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}

	// Create exporter with tab delimiter
	config := export.DefaultCSVConfig()
	config.Delimiter = '\t'
	exporter := export.NewCSVExporterWithConfig(config)

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Export
	engine := export.NewEngine()
	rowCount, err := engine.Export(file, iterator, exporter, nil)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Printf("  → Exported %d rows\n", rowCount)
	return nil
}

// exportWithProgress exports data with progress tracking
func exportWithProgress(source datatable.DataSource, visibleRows []int, filename string) error {
	// Create iterator
	iterator, err := export.NewModelIterator(source, visibleRows)
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}

	// Create exporter
	exporter := export.NewCSVExporter()

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Progress callback
	progressCallback := func(current, total int) bool {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\r  → Progress: %d/%d rows (%.1f%%)", current, total, percent)
		return true // Continue export
	}

	// Export
	engine := export.NewEngine()
	rowCount, err := engine.Export(file, iterator, exporter, progressCallback)
	fmt.Println() // New line after progress

	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	fmt.Printf("  → Exported %d rows total\n", rowCount)
	return nil
}
