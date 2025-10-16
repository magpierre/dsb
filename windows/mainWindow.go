package windows

import (
	"context"
	"fmt"
	"time"

	"dsb/windows/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
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
	share                    []string
	schemas                  []string
	tables                   []string
	files                    []string
	selected                 Selected
	docTabs                  *container.DocTabs
	dataBrowser              *DataBrowser
	shareBindingList         binding.StringList
	schemaBindingList        binding.StringList
	tablesBindingList        binding.StringList
	statusBar                *widget.Label
}

func CreateMainWindow() *MainWindow {
	var v MainWindow
	v.NewMainWindow()
	return &v
}

func (t *MainWindow) OpenProfile() {
	pd := NewProfileDialog(t.w, t.a, func(content string, err error) {
		if err != nil {
			t.SetStatus("Error opening profile")
			dialog.ShowError(err, t.w)
			return
		}

		if content == "" {
			return
		}

		t.SetStatus("Loading profile...")
		t.profile = content

		ds, err := delta_sharing.NewSharingClientFromString(context.Background(), t.profile, "")
		if err != nil {
			t.SetStatus("Error connecting to Delta Sharing")
			dialog.ShowError(err, t.w)
			return
		}

		share, err := ds.ListShares()
		if err != nil {
			t.SetStatus("Error listing shares")
			dialog.ShowError(err, t.w)
			return
		}
		t.share = make([]string, 0)
		t.schemas = make([]string, 0)
		t.tables = make([]string, 0)
		t.files = make([]string, 0)
		t.selected = Selected{}
		t.w.Content().Refresh()
		for _, s := range share {
			t.share = append(t.share, s.Name)
		}

		t.shareBindingList.Set(t.share)
		t.schemaBindingList.Set(t.schemas)
		t.tablesBindingList.Set(t.tables)
		t.SetStatus("Profile loaded successfully")
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
	t.a.Settings().SetTheme(&CustomTheme{})
	t.top = widget.NewToolbar()
	t.left = container.NewVBox()
	t.right = container.NewVBox()

	// Create status bar
	t.statusBar = widget.NewLabel("Ready")
	t.statusBar.TextStyle = fyne.TextStyle{Italic: true}
	t.bottom = container.NewHBox(t.statusBar)

	t.shareBindingList = binding.NewStringList()
	t.schemaBindingList = binding.NewStringList()
	t.tablesBindingList = binding.NewStringList()
	t.w = t.a.NewWindow("Delta Sharing Browser")
	t.w.Resize(fyne.NewSize(700, 600))

	logo := canvas.NewImageFromResource(resources.ResourceDeltasharingPng)
	logo.FillMode = canvas.ImageFillContain

	shareWidget := widget.NewListWithData(t.shareBindingList, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(di binding.DataItem, co fyne.CanvasObject) {
		co.(*widget.Label).Bind(di.(binding.String))
	})

	schemaWidget := widget.NewListWithData(t.schemaBindingList, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(di binding.DataItem, co fyne.CanvasObject) {
		co.(*widget.Label).Bind(di.(binding.String))
	})

	// Store reference to the context menu callback
	var showTableContextMenu func(widget.ListItemID, *fyne.PointEvent)

	tablesWidget := widget.NewListWithData(t.tablesBindingList, func() fyne.CanvasObject {
		return newTappableListItem(showTableContextMenu)
	}, func(di binding.DataItem, co fyne.CanvasObject) {
		item := co.(*TappableListItem)
		item.Bind(di.(binding.String))
	})

	// Set item IDs and tap handler when updating
	originalUpdateItem := tablesWidget.UpdateItem
	tablesWidget.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		if tappableItem, ok := item.(*TappableListItem); ok {
			tappableItem.SetItemID(id)
			// Connect regular tap to the list's OnSelected handler
			tappableItem.SetOnTap(func(itemID widget.ListItemID) {
				if tablesWidget.OnSelected != nil {
					tablesWidget.OnSelected(itemID)
				}
			})
		}
		originalUpdateItem(id, item)
	}

	gr := container.NewVSplit(widget.NewCard("", "Shares", shareWidget), widget.NewCard("", "Schemas", schemaWidget))
	t.left = container.NewGridWrap(fyne.NewSize(150, 768), gr)

	shareWidget.OnSelected = func(id widget.ListItemID) {
		x := t.share[id]
		t.selected.share = x
		t.SetStatus("Loading schemas for share: " + x)
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tables = make([]string, 0)
		t.files = make([]string, 0)
		t.tablesBindingList.Set(t.tables)
		schemaWidget.UnselectAll()
		tablesWidget.UnselectAll()
		t.SetStatus("Share selected: " + x)
	}
	schemaWidget.OnSelected = func(id widget.ListItemID) {
		x := t.schemas[id]
		t.selected.schema = x
		t.SetStatus("Loading tables for schema: " + x)
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tablesBindingList.Set(t.tables)
		t.files = make([]string, 0)
		tablesWidget.UnselectAll()
		t.SetStatus("Schema selected: " + x)
	}

	tablesWidget.OnSelected = func(id widget.ListItemID) {
		x := t.tables[id]
		t.selected.table_name = x
		t.SetStatus("Loading table data: " + x)
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tablesBindingList.Set(t.tables)
		fileSelected := t.files[0]
		if t.dataBrowser == nil {
			var db DataBrowser
			db.CreateWindow(t.docTabs)
			t.dataBrowser = &db
		}
		t.dataBrowser.GetData(t.profile, t.selected.table, fileSelected, nil)
		t.SetStatus("Table loaded: " + x)
	}

	// Create context menu for tables list
	// Define the context menu callback function
	showTableContextMenu = func(itemID widget.ListItemID, e *fyne.PointEvent) {
		// Get the table name
		tableName := "unknown"
		if itemID >= 0 && itemID < widget.ListItemID(len(t.tables)) {
			tableName = t.tables[itemID]
		}

		t.SetStatus(fmt.Sprintf("Right-click on table: %s", tableName))

		// Create the context menu
		tableContextMenu := fyne.NewMenu("",
			fyne.NewMenuItem("Load with Options...", func() {
				if itemID < 0 || itemID >= widget.ListItemID(len(t.tables)) {
					dialog.ShowInformation("Select a Table", "Please select a table first", t.w)
					return
				}

				x := t.tables[itemID]
				t.selected.table_name = x
				t.ScanTree()

				if len(t.files) == 0 {
					dialog.ShowError(fmt.Errorf("no files available for table"), t.w)
					return
				}

				fileSelected := t.files[0]

				// Load the table schema first
				t.SetStatus("Loading schema for table: " + x)
				ds, err := delta_sharing.NewSharingClientFromString(context.Background(), t.profile, "")
				if err != nil {
					dialog.ShowError(err, t.w)
					return
				}

				// Load Arrow table to get schema
				arrow_table, err := delta_sharing.LoadArrowTable(ds, t.selected.table, fileSelected)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to load table schema: %w", err), t.w)
					return
				}
				defer arrow_table.Release()

				schema := arrow_table.Schema()

				// Show query options dialog with schema
				queryDialog := NewQueryOptionsDialog(t.w, schema, func(options *QueryOptions) {
					t.SetStatus("Loading table data with options: " + x)
					if t.dataBrowser == nil {
						var db DataBrowser
						db.CreateWindow(t.docTabs)
						t.dataBrowser = &db
					}
					t.dataBrowser.GetData(t.profile, t.selected.table, fileSelected, options)
					t.SetStatus("Table loaded with options: " + x)
				})
				queryDialog.Show()
			}),
			fyne.NewMenuItem("Load All Data", func() {
				if itemID >= 0 {
					tablesWidget.OnSelected(itemID)
				}
			}),
		)

		// Show the context menu at the click position
		widget.ShowPopUpMenuAtPosition(tableContextMenu, t.w.Canvas(), e.AbsolutePosition)
	}

	tabs := container.NewDocTabs(container.NewTabItem("Tables", widget.NewCard("", "Tables", tablesWidget)))
	tabs.CloseIntercept = func(ti *container.TabItem) {
		if ti.Text == "Browser" {
			tabs.Remove(ti)
		}
	}

	t.docTabs = tabs

	t.top.(*widget.Toolbar).Append(widget.NewToolbarAction(theme.MenuIcon(), func() {
		if !t.left.Visible() {
			t.left.Show()
		} else {
			t.left.Hide()
		}
	}))
	t.top.(*widget.Toolbar).Append(widget.NewToolbarSeparator())
	t.top.(*widget.Toolbar).Append(widget.NewToolbarAction(
		theme.FileIcon(), func() {
			t.OpenProfile()
		}))

	t.top.(*widget.Toolbar).Append(widget.NewToolbarSpacer())

	llo := container.NewWithoutLayout(logo)
	logo.Resize(fyne.NewSize(200, 50))
	logo.Move(fyne.NewPos(160, -10))
	t.top = container.NewStack(t.top, llo)

	c := container.NewBorder(t.top, t.bottom, t.left, t.right, widget.NewCard("", "", tabs))
	t.w.SetContent(c)
	t.OpenProfile()
	t.w.ShowAndRun()
}

