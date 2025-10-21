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
)

func main() {
	// Create sample data
	data := [][]string{
		{"Alice", "30", "Engineer"},
		{"Bob", "25", "Designer"},
		{"Charlie", "35", "Manager"},
		{"Diana", "28", "Developer"},
	}
	columnNames := []string{"Name", "Age", "Role"}

	// Create data source
	source, err := memory.NewDataSource(data, columnNames)
	if err != nil {
		log.Fatal(err)
	}

	// Demonstrate DataSource usage
	fmt.Printf("Data has %d rows and %d columns\n", source.RowCount(), source.ColumnCount())
	fmt.Println()

	// Print column names
	fmt.Print("Columns: ")
	for i := 0; i < source.ColumnCount(); i++ {
		name, _ := source.ColumnName(i)
		fmt.Printf("%s ", name)
	}
	fmt.Println()
	fmt.Println()

	// Print all data
	for row := 0; row < source.RowCount(); row++ {
		rowData, _ := source.Row(row)
		for col, value := range rowData {
			colName, _ := source.ColumnName(col)
			fmt.Printf("%s: %s  ", colName, value.Formatted)
		}
		fmt.Println()
	}
	fmt.Println()

	// Access specific cell
	cell, _ := source.Cell(1, 0)
	fmt.Printf("Cell (1,0): %s (type: %s)\n", cell.Formatted, cell.Type)

	fmt.Println("\n✓ Basic example completed successfully!")
}
