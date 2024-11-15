package windows

import (
	"context"
	"io"
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
	th "fyne.io/x/fyne/theme"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

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
}

func CreateMainWindow() *MainWindow {
	var v MainWindow
	v.NewMainWindow()
	return &v
}

func (t *MainWindow) OpenProfile() *dialog.FileDialog {
	d := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil || uc == nil {
			return
		}

		d, err := io.ReadAll(uc)
		if err != nil {
			dialog.NewError(err, t.w)
			return
		}
		t.profile = string(d)

		ds, err := delta_sharing.NewSharingClientFromString(context.Background(), t.profile, "")
		if err != nil {
			dialog.NewError(err, t.w).Show()
		}

		ds.ListShares()

		share, _ := ds.ListShares()
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
	}, t.w)
	return d
}

func (t *MainWindow) NewMainWindow() {
	t.selected = Selected{}
	t.a = app.NewWithID("dsb")
	t.a.Settings().SetTheme(th.AdwaitaTheme())
	t.top = widget.NewToolbar()
	t.left = container.NewVBox()
	t.right = container.NewVBox()
	t.bottom = container.NewHBox()
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

	tablesWidget := widget.NewListWithData(t.tablesBindingList, func() fyne.CanvasObject {
		return widget.NewLabel("template")
	}, func(di binding.DataItem, co fyne.CanvasObject) {
		co.(*widget.Label).Bind(di.(binding.String))
	})

	gr := container.NewVSplit(widget.NewCard("", "Shares", shareWidget), widget.NewCard("", "Schemas", schemaWidget))
	t.left = container.NewGridWrap(fyne.NewSize(150, 768), gr)
	tabs := container.NewDocTabs(container.NewTabItem("Tables", widget.NewCard("", "Tables", tablesWidget)))
	tabs.CloseIntercept = func(ti *container.TabItem) {
		if ti.Text == "Browser" {
			tabs.Remove(ti)
		}
	}

	t.docTabs = tabs
	shareWidget.OnSelected = func(id widget.ListItemID) {
		x := t.share[id]
		t.selected.share = x
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tables = make([]string, 0)
		t.files = make([]string, 0)
		t.tablesBindingList.Set(t.tables)
		schemaWidget.UnselectAll()
		tablesWidget.UnselectAll()
		tabs.Refresh()
	}
	schemaWidget.OnSelected = func(id widget.ListItemID) {
		x := t.schemas[id]
		t.selected.schema = x
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tablesBindingList.Set(t.tables)
		t.files = make([]string, 0)
		tablesWidget.UnselectAll()
		tabs.Refresh()
	}

	tablesWidget.OnSelected = func(id widget.ListItemID) {
		x := t.tables[id]
		t.selected.table_name = x
		t.ScanTree()
		t.schemaBindingList.Set(t.schemas)
		t.tablesBindingList.Set(t.tables)
		fileSelected := t.files[0]
		if t.dataBrowser == nil {
			var db DataBrowser
			db.CreateWindow(t.docTabs)
			t.dataBrowser = &db
		}
		t.dataBrowser.GetData(t.profile, t.selected.table, fileSelected)
		/*da := NewDataAggregator()
		ti := da.CreateTab(t.dataBrowser.parseRecord().header)
		t.docTabs.Append(ti)
		tabs.Refresh()
		*/
		t.docTabs.SelectIndex(1)
	}

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
			d := t.OpenProfile()
			d.Show()
		}))

	t.top.(*widget.Toolbar).Append(widget.NewToolbarSpacer())

	llo := container.NewWithoutLayout(logo)
	logo.Resize(fyne.NewSize(200, 50))
	logo.Move(fyne.NewPos(160, -10))
	t.top = container.NewStack(t.top, llo)

	c := container.NewBorder(t.top, t.bottom, t.left, t.right, widget.NewCard("", "", tabs))
	t.w.SetContent(c)
	t.OpenProfile().Show()
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