func (t *MainWindow) ScanTree() {
	c := make(chan bool)
	go func(c chan bool) {
		pbi := widget.NewProgressBarInfinite()
		di := dialog.NewCustomWithoutButtons("Please wait", pbi, t.w)
		di.Resize(fyne.NewSize(200, 100))
		di.Show()
		pbi.Start()
		for {
			select {
			case <-c:
				di.Hide()
				pbi.Stop()
				return
			default:
				time.Sleep(time.Millisecond + 500)
			}
		}
	}(c)
	ds, err := delta_sharing.NewSharingClientFromString(context.Background(), t.profile, "")
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	ls, err := ds.ListShares()
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	for _, v := range ls {
		if v.Name == t.selected.share {
			sh, err := ds.ListSchemas(v)
			if err != nil {
				dialog.NewError(err, t.w).Show()
			}
			t.schemas = make([]string, 0)
			t.tables = make([]string, 0)
			t.files = make([]string, 0)
			for _, v2 := range sh {
				t.schemas = append(t.schemas, v2.Name)
				if v2.Name == t.selected.schema && v2.Share == t.selected.share {
					tl, err := ds.ListTables(v2)
					if err != nil {
						dialog.NewError(err, t.w).Show()
					}
					for _, tle := range tl {
						t.tables = append(t.tables, tle.Name)
						if tle.Schema == t.selected.schema && tle.Share == t.selected.share && tle.Name == t.selected.table_name {
							t.selected.table = tle
							re, err := ds.ListFilesInTable(tle)
							if err != nil {
								dialog.NewError(err, t.w).Show()
							}
							t.files = make([]string, 0)
							for _, v := range re.AddFiles {
								t.files = append(t.files, v.Id)
							}
						}
					}
				}
			}
		}
	}
	c <- true
}
