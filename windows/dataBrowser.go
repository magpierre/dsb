package windows

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

type Data struct {
	data        [][]string
	header      []string
	arrow_table arrow.Table
	arrow_rec   arrow.Record
	tab         *container.TabItem
	tableName   string
}

type DataBrowser struct {
	w            fyne.Window
	Data         []Data
	innerTabs    *container.DocTabs
	docTabs      *container.DocTabs
	browserTab   *container.TabItem
	tabDataMap   map[*container.TabItem]*Data
}

func (t *DataBrowser) CreateWindow(docTabs *container.DocTabs) {
	t.w = fyne.CurrentApp().Driver().AllWindows()[0]
	t.docTabs = docTabs
	t.Data = make([]Data, 0)
	t.tabDataMap = make(map[*container.TabItem]*Data)

	// Create persistent inner tabs for individual tables
	t.innerTabs = container.NewDocTabs()
	t.innerTabs.SetTabLocation(container.TabLocationBottom)

	// Set up close intercept to clean up memory when tabs are closed
	t.innerTabs.CloseIntercept = func(ti *container.TabItem) {
		// Find and clean up the data associated with this tab
		if data, exists := t.tabDataMap[ti]; exists {
			// Release Arrow resources if they haven't been released yet
			if data.arrow_rec != nil {
				data.arrow_rec.Release()
			}
			if data.arrow_table != nil {
				data.arrow_table.Release()
			}
			// Clear the data arrays to help GC
			data.data = nil
			data.header = nil
			// Remove from map
			delete(t.tabDataMap, ti)
		}
		// Remove the tab
		t.innerTabs.Remove(ti)
	}

	// Create persistent Browser tab
	t.browserTab = container.NewTabItem("Browser", t.innerTabs)
	t.docTabs.Append(t.browserTab)
}

func (t *DataBrowser) CreateDataBrowser(dataItem *Data, delta_table delta_sharing.Table) {
	table := widget.NewTableWithHeaders(func() (rows int, cols int) {
		return len(dataItem.data), len(dataItem.data[0])
	}, func() fyne.CanvasObject {
		return widget.NewLabel("template.............")
	}, func(tci widget.TableCellID, co fyne.CanvasObject) {
		co.(*widget.Label).SetText(dataItem.data[tci.Row][tci.Col])
		co.(*widget.Label).Truncation = fyne.TextTruncateClip
	})

	table.ShowHeaderColumn = false
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		template.(*widget.Label).SetText(dataItem.header[id.Col])
		template.(*widget.Label).Truncation = fyne.TextTruncateClip
	}

	// Calculate and set column widths based on header text
	for i, headerText := range dataItem.header {
		// Measure the header text width and add padding
		textSize := fyne.MeasureText(headerText, theme.TextSize(), fyne.TextStyle{})
		columnWidth := textSize.Width + theme.Padding()*4 // Add padding for better spacing

		// Ensure a minimum width
		if columnWidth < 80 {
			columnWidth = 80
		}

		table.SetColumnWidth(i, columnWidth)
	}

	// Create table card with column and row count
	rowCount := len(dataItem.data)
	colCount := len(dataItem.header)
	cardTitle := fmt.Sprintf("Table %s (%d columns x %d rows)", delta_table.Name, colCount, rowCount)
	content := widget.NewCard("", cardTitle, table)

	// Add new tab to the persistent inner tabs
	newTab := container.NewTabItem(delta_table.Name, content)

	// Store the tab reference in the data item and register in map
	dataItem.tab = newTab
	dataItem.tableName = delta_table.Name
	t.tabDataMap[newTab] = dataItem

	t.innerTabs.Append(newTab)

	// Select the newly added tab
	t.innerTabs.Select(newTab)

	// Check if Browser tab still exists in docTabs, if not recreate it
	browserExists := false
	for _, item := range t.docTabs.Items {
		if item == t.browserTab {
			browserExists = true
			break
		}
	}

	if !browserExists {
		// Recreate the Browser tab
		t.browserTab = container.NewTabItem("Browser", t.innerTabs)
		t.docTabs.Append(t.browserTab)
	}

	// Select the Browser tab in the main tabs
	t.docTabs.Select(t.browserTab)
}

