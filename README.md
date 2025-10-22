# DSB - Delta Sharing Browser

A desktop application for browsing and exploring Delta Sharing tables with an intuitive graphical interface.

## Overview

**DSB (Delta Sharing Browser)** is a cross-platform desktop application written in Go that provides a graphical user interface for browsing Delta Sharing datasets. Delta Sharing is an open protocol for secure real-time exchange of large datasets, enabling organizations to share data in Delta Lake and Apache Parquet format without copying it.

This application allows users to:
- Connect to Delta Sharing servers using profile files
- Navigate through shares, schemas, and tables hierarchically
- View and explore table data in a formatted grid interface
- Work with multiple tables simultaneously in a tabbed interface

## Features

### Core Functionality
- **Profile Management**: Load and connect to Delta Sharing endpoints using standard profile files (`.share`)
- **Hierarchical Navigation**: Browse data hierarchically from shares → schemas → tables
- **Data Visualization**: View table data in a responsive, formatted grid with proper type handling
- **Multi-Tab Interface**: Open and work with multiple tables simultaneously
- **Async Operations**: Non-blocking data loading with progress indicators for better UX

### Data Handling
- **Apache Arrow Integration**: Efficient columnar data processing for optimal performance
- **Comprehensive Type Support**: Proper formatting and display for all Apache Arrow data types including:
  - Primitives: integers, floats, booleans, strings, binary data
  - Temporal: dates, timestamps, intervals
  - Complex: structs, lists, decimals
- **Large Dataset Support**: Handles large tables efficiently using Arrow's memory model

### User Interface
- **Modern Design**: Clean, desktop-native appearance using the Adwaita theme
- **Responsive Layout**: Collapsible navigation panels and resizable windows
- **Progressive Disclosure**: Data loads on-demand as you navigate through the hierarchy
- **Visual Feedback**: Progress indicators and informative error dialogs

## Technology Stack

