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
	"context"
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

// TappableListItem is a label that supports both regular click and right-click
type TappableListItem struct {
	widget.Label
	onRightClick func(widget.ListItemID, *fyne.PointEvent)
	onTap        func(widget.ListItemID)
	itemID       widget.ListItemID
}

func newTappableListItem(onRightClick func(widget.ListItemID, *fyne.PointEvent)) *TappableListItem {
	item := &TappableListItem{
		onRightClick: onRightClick,
		itemID:       -1,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (t *TappableListItem) SetItemID(id widget.ListItemID) {
	t.itemID = id
}

func (t *TappableListItem) SetOnTap(callback func(widget.ListItemID)) {
	t.onTap = callback
}

// Tapped handles regular left-click
func (t *TappableListItem) Tapped(e *fyne.PointEvent) {
	if t.onTap != nil && t.itemID >= 0 {
		t.onTap(t.itemID)
	}
}

// TappedSecondary handles right-click
func (t *TappableListItem) TappedSecondary(e *fyne.PointEvent) {
	if t.onRightClick != nil && t.itemID >= 0 {
		t.onRightClick(t.itemID, e)
	}
}

// TappableTreeNode is a container that supports right-click for tree nodes
type TappableTreeNode struct {
	widget.BaseWidget
	content      *fyne.Container
	nodeID       widget.TreeNodeID
	onRightClick func(widget.TreeNodeID, *fyne.PointEvent)
	treeWidget   *widget.Tree
}

func newTappableTreeNode(content *fyne.Container, nodeID widget.TreeNodeID, onRightClick func(widget.TreeNodeID, *fyne.PointEvent)) *TappableTreeNode {
	t := &TappableTreeNode{
		content:      content,
		nodeID:       nodeID,
		onRightClick: onRightClick,
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TappableTreeNode) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

// Tapped handles regular left-click - pass through to tree widget
func (t *TappableTreeNode) Tapped(e *fyne.PointEvent) {
	// Trigger tree selection for this node
	if t.treeWidget != nil && t.nodeID != "" {
		t.treeWidget.Select(t.nodeID)
	}
}

func (t *TappableTreeNode) TappedSecondary(e *fyne.PointEvent) {
	if t.onRightClick != nil {
		t.onRightClick(t.nodeID, e)
	}
}

func (t *TappableTreeNode) UpdateContent(content *fyne.Container) {
	t.content = content
	t.Refresh()
}

func (t *TappableTreeNode) UpdateNodeID(nodeID widget.TreeNodeID) {
	t.nodeID = nodeID
}

func (t *TappableTreeNode) SetTreeWidget(tree *widget.Tree) {
	t.treeWidget = tree
}

type Selected struct {
	share      string
	schema     string
	table      delta_sharing.Table
	table_name string
}
type MainWindow struct {
	a                        fyne.App
	w                        fyne.Window
	top, left, right, bottom fyne.CanvasObject
	profile                  string
	files                    []string
	selected                 Selected
	docTabs                  *container.DocTabs
	dataBrowser              *DataBrowser
	goEditor                 *GoEditor
	statusBar                *widget.Label
	exportButton             *widget.Button
	toolbar                  *widget.Toolbar
	themeManager             *ThemeManager
	navTree                  *NavigationTree
	treeWidget               *widget.Tree
	// Go Editor toolbar buttons container
	goEditorButtonsContainer *fyne.Container
}

func CreateMainWindow() *MainWindow {
	var v MainWindow
	v.NewMainWindow()
	return &v
}

func (t *MainWindow) OpenFile() {
	d := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}

		content, err := io.ReadAll(uc)
		if err != nil {
			dialog.ShowError(err, t.w)
			return
		}

		t.profile = string(content)
		t.SetStatus("Loading profile...")

		// Initialize navigation tree with shares
		err = t.navTree.LoadShares(t.profile)
		if err != nil {
			t.SetStatus("Error loading shares")
			dialog.ShowError(err, t.w)
			return
		}

		t.files = make([]string, 0)
		t.selected = Selected{}

		// Refresh tree widget to show shares
		if t.treeWidget != nil {
			t.treeWidget.Refresh()
		}

		t.w.Content().Refresh()
		t.SetStatus("Profile loaded successfully")
	}, t.w)
	d.Show()
}

