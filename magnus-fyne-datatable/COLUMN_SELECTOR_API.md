# Column Selector API

The fyne-datatable library now includes a built-in ColumnSelector widget that allows users to show/hide table columns dynamically.

## Simple API Usage

### Enable Column Selector

```go
// Create model
model, err := datatable.NewTableModel(source)

// Create config and enable column selector
config := dtwidget.DefaultConfig()
config.ShowColumnSelector = true  // Enable column selection

// Create DataTable with config
table := dtwidget.NewDataTableWithConfig(model, config)
```

### Disable Column Selector (Default)

```go
// Column selector is disabled by default
table := dtwidget.NewDataTable(model)

// Or explicitly disable it
config := dtwidget.DefaultConfig()
config.ShowColumnSelector = false
table := dtwidget.NewDataTableWithConfig(model, config)
```

## Configuration Options

The `Config` struct supports the following options:

```go
type Config struct {
    ShowFilterBar      bool  // Show row filter bar (default: true)
    ShowStatusBar      bool  // Show status bar (default: true)
    ShowColumnSelector bool  // Show column selector (default: false)
    MinColumnWidth     int   // Minimum column width (default: 100)
}
```

## Complete Example

```go
package main

import (
    "log"
    "fyne.io/fyne/v2/app"
    "github.com/magpierre/fyne-datatable/adapters/slice"
    "github.com/magpierre/fyne-datatable/datatable"
    dtwidget "github.com/magpierre/fyne-datatable/widget"
)

func main() {
    myApp := app.New()
    window := myApp.NewWindow("DataTable Demo")

    // Create sample data
    data := [][]interface{}{
        {"Alice", 30, "Engineer"},
        {"Bob", 25, "Designer"},
        {"Charlie", 35, "Manager"},
    }
    headers := []string{"Name", "Age", "Role"}

    // Create data source and model
    source, _ := slice.NewFromInterfaces(data, headers)
    model, _ := datatable.NewTableModel(source)

    // Configure DataTable with column selector enabled
    config := dtwidget.DefaultConfig()
    config.ShowColumnSelector = true
    config.ShowFilterBar = true
    config.ShowStatusBar = true
    
    table := dtwidget.NewDataTableWithConfig(model, config)

    // Add sorting handler
    table.OnHeaderClick(func(col int) {
        // Handle sorting...
    })

    window.SetContent(table)
    window.ShowAndRun()
}
```

## Features

The ColumnSelector widget provides:

- **Accordion UI** - Collapsible section to save screen space
- **Checkboxes** - One checkbox per column
- **Select All** - Button to show all columns
- **Deselect All** - Button to hide all but the first column
- **Scrollable** - Handles tables with many columns
- **Real-time updates** - Table refreshes immediately when columns are toggled

## Advanced Usage

If you need direct access to the ColumnSelector widget for custom functionality:

```go
// The ColumnSelector is accessible through the DataTable's internal fields
// but this is not recommended for normal usage
```

## Migration from Manual Implementation

Before:
```go
// 70+ lines of manual column selector code
columnChecks := make(map[int]*widget.Check)
// ... lots of checkbox creation
// ... manual accordion setup
// ... manual event handlers
```

After:
```go
// One line!
config.ShowColumnSelector = true
```

## Integration with DSB Application

To use in the DSB application (dataBrowser.go):

```go
config := dtwidget.DefaultConfig()
config.ShowFilterBar = true
config.ShowStatusBar = true
config.ShowColumnSelector = true  // Add this line

dataTable := dtwidget.NewDataTableWithConfig(model, config)
```

Then remove the manual column selector code (lines 208-280 in dataBrowser.go).