func (t *DataBrowser) GetData(profile string, table delta_sharing.Table, file_id string) {
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
	ds, err := delta_sharing.NewSharingClientFromString(context.Background(), profile, "")
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	resp, err := ds.ListFilesInTable(table)
	if err != nil {
		dialog.NewError(err, t.w).Show()
	}
	var data Data
	for _, v := range resp.AddFiles {
		if v.Id == file_id {
			arrow_table, err := delta_sharing.LoadArrowTable(ds, table, file_id)
			if err != nil {
				dialog.NewError(err, t.w).Show()
			}
			data.arrow_table = arrow_table
			data.arrow_table, err = t.test(data.arrow_table)
			if err != nil {
				fmt.Println(err)
				c <- true
				return
			}
			var header []string = make([]string, data.arrow_table.NumCols())
			for i, f := range data.arrow_table.Schema().Fields() {
				header[i] = f.Name
			}

			data.data = make([][]string, 0)
			data.header = header

			tr := array.NewTableReader(data.arrow_table, 1000)
			tr.Retain()
			tr.Next()
			data.arrow_rec = tr.Record()
			t.Data = append(t.Data, data)
			dt := t.parseRecord()

			t.CreateDataBrowser(dt, table)

			c <- true
			t.w.Content().Refresh()
		}
	}
}

func (t *DataBrowser) parseRecord() *Data {
	dp := len(t.Data) - 1
	for pos := 0; pos < int(t.Data[dp].arrow_rec.NumRows()); pos++ {
		var v []string = make([]string, t.Data[dp].arrow_rec.NumCols())
		for i, col := range t.Data[dp].arrow_rec.Columns() {
			switch col.DataType().ID() {
			case arrow.STRUCT:
				s := col.(*array.Struct)

				b, err := s.MarshalJSON()
				if err != nil {
					log.Fatal(err)
				}
				v[i] = string(b)

			case arrow.LIST:
				as := array.NewSlice(col, int64(pos), int64(pos+1))
				str := fmt.Sprintf("%v", as)
				if len(str) > 253 {
					v[i] = str[1:253] + "..."
				} else {
					v[i] = str
				}
			case arrow.STRING:
				s := col.(*array.String)
				v[i] = s.Value(pos)
			case arrow.BINARY:
				b := col.(*array.Binary)
				v[i] = string(b.Value(pos))
			case arrow.BOOL:
				b := col.(*array.Boolean)
				v[i] = fmt.Sprintf("%v", b.Value(pos))
			case arrow.DATE32:
				d32 := col.(*array.Date32)
				v[i] = d32.Value(pos).ToTime().String()
			case arrow.DATE64:
				d64 := col.(*array.Date64)
				v[i] = d64.Value(pos).ToTime().String()
			case arrow.DECIMAL:
				d128 := col.(*array.Decimal128)
				v[i] = d128.Value(pos).BigInt().String()
			case arrow.INT8:
				i8 := col.(*array.Int8)
				v[i] = fmt.Sprintf("%d", i8.Value(pos))
			case arrow.INT16:
				i16 := col.(*array.Int16)
				v[i] = fmt.Sprintf("%d", i16.Value(pos))
			case arrow.INT32:
				i32 := col.(*array.Int32)
				v[i] = fmt.Sprintf("%d", i32.Value(pos))
			case arrow.INT64:
				i64 := col.(*array.Int64)
				v[i] = fmt.Sprintf("%d", i64.Value(pos))
			case arrow.FLOAT16:
				f16 := col.(*array.Float16)
				v[i] = f16.Value(pos).String()
			case arrow.FLOAT32:
				f32 := col.(*array.Float32)
				v[i] = fmt.Sprintf("%.2f", f32.Value(pos))
			case arrow.FLOAT64:
				f64 := col.(*array.Float64)
				v[i] = fmt.Sprintf("%.2f", f64.Value(pos))
			case arrow.INTERVAL_MONTHS:
				intV := col.(*array.DayTimeInterval)
				v[i] = fmt.Sprintf("%v", intV.Value(pos))
			case arrow.INTERVAL_DAY_TIME:
				intV := col.(*array.DayTimeInterval)
				v[i] = fmt.Sprintf("%v", intV.Value(pos))
			case arrow.TIMESTAMP:
				ts := col.(*array.Timestamp)
				v[i] = ts.Value(pos).ToTime(arrow.Nanosecond).String()
			}
		}
		t.Data[dp].data = append(t.Data[dp].data, v)
	}
	// Don't release Arrow resources here - they will be released when the tab is closed
	// t.Data[dp].arrow_rec.Release()
	// t.Data[dp].arrow_table.Release()
	return &t.Data[dp]
}

func (d *DataBrowser) test(t arrow.Table) (arrow.Table, error) {
	/* table := t

	pool := compute.GetAllocator(context.Background())
	intBuilder := array.NewFloat32Builder(pool)
	intBuilder.Append(21)
	a := intBuilder.NewArray()
	//c := compute.Equal(compute.NewFieldRef("deaths"), compute.NewLiteral(21))
	tab2, err := compute.FilterTable(context.Background(), table, compute.NewDatum(a), compute.DefaultFilterOptions())
	if err != nil {
		return nil, err
	}
	*/
	return t, nil
}