func (t *MainWindow) OpenProfile() {
	var pd *ProfileDialog
	pd = NewProfileDialog(t.w, t.a, func(content string, err error) {
		if err != nil {
			t.SetStatus("Error opening file")
			dialog.ShowError(err, t.w)
			return
		}

		if content == "" {
			return
		}

		// Get file path from dialog
		filePath := pd.filePath

		// Detect file type
		fileType := DetectFileType(filePath, content)

		switch fileType {
		case FileTypeCSV, FileTypeParquet, FileTypeJSON:
			// Handle data files
			t.handleDataFileLoad(filePath)

		case FileTypeDeltaSharingProfile:
			// Handle Delta Sharing profile
			t.SetStatus("Loading profile...")
			t.profile = content

			// Initialize navigation tree with shares
			err = t.navTree.LoadShares(t.profile)
			if err != nil {
				t.SetStatus("Error loading shares")
				dialog.ShowError(err, t.w)
				return
			}

			t.files = make([]string, 0)
			t.selected = Selected{}

			// Refresh tree widget to show shares
			if t.treeWidget != nil {
				t.treeWidget.Refresh()
			}

			t.w.Content().Refresh()
			t.SetStatus("Profile loaded successfully")

		default:
			t.SetStatus("Unknown file type")
			dialog.ShowError(fmt.Errorf("unsupported file type"), t.w)
		}
	})
	pd.Show()
}

// SetStatus updates the status bar message
func (t *MainWindow) SetStatus(message string) {
	if t.statusBar != nil {
		t.statusBar.SetText(message)
	}
}

