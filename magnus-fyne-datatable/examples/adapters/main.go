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
	"strings"

	"github.com/magpierre/fyne-datatable/adapters/csv"
	"github.com/magpierre/fyne-datatable/adapters/memory"
	"github.com/magpierre/fyne-datatable/adapters/slice"
	"github.com/magpierre/fyne-datatable/datatable"
)

func main() {
	fmt.Println("=== Fyne DataTable: Adapters Example ===")
	fmt.Println()

	// Example 1: Memory Adapter (from Phase 1)
	fmt.Println("--- Example 1: Memory Adapter ---")
	memoryExample()

	// Example 2: Slice Adapter (from interfaces)
	fmt.Println("\n--- Example 2: Slice Adapter (from interfaces) ---")
	sliceInterfaceExample()

	// Example 3: Slice Adapter (from strings)
	fmt.Println("\n--- Example 3: Slice Adapter (from strings) ---")
	sliceStringExample()

	// Example 4: Slice Adapter (from maps)
	fmt.Println("\n--- Example 4: Slice Adapter (from maps) ---")
	sliceMapExample()

	// Example 5: CSV Adapter (from file)
	fmt.Println("\n--- Example 5: CSV Adapter (from file) ---")
	csvFileExample()

	// Example 6: CSV Adapter (from string)
	fmt.Println("\n--- Example 6: CSV Adapter (from string/reader) ---")
	csvReaderExample()

	// Example 7: CSV Adapter with custom configuration
	fmt.Println("\n--- Example 7: CSV Adapter (TSV - tab-separated) ---")
	tsvExample()

	fmt.Println("\n=== Examples Complete ===")
}

func memoryExample() {
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
	}
	headers := []string{"Name", "Age", "Role"}

	source, err := memory.NewDataSource(data, headers)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "Memory Adapter")
}

func sliceInterfaceExample() {
	// Mixed types: string, int, float, bool
	data := [][]interface{}{
		{"Alice", 30, 50000.50, true},
		{"Bob", 25, 45000.00, false},
		{"Charlie", 35, 60000.75, true},
	}
	headers := []string{"Name", "Age", "Salary", "Active"}

	source, err := slice.NewFromInterfaces(data, headers)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "Slice Adapter (Interfaces)")
	printColumnTypes(source)
}

func sliceStringExample() {
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
	}
	headers := []string{"Name", "Age", "Role"}

	source, err := slice.NewFromStrings(data, headers)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "Slice Adapter (Strings)")
}

func sliceMapExample() {
	data := []map[string]interface{}{
		{"Name": "Alice", "Age": 30, "Role": "Engineer", "Active": true},
		{"Name": "Bob", "Age": 25, "Role": "Designer", "Active": false},
		{"Name": "Charlie", "Age": 35, "Role": "Manager", "Active": true},
	}

	source, err := slice.NewFromMaps(data)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "Slice Adapter (Maps)")
}

func csvFileExample() {
	// Create a sample CSV file
	csvContent := `Name,Age,Role,Salary
Alice,30,Engineer,50000.50
Bob,25,Designer,45000.00
Charlie,35,Manager,60000.75
Diana,28,Developer,48000.25`

	// Write to file
	filename := "sample_data.csv"
	if err := os.WriteFile(filename, []byte(csvContent), 0644); err != nil {
		log.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer os.Remove(filename) // Clean up

	// Load from file
	source, err := csv.NewFromFile(filename, csv.DefaultConfig())
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "CSV Adapter (from file)")
	printColumnTypes(source)
	fmt.Println("✓ Loaded data from sample_data.csv")
}

func csvReaderExample() {
	csvData := `Name,Age,Department
Alice,30,Engineering
Bob,25,Design
Charlie,35,Management`

	reader := strings.NewReader(csvData)
	source, err := csv.NewFromReader(reader, csv.DefaultConfig())
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "CSV Adapter (from reader)")
}

func tsvExample() {
	tsvData := "Name\tAge\tRole\nAlice\t30\tEngineer\nBob\t25\tDesigner"

	config := csv.DefaultConfig()
	config.Delimiter = '\t' // Tab-separated

	reader := strings.NewReader(tsvData)
	source, err := csv.NewFromReader(reader, config)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	printDataSource(source, "TSV Adapter (tab-separated)")
}

// Helper function to print a DataSource
func printDataSource(source datatable.DataSource, title string) {
	fmt.Printf("\n%s:\n", title)

	// Print headers
	for i := 0; i < source.ColumnCount(); i++ {
		name, _ := source.ColumnName(i)
		fmt.Printf("%-15s", name)
	}
	fmt.Println()

	// Print separator
	for i := 0; i < source.ColumnCount(); i++ {
		fmt.Printf("%-15s", "---------------")
	}
	fmt.Println()

	// Print rows
	for r := 0; r < source.RowCount(); r++ {
		for c := 0; c < source.ColumnCount(); c++ {
			cell, err := source.Cell(r, c)
			if err != nil {
				fmt.Printf("%-15s", "ERROR")
				continue
			}
			if cell.IsNull {
				fmt.Printf("%-15s", "NULL")
			} else {
				fmt.Printf("%-15s", cell.Formatted)
			}
		}
		fmt.Println()
	}

	fmt.Printf("\n(%d rows, %d columns)\n", source.RowCount(), source.ColumnCount())
}

// Helper function to print column types
func printColumnTypes(source datatable.DataSource) {
	fmt.Println("\nColumn Types:")
	for i := 0; i < source.ColumnCount(); i++ {
		name, _ := source.ColumnName(i)
		colType, _ := source.ColumnType(i)
		fmt.Printf("  %s: %s\n", name, colType)
	}
}
