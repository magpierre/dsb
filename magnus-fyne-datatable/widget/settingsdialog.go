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

package widget

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SettingsDialog provides a dialog for configuring DataTable options.
type SettingsDialog struct {
	dataTable *DataTable
	window    fyne.Window

	// UI components
	filterBarCheck      *widget.Check
	statusBarCheck      *widget.Check
	columnSelectorCheck *widget.Check
	autoAdjustCheck     *widget.Check
	selectionModeSelect *widget.RadioGroup
	minWidthEntry       *widget.Entry

	dialog dialog.Dialog
}

// NewSettingsDialog creates a new settings dialog for the given DataTable.
func NewSettingsDialog(dt *DataTable, window fyne.Window) *SettingsDialog {
	sd := &SettingsDialog{
		dataTable: dt,
		window:    window,
	}

	sd.buildDialog()
	return sd
}

// buildDialog constructs the settings dialog UI.
func (sd *SettingsDialog) buildDialog() {
	// Create checkboxes for boolean options
	sd.filterBarCheck = widget.NewCheck("Show Filter Bar", nil)
	sd.filterBarCheck.Checked = sd.dataTable.config.ShowFilterBar

	sd.statusBarCheck = widget.NewCheck("Show Status Bar", nil)
	sd.statusBarCheck.Checked = sd.dataTable.config.ShowStatusBar

	sd.columnSelectorCheck = widget.NewCheck("Show Column Selector", nil)
	sd.columnSelectorCheck.Checked = sd.dataTable.config.ShowColumnSelector

	sd.autoAdjustCheck = widget.NewCheck("Auto-Adjust Column Widths", nil)
	sd.autoAdjustCheck.Checked = sd.dataTable.config.AutoAdjustColumnWidths

	// Selection mode radio group
	sd.selectionModeSelect = widget.NewRadioGroup([]string{
		"Cell Selection (individual cells)",
		"Row Selection (entire rows)",
	}, nil)

	if sd.dataTable.config.SelectionMode == SelectionModeRow {
		sd.selectionModeSelect.SetSelected("Row Selection (entire rows)")
	} else {
		sd.selectionModeSelect.SetSelected("Cell Selection (individual cells)")
	}

	// Minimum column width entry
	sd.minWidthEntry = widget.NewEntry()
	sd.minWidthEntry.SetPlaceHolder("100")
	sd.minWidthEntry.SetText(formatInt(sd.dataTable.config.MinColumnWidth))

	// Create form layout
	form := container.NewVBox(
		widget.NewLabel("Display Options:"),
		sd.filterBarCheck,
		sd.statusBarCheck,
		sd.columnSelectorCheck,
		widget.NewSeparator(),
		widget.NewLabel("Column Options:"),
		sd.autoAdjustCheck,
		container.NewBorder(nil, nil, widget.NewLabel("Min Column Width:"), nil, sd.minWidthEntry),
		widget.NewSeparator(),
		widget.NewLabel("Selection Mode:"),
		sd.selectionModeSelect,
	)

	// Create dialog with Apply and Cancel buttons
	sd.dialog = dialog.NewCustomConfirm(
		"DataTable Settings",
		"Apply",
		"Cancel",
		form,
		func(apply bool) {
			if apply {
				sd.applySettings()
			}
		},
		sd.window,
	)

	sd.dialog.Resize(fyne.NewSize(400, 500))
}

// Show displays the settings dialog.
func (sd *SettingsDialog) Show() {
	sd.dialog.Show()
}

// applySettings applies the selected settings to the DataTable.
func (sd *SettingsDialog) applySettings() {
	// Update config
	newConfig := sd.dataTable.config

	newConfig.ShowFilterBar = sd.filterBarCheck.Checked
	newConfig.ShowStatusBar = sd.statusBarCheck.Checked
	newConfig.ShowColumnSelector = sd.columnSelectorCheck.Checked
	newConfig.AutoAdjustColumnWidths = sd.autoAdjustCheck.Checked

	// Update selection mode
	if sd.selectionModeSelect.Selected == "Row Selection (entire rows)" {
		newConfig.SelectionMode = SelectionModeRow
	} else {
		newConfig.SelectionMode = SelectionModeCell
	}

	// Update minimum column width
	if width := parseInt(sd.minWidthEntry.Text); width > 0 {
		newConfig.MinColumnWidth = width
	}

	// Apply new configuration by reconfiguring the DataTable
	sd.dataTable.Reconfigure(newConfig)
}

// Helper functions for string/int conversion
func formatInt(value int) string {
	return strconv.Itoa(value)
}

func parseInt(text string) int {
	if text == "" {
		return 0
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0
	}
	return value
}