func (t *MainWindow) NewMainWindow() {
	t.selected = Selected{}
	t.a = app.NewWithID("dsb")

	// Initialize theme manager and set theme
	t.themeManager = NewThemeManager(t.a)
	t.a.Settings().SetTheme(t.themeManager.GetCurrentTheme())

	t.toolbar = widget.NewToolbar()
	t.top = t.toolbar
	t.left = container.NewVBox()
	t.right = container.NewVBox()

	// Create status bar
	t.statusBar = widget.NewLabel("Ready")
	t.statusBar.TextStyle = fyne.TextStyle{Italic: true}
	t.bottom = container.NewHBox(t.statusBar)

	// Initialize navigation tree
	t.navTree = NewNavigationTree(t)

	t.w = t.a.NewWindow("Delta Sharing Browser")
	t.w.Resize(fyne.NewSize(700, 600))

	// Set up drag and drop handler
	t.w.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			uri := uris[0]
			// You can now use the uri to access the dropped file
			fileContent, err := storage.Reader(uri)
			if err != nil {
				t.SetStatus("Error reading file")
				return
			}
			defer fileContent.Close()

			// Read the content and convert to string
			content, err := io.ReadAll(fileContent)
			if err != nil {
				t.SetStatus("Error reading file content")
				return
			}
			t.profile = string(content)
			t.SetStatus("Loading profile...")

			// Initialize navigation tree with shares
			err = t.navTree.LoadShares(t.profile)
			if err != nil {
				t.SetStatus("Error loading shares")
				dialog.ShowError(err, t.w)
				return
			}

			t.files = make([]string, 0)
			t.selected = Selected{}

			// Refresh tree widget to show shares
			if t.treeWidget != nil {
				t.treeWidget.Refresh()
			}

			t.w.Content().Refresh()
			t.SetStatus("Profile loaded successfully")
		}
	})

	// Create tree widget for navigation
	t.treeWidget = widget.NewTree(
		// ChildUIDs: Return child node IDs for a given parent
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return t.navTree.GetChildren(uid)
		},

		// IsBranch: Return true if node has/can have children
		func(uid widget.TreeNodeID) bool {
			return t.navTree.IsBranch(uid)
		},

		// CreateNode: Template for tree nodes
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.FolderIcon())
			label := widget.NewLabel("Template")
			label.Truncation = fyne.TextTruncateOff // Disable truncation to allow horizontal scrolling
			// Set a minimum width to enable horizontal scrolling for long names
			label.Resize(fyne.NewSize(500, label.MinSize().Height))
			content := container.NewHBox(icon, label)

			// Wrap in TappableTreeNode to support right-click
			tappable := newTappableTreeNode(content, "", t.handleTreeRightClick)
			return tappable
		},

		// UpdateNode: Apply data to template
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			// Update the tappable node's ID
			if tappable, ok := obj.(*TappableTreeNode); ok {
				tappable.UpdateNodeID(uid)
				tappable.SetTreeWidget(t.treeWidget)
				// Update the display content
				if tappable.content != nil {
					t.navTree.UpdateNodeDisplay(uid, tappable.content, branch)
				}
			}
		},
	)

	// Set the tree widget reference for all tappable nodes - this is a bit hacky but necessary
	// We need to do a second pass after creating the tree
	// This will be set in UpdateNode as nodes are rendered

	// Handle tree node selection
	t.treeWidget.OnSelected = func(uid widget.TreeNodeID) {
		t.handleTreeSelection(uid)
	}

	// No need for OnBranchOpened handler - all data is preloaded

	// Set up navigation panel with tree - use scroll container
	treeScroll := container.NewScroll(t.treeWidget)
	navCard := widget.NewCard("", "Navigation", treeScroll)
	// Make navigation panel wider to accommodate longer names (350px instead of 250px)
	t.left = container.NewGridWrap(fyne.NewSize(350, 768), navCard)

	tabs := container.NewDocTabs()
	tabs.CloseIntercept = func(ti *container.TabItem) {
		// Prevent closing the Browser tab - it should always be available
		if ti.Text == "Browser" {
			// Don't remove the Browser tab, just ignore the close request
			return
		}
		// Allow other tabs to be closed
		tabs.Remove(ti)
	}

	t.docTabs = tabs

	t.toolbar.Append(widget.NewToolbarAction(theme.MenuIcon(), func() {
		if !t.left.Visible() {
			t.left.Show()
		} else {
			t.left.Hide()
		}
	}))
	t.toolbar.Append(widget.NewToolbarSeparator())
	t.toolbar.Append(widget.NewToolbarAction(
		theme.FileIcon(), func() {
			t.OpenProfile()
		}))
	t.toolbar.Append(widget.NewToolbarSeparator())
	t.toolbar.Append(widget.NewToolbarAction(
		theme.ComputerIcon(), func() {
			t.showGoEditor()
		}))
	t.toolbar.Append(widget.NewToolbarSeparator())
	t.toolbar.Append(widget.NewToolbarAction(
		theme.ColorPaletteIcon(), func() {
			t.showThemeSelector()
		}))

	t.toolbar.Append(widget.NewToolbarSpacer())

	// Create Go Editor buttons container (separate from toolbar)
	executeBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if t.goEditor != nil {
			t.goEditor.executeCode()
		}
	})
	executeBtn.Importance = widget.HighImportance

	clearOutputBtn := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		if t.goEditor != nil {
			t.goEditor.clearOutput()
		}
	})

	clearEditorBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if t.goEditor != nil {
			t.goEditor.codeEditor.SetText("")
		}
	})

	saveBtn := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		if t.goEditor != nil {
			t.goEditor.saveCode()
		}
	})

	loadBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		if t.goEditor != nil {
			t.goEditor.loadCode()
		}
	})

	separator := widget.NewSeparator()

	t.goEditorButtonsContainer = container.NewHBox(
		separator,
		widget.NewLabel(" "), // Small spacer
		executeBtn,
		clearOutputBtn,
		clearEditorBtn,
		saveBtn,
		loadBtn,
	)
	t.goEditorButtonsContainer.Hide()

	// Create export button (initially hidden)
	t.exportButton = widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), func() {
		t.showExportMenu()
	})
	t.exportButton.Hide()

	// Create a container for the export button positioned on the right
	exportContainer := container.NewWithoutLayout(t.exportButton)
	t.exportButton.Resize(fyne.NewSize(100, 36))

	// Create a container that includes toolbar and Go Editor buttons
	toolbarRow := container.NewBorder(nil, nil, nil, nil,
		container.NewHBox(t.toolbar, t.goEditorButtonsContainer))

	t.top = container.NewStack(toolbarRow, exportContainer)

	// Set up tab change callback to show/hide export button and Go Editor buttons
	tabs.OnSelected = func(ti *container.TabItem) {
		t.updateExportButton()
		t.updateGoEditorButtons()
	}

	c := container.NewBorder(t.top, t.bottom, t.left, t.right, widget.NewCard("", "", tabs))
	t.w.SetContent(c)

	t.w.SetOnClosed(func() {
		// Cleanup if needed
	})

	t.OpenProfile()
	t.w.ShowAndRun()
}