- **Language**: Go 1.22.7+
- **GUI Framework**: [Fyne v2.5.2](https://fyne.io/) - Cross-platform GUI toolkit for Go
- **Theme**: Adwaita theme (via fyne.io/x/fyne)
- **Data Format**: [Apache Arrow v18](https://arrow.apache.org/) for efficient columnar data handling
- **Delta Sharing Client**: Custom Go client library ([go_delta_sharing_client](https://github.com/magpierre/go_delta_sharing_client))
- **License**: Apache License 2.0

## Prerequisites

To build and run DSB, you need:

- **Go**: Version 1.22.7 or higher
- **Operating System**: Linux, macOS, or Windows
- **Build Dependencies**:
  - On Linux: `libgl1-mesa-dev`, `xorg-dev` (for GUI support)
  - On macOS: Xcode command line tools
  - On Windows: gcc (via MinGW-w64)

## Installation

### Building from Source

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/dsb_fyne.git
   cd dsb_fyne
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the application**:
   ```bash
   go build -o dsb
   ```

4. **Run the application**:
   ```bash
   ./dsb
   ```

### Using Pre-built Binary

If a pre-built binary is available for your platform, simply download and run it:

```bash
chmod +x dsb  # On Linux/macOS
./dsb
```

## Usage

### Getting Started

1. **Launch the application**: Run the `dsb` executable

2. **Open a Delta Sharing profile**:
   - Click the "Open File" button in the toolbar
   - Select your Delta Sharing profile file (`.share` format)
   - The profile typically contains connection information including endpoint URL and bearer token

3. **Navigate the data hierarchy**:
   - **Shares**: Select a share from the left panel to view its schemas
   - **Schemas**: Select a schema to view its tables
   - **Tables**: Select a table to load and display its data

4. **View data**:
   - Table data appears in the "Browser" tab
   - Data is displayed in a grid with proper formatting for each data type
   - Long values are truncated for display (hover to see full values)

5. **Work with multiple tables**:
   - Each table opens in its own accordion section
   - Expand/collapse sections to focus on specific tables
   - Close tabs when no longer needed

### Delta Sharing Profile Format

A Delta Sharing profile is a JSON file with the following structure:

```json
{
  "shareCredentialsVersion": 1,
  "endpoint": "https://sharing.example.com/delta-sharing/",
  "bearerToken": "your-token-here"
}
```

## Project Structure

```
dsb_fyne/
├── main.go                          # Application entry point
├── windows/
│   ├── mainWindow.go                # Main application window and navigation
│   ├── dataBrowser.go              # Data display and table rendering
│   └── resources/
│       └── bundled.go              # Bundled resources (icons, images)
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
├── LICENSE                         # Apache 2.0 license
├── README.md                       # This file
├── CLAUDE.md                       # Detailed codebase documentation
└── FyneApp.toml                   # Fyne application metadata
```

### Key Components

#### Main Application (`main.go`)
Minimal entry point that delegates to the windows package.

#### Main Window (`windows/mainWindow.go`)
- Manages the primary application interface and navigation
- Handles profile loading and Delta Sharing client initialization
- Coordinates hierarchical data browsing (shares → schemas → tables)
- Provides split-pane UI with toolbar, navigation panels, and content area

#### Data Browser (`windows/dataBrowser.go`)
- Handles data fetching from Delta Sharing tables
- Converts Apache Arrow records to displayable format
- Creates table widgets with proper type formatting
- Manages multi-table tabbed interface

## Development

### Code Organization

The codebase follows clean architecture principles:
- **Separation of Concerns**: UI logic (mainWindow) separate from data handling (dataBrowser)
- **Resource Management**: Proper cleanup of Apache Arrow records
- **Type Safety**: Comprehensive type handling for all Arrow data types
- **Async Operations**: Goroutines and channels for non-blocking operations

### Key Design Patterns

- **Data Binding**: Reactive UI updates using Fyne's data binding
- **Factory Pattern**: `CreateMainWindow()` for window initialization
- **Progressive Loading**: Data fetched on-demand as user navigates
- **Context Preservation**: Selected state maintained during navigation

### Adding Features

To extend DSB:

1. **New UI Components**: Add to `windows/mainWindow.go` or create new files in `windows/`
2. **Data Processing**: Extend `windows/dataBrowser.go` for new data operations
3. **Type Support**: Modify `parseRecord()` in dataBrowser.go:145-224 for custom type handling
4. **Resources**: Add assets to `windows/resources/` and regenerate with `fyne bundle`

### Running Tests

```bash
go test ./...
```

### Code Style

Follow standard Go conventions:
- Run `gofmt` before committing
- Use meaningful variable names
- Document exported functions and types
- Handle errors appropriately

## Performance Considerations

DSB is optimized for handling large datasets:

- **Apache Arrow**: Columnar memory format reduces memory overhead
- **Lazy Loading**: Data fetched only when needed
- **Async Operations**: UI remains responsive during data loading
- **Efficient Type Conversion**: Direct type mapping from Arrow to display format

## Troubleshooting

### Common Issues

**"Cannot open profile file"**
- Ensure the file is a valid JSON Delta Sharing profile
- Check file permissions

**"Failed to connect to Delta Sharing endpoint"**
- Verify network connectivity
- Check that the endpoint URL is accessible
- Ensure the bearer token is valid and not expired

**"Failed to load table data"**
- Verify you have permissions to access the table
- Check that the table exists and is properly shared
- Look for error messages in the dialog

### Debug Mode

For development, add debug logging:
```go
import "log"
log.Printf("Debug info: %v", variable)
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Delta Sharing**: Open protocol by Databricks for secure data sharing
- **Fyne**: Cross-platform GUI toolkit for Go
- **Apache Arrow**: Columnar in-memory analytics layer
- **Go Community**: For excellent libraries and tools

## Links

- [Delta Sharing Specification](https://github.com/delta-io/delta-sharing)
- [Fyne Documentation](https://developer.fyne.io/)
- [Apache Arrow Go Documentation](https://pkg.go.dev/github.com/apache/arrow/go/v18)
- [go_delta_sharing_client](https://github.com/magpierre/go_delta_sharing_client)

## Status

For issues, feature requests, or questions, please open an issue on GitHub.

---

**Built with ❤️ using Go and Fyne**
