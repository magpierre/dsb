package windows

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	tab         container.TabItem
}

type DataBrowser struct {
	w       fyne.Window
	content fyne.Container
	Data    []Data
	tabs    []*container.TabItem
	docTabs *container.DocTabs
}

func (t *DataBrowser) CreateWindow(docTabs *container.DocTabs) {
	t.w = fyne.CurrentApp().Driver().AllWindows()[0]
	t.docTabs = docTabs
	t.Data = make([]Data, 0)
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

	content := widget.NewCard("", "", table)
	t.tabs = append(t.tabs, container.NewTabItem(delta_table.Name, content))

	tabs := container.NewDocTabs(t.tabs...)
	tabs.CloseIntercept = func(ti *container.TabItem) {
	}
	tabs.SetTabLocation(container.TabLocationBottom)

	for _, v := range t.docTabs.Items {
		if v.Text == "Browser" {
			t.docTabs.Remove(v)
		}
	}

	browserAccordionItem := widget.NewAccordionItem("Browser", tabs)
	browserAccordionItem.Open = true
	accordion := widget.NewAccordion(browserAccordionItem)
	t.docTabs.Append(container.NewTabItem("Browser", accordion))

	tabs.SelectIndex(len(t.tabs) - 1)
	tabs.Refresh()
	t.docTabs.SelectIndex(2)
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
	t.Data[dp].arrow_rec.Release()

	t.Data[dp].arrow_table.Release()
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