// updateExportButton shows or hides the export button based on the current tab
func (t *MainWindow) updateExportButton() {
	if t.docTabs.Selected() != nil && t.docTabs.Selected().Text == "Browser" {
		t.exportButton.Show()
		// Position the button on the right side of the content area
		windowSize := t.w.Canvas().Size()
		// Position button: aligned to the right edge of the window
		buttonX := windowSize.Width - 130
		t.exportButton.Move(fyne.NewPos(buttonX, 4))
	} else {
		t.exportButton.Hide()
	}
	t.exportButton.Refresh()
}

// updateGoEditorButtons shows or hides Go Editor buttons based on the current tab
func (t *MainWindow) updateGoEditorButtons() {
	if t.docTabs.Selected() != nil && t.docTabs.Selected().Text == "Go Editor" {
		t.showGoEditorButtons()
	} else {
		t.hideGoEditorButtons()
	}
}

// showGoEditorButtons shows the Go Editor buttons container
func (t *MainWindow) showGoEditorButtons() {
	if t.goEditorButtonsContainer != nil {
		t.goEditorButtonsContainer.Show()
	}
}

// hideGoEditorButtons hides the Go Editor buttons container
func (t *MainWindow) hideGoEditorButtons() {
	if t.goEditorButtonsContainer != nil {
		t.goEditorButtonsContainer.Hide()
	}
}

// showThemeSelector displays a dialog for selecting the application theme
func (t *MainWindow) showThemeSelector() {
	currentTheme := t.themeManager.GetCurrentType()

	// Create radio group with theme options
	themeOptions := []string{
		GetThemeName(ThemeTypeCustom),
		GetThemeName(ThemeTypeShadcnSlate),
		GetThemeName(ThemeTypeShadcnStone),
		GetThemeName(ThemeTypeDefault),
	}

	selectedIndex := 0
	switch currentTheme {
	case ThemeTypeCustom:
		selectedIndex = 0
	case ThemeTypeShadcnSlate:
		selectedIndex = 1
	case ThemeTypeShadcnStone:
		selectedIndex = 2
	case ThemeTypeDefault:
		selectedIndex = 3
	}

	radio := widget.NewRadioGroup(themeOptions, nil)
	radio.SetSelected(themeOptions[selectedIndex])

	// Create info text
	infoLabel := widget.NewLabel("Choose a theme for the application.\nChanges will be applied immediately and saved.")
	infoLabel.Wrapping = fyne.TextWrapWord

	// Create the dialog content
	content := container.NewVBox(
		infoLabel,
		widget.NewSeparator(),
		radio,
	)

	// Create custom dialog
	d := dialog.NewCustom("Select Theme", "Close", content, t.w)
	d.Resize(fyne.NewSize(400, 300))

	// Handle theme selection changes
	radio.OnChanged = func(selected string) {
		var newTheme ThemeType
		switch selected {
		case GetThemeName(ThemeTypeCustom):
			newTheme = ThemeTypeCustom
		case GetThemeName(ThemeTypeShadcnSlate):
			newTheme = ThemeTypeShadcnSlate
		case GetThemeName(ThemeTypeShadcnStone):
			newTheme = ThemeTypeShadcnStone
		case GetThemeName(ThemeTypeDefault):
			newTheme = ThemeTypeDefault
		}

		if newTheme != t.themeManager.GetCurrentType() {
			t.themeManager.SetTheme(newTheme)
			t.SetStatus(fmt.Sprintf("Theme changed to: %s", selected))
		}
	}

	d.Show()
}

