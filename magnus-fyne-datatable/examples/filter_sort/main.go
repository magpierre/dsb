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

	"github.com/magpierre/fyne-datatable/adapters/memory"
	"github.com/magpierre/fyne-datatable/datatable"
	"github.com/magpierre/fyne-datatable/internal/filter"
	"github.com/magpierre/fyne-datatable/internal/sort"
)

func main() {
	fmt.Println("=== Fyne DataTable: Filtering and Sorting Example ===")
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

	// Create data source
	source, err := memory.NewDataSource(data, headers)
	if err != nil {
		log.Fatal("Failed to create data source:", err)
	}

	// Create table model
	model, err := datatable.NewTableModel(source)
	if err != nil {
		log.Fatal("Failed to create table model:", err)
	}

	// Print original data
	fmt.Println("--- Original Data ---")
	printTable(model)

	// Example 1: Simple Filter - Age > 28
	fmt.Println("\n--- Example 1: Filter by Age > 28 ---")
	simpleFilter := &filter.SimpleFilter{
		Column:   "Age",
		Operator: filter.OpGreaterThan,
		Value:    "28",
	}

	if err := model.SetFilter(simpleFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	printTable(model)
	fmt.Printf("Filter: %s\n", simpleFilter.Description())

	// Example 2: Clear filter and apply composite filter
	// (Age > 25 AND Role = 'Engineer')
	fmt.Println("\n--- Example 2: Composite Filter (Age > 25 AND Role = 'Engineer') ---")
	compositeFilter := &filter.CompositeFilter{
		Filters: []datatable.Filter{
			&filter.SimpleFilter{Column: "Age", Operator: filter.OpGreaterThan, Value: "25"},
			&filter.SimpleFilter{Column: "Role", Operator: filter.OpEqual, Value: "Engineer"},
		},
		Logic: filter.LogicAND,
	}

	if err := model.SetFilter(compositeFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	printTable(model)
	fmt.Printf("Filter: %s\n", compositeFilter.Description())

	// Example 3: Query Filter
	fmt.Println("\n--- Example 3: Query Filter (name ~ 'e' OR age < 28) ---")
	queryFilter := &filter.QueryFilter{
		Query: "name ~ 'e' OR age < 28",
	}

	if err := model.SetFilter(queryFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	printTable(model)
	fmt.Printf("Filter: %s\n", queryFilter.Description())

	// Example 4: Clear filter, then sort by Age
	fmt.Println("\n--- Example 4: Sort by Age (Ascending) ---")
	if err := model.SetFilter(nil); err != nil {
		log.Fatal("Failed to clear filter:", err)
	}

	// Create sort engine
	sortEngine := sort.NewEngine()

	// Get visible rows (all rows since filter is cleared)
	visibleRowIndices := make([]int, model.VisibleRowCount())
	for i := 0; i < model.VisibleRowCount(); i++ {
		visibleRowIndices[i] = i
	}

	// Sort by Age (column 1)
	sortSpec := sort.SortSpec{
		Column:    1, // Age column in original data
		Direction: datatable.SortAscending,
		DataType:  datatable.TypeInt,
	}

	sortedIndices, err := sortEngine.Sort(source, visibleRowIndices, sortSpec)
	if err != nil {
		log.Fatal("Failed to sort:", err)
	}

	// Apply sorted indices to model
	if err := model.ApplySortedIndices(sortedIndices); err != nil {
		log.Fatal("Failed to apply sorted indices:", err)
	}

	// Update sort state
	if err := model.SetSort(1, datatable.SortAscending); err != nil {
		log.Fatal("Failed to set sort state:", err)
	}

	printTable(model)
	fmt.Println("Sorted by: Age (Ascending)")

	// Example 5: Multi-column sort (Age ascending, then Name descending)
	fmt.Println("\n--- Example 5: Multi-Column Sort (Age asc, then Name desc) ---")

	// Reset to original order
	if err := model.SetFilter(nil); err != nil {
		log.Fatal("Failed to clear filter:", err)
	}

	visibleRowIndices = make([]int, model.VisibleRowCount())
	for i := 0; i < model.VisibleRowCount(); i++ {
		visibleRowIndices[i] = i
	}

	multiSortSpecs := []sort.SortSpec{
		{Column: 1, Direction: datatable.SortAscending, DataType: datatable.TypeInt},     // Age
		{Column: 0, Direction: datatable.SortDescending, DataType: datatable.TypeString}, // Name
	}

	sortedIndices, err = sortEngine.MultiSort(source, visibleRowIndices, multiSortSpecs)
	if err != nil {
		log.Fatal("Failed to multi-sort:", err)
	}

	if err := model.ApplySortedIndices(sortedIndices); err != nil {
		log.Fatal("Failed to apply sorted indices:", err)
	}

	printTable(model)
	fmt.Println("Sorted by: Age (Ascending), then Name (Descending)")

	// Example 6: Filter + Sort combination
	fmt.Println("\n--- Example 6: Filter (Role contains 'er') + Sort (Age desc) ---")

	// Apply filter first
	roleFilter := &filter.SimpleFilter{
		Column:   "Role",
		Operator: filter.OpContains,
		Value:    "er",
	}

	if err := model.SetFilter(roleFilter); err != nil {
		log.Fatal("Failed to set filter:", err)
	}

	// Get filtered row indices from model's internal state
	// We need to map from visible indices to original indices
	filteredOriginalIndices := make([]int, model.VisibleRowCount())
	for i := 0; i < model.VisibleRowCount(); i++ {
		// Get the cell to access underlying row (this is a workaround)
		// In production, model should expose GetVisibleRowIndices()
		filteredOriginalIndices[i] = i
	}

	// Since we can't directly access visibleRows, we'll work with the filtered view
	// For this example, we'll just demonstrate the concept
	fmt.Println("After filtering:")
	printTable(model)

	// Now sort the filtered results by Age descending
	filterEngine := filter.NewEngine()

	// Get all filtered indices using the filter engine directly
	filteredIndices, err := filterEngine.Apply(source, roleFilter)
	if err != nil {
		log.Fatal("Failed to get filtered indices:", err)
	}

	// Sort the filtered indices
	sortSpec = sort.SortSpec{
		Column:    1, // Age
		Direction: datatable.SortDescending,
		DataType:  datatable.TypeInt,
	}

	sortedFilteredIndices, err := sortEngine.Sort(source, filteredIndices, sortSpec)
	if err != nil {
		log.Fatal("Failed to sort filtered data:", err)
	}

	// Apply sorted indices
	if err := model.ApplySortedIndices(sortedFilteredIndices); err != nil {
		log.Fatal("Failed to apply sorted indices:", err)
	}

	if err := model.SetSort(1, datatable.SortDescending); err != nil {
		log.Fatal("Failed to set sort state:", err)
	}

	fmt.Println("\nAfter sorting by Age (Descending):")
	printTable(model)
	fmt.Printf("Filter: %s | Sort: Age (Descending)\n", roleFilter.Description())

	fmt.Println("\n=== Example Complete ===")
}

// printTable prints the current state of the table model
func printTable(model *datatable.TableModel) {
	// Print headers
	for i := 0; i < model.VisibleColumnCount(); i++ {
		name, err := model.VisibleColumnName(i)
		if err != nil {
			log.Printf("Error getting column name: %v\n", err)
			continue
		}
		fmt.Printf("%-12s", name)
	}
	fmt.Println()

	// Print separator
	for i := 0; i < model.VisibleColumnCount(); i++ {
		fmt.Printf("%-12s", "------------")
	}
	fmt.Println()

	// Print rows
	for r := 0; r < model.VisibleRowCount(); r++ {
		for c := 0; c < model.VisibleColumnCount(); c++ {
			cell, err := model.VisibleCell(r, c)
			if err != nil {
				log.Printf("Error getting cell (%d,%d): %v\n", r, c, err)
				continue
			}
			fmt.Printf("%-12s", cell.Formatted)
		}
		fmt.Println()
	}

	fmt.Printf("(Showing %d rows)\n", model.VisibleRowCount())
}