// showExportMenu displays the export menu for the currently selected browser tab
func (t *MainWindow) showExportMenu() {
	if t.dataBrowser == nil || t.dataBrowser.innerTabs == nil {
		return
	}

	selectedTab := t.dataBrowser.innerTabs.Selected()
	if selectedTab == nil {
		dialog.ShowInformation("No Table Selected", "Please select a table tab to export", t.w)
		return
	}

	// Get the data item for the selected tab
	dataItem, exists := t.dataBrowser.tabDataMap[selectedTab]
	if !exists {
		dialog.ShowError(fmt.Errorf("could not find data for selected tab"), t.w)
		return
	}

	// Get table name from the data item
	tableName := dataItem.tableName

	// Create export menu
	exportMenu := fyne.NewMenu("Export",
		fyne.NewMenuItem("Export as Parquet", func() {
			t.dataBrowser.exportData(dataItem, FormatParquet, tableName)
		}),
		fyne.NewMenuItem("Export as CSV", func() {
			t.dataBrowser.exportData(dataItem, FormatCSV, tableName)
		}),
		fyne.NewMenuItem("Export as JSON", func() {
			t.dataBrowser.exportData(dataItem, FormatJSON, tableName)
		}),
	)

	// Show the menu at the export button position
	widget.ShowPopUpMenuAtPosition(exportMenu, t.w.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(t.exportButton))
}

// showGoEditor shows or creates the Go editor tab
func (t *MainWindow) showGoEditor() {
	// Check if Go tab already exists
	for _, tab := range t.docTabs.Items {
		if tab.Text == "Go Editor" {
			// Tab exists, just select it
			t.docTabs.Select(tab)
			// Explicitly update Go Editor buttons to ensure they are shown
			t.updateGoEditorButtons()
			t.SetStatus("Go editor opened")
			return
		}
	}

	// Create new Go editor if it doesn't exist
	if t.goEditor == nil {
		t.goEditor = NewGoEditor(t.w)
	}

	// Create and add the Go tab
	goTab := container.NewTabItem("Go Editor", t.goEditor.GetContainer())
	t.docTabs.Append(goTab)
	t.docTabs.Select(goTab)

	// Explicitly update Go Editor buttons to ensure they are shown
	t.updateGoEditorButtons()

	// Hide navigation menu when Go editor is opened
	if t.left.Visible() {
		t.left.Hide()
	}

	t.SetStatus("Go editor opened")
}

// handleTreeSelection handles selection of a node in the navigation tree
func (t *MainWindow) handleTreeSelection(nodeID widget.TreeNodeID) {
	node := t.navTree.GetNode(nodeID)
	if node == nil {
		return
	}

	switch node.NodeType {
	case NodeTypeShare:
		t.selected.share = node.Name
		t.selected.schema = ""
		t.selected.table_name = ""
		t.SetStatus("Share selected: " + node.Name)

	case NodeTypeSchema:
		t.selected.share = node.Share
		t.selected.schema = node.Name
		t.selected.table_name = ""
		t.SetStatus("Schema selected: " + node.Name)

	case NodeTypeTable:
		t.selected.share = node.Share
		t.selected.schema = node.Schema
		t.selected.table_name = node.Name
		t.selected.table = node.Table
		t.SetStatus("Loading table data (first 1000 rows): " + node.Name)

		// Load table data with default 1000 row limit
		t.loadTableData(node.Table, &QueryOptions{Limit: 1000})
	}
}

// handleTreeRightClick handles right-click on tree nodes
func (t *MainWindow) handleTreeRightClick(nodeID widget.TreeNodeID, e *fyne.PointEvent) {
	node := t.navTree.GetNode(nodeID)
	if node == nil {
		return
	}

	// Create context menu based on node type
	var menuItems []*fyne.MenuItem

	switch node.NodeType {
	case NodeTypeTable:
		// Menu items for table nodes
		menuItems = []*fyne.MenuItem{
			fyne.NewMenuItem("Open Table", func() {
				// Select and load the table with default options (no filtering)
				t.treeWidget.Select(nodeID)
			}),
			fyne.NewMenuItem("Open with Query Options...", func() {
				// Show enhanced query options dialog with column checkboxes
				ShowQueryOptionsDialogWithSchema(t.w, t.profile, node.Table, func(options *QueryOptions) {
					// Update selected state
					t.selected.share = node.Share
					t.selected.schema = node.Schema
					t.selected.table_name = node.Name
					t.selected.table = node.Table
					t.SetStatus("Loading table data with options: " + node.Name)

					// Load table data with options
					t.loadTableData(node.Table, options)
				})
			}),
			fyne.NewMenuItem("Copy Table Name", func() {
				t.w.Clipboard().SetContent(node.Name)
				t.SetStatus("Table name copied to clipboard")
			}),
		}

	case NodeTypeSchema:
		// Menu items for schema nodes
		menuItems = []*fyne.MenuItem{
			fyne.NewMenuItem("Expand Schema", func() {
				t.treeWidget.OpenBranch(nodeID)
			}),
			fyne.NewMenuItem("Copy Schema Name", func() {
				t.w.Clipboard().SetContent(node.Name)
				t.SetStatus("Schema name copied to clipboard")
			}),
		}

	case NodeTypeShare:
		// Menu items for share nodes
		menuItems = []*fyne.MenuItem{
			fyne.NewMenuItem("Expand Share", func() {
				t.treeWidget.OpenBranch(nodeID)
			}),
			fyne.NewMenuItem("Copy Share Name", func() {
				t.w.Clipboard().SetContent(node.Name)
				t.SetStatus("Share name copied to clipboard")
			}),
		}
	}

	if len(menuItems) > 0 {
		menu := fyne.NewMenu("", menuItems...)
		popUpMenu := widget.NewPopUpMenu(menu, t.w.Canvas())
		popUpMenu.ShowAtPosition(e.AbsolutePosition)
	}
}

// loadTableData loads and displays data for a table
func (t *MainWindow) loadTableData(table delta_sharing.Table, options *QueryOptions) {
	ds, err := delta_sharing.NewSharingClientV2FromString(t.profile)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to create client: %w", err), t.w)
		return
	}

	re, err := ds.ListFilesInTable(context.Background(), table)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to list files: %w", err), t.w)
		return
	}

	t.files = make([]string, 0)
	for _, v := range re.AddFiles {
		t.files = append(t.files, v.Id)
	}

	if len(t.files) == 0 {
		dialog.ShowError(fmt.Errorf("no files available for table"), t.w)
		return
	}

	fileSelected := t.files[0]

	// Initialize data browser if needed
	if t.dataBrowser == nil {
		var db DataBrowser
		db.CreateWindow(t.docTabs, t.SetStatus)
		t.dataBrowser = &db
	}

	// Load data - GetData handles its own threading
	t.dataBrowser.GetData(t.profile, table, fileSelected, options)
}

// ScanTree is deprecated - tree navigation now uses lazy loading
// This method is kept for compatibility but does nothing
func (t *MainWindow) ScanTree() {
	// No longer needed with tree-based navigation
	// Data is loaded on-demand when tree nodes are expanded
}
