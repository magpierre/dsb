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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ProfileDialog struct {
	dialog         dialog.Dialog
	window         fyne.Window
	callback       func(string, error)
	fileList       *widget.List
	recentList     *widget.List
	files          []string
	recentProfiles []string
	homeDir        string
	currentPath    string
	pathLabel      *widget.Label
	app            fyne.App
	filePath       string // Store the selected file path
}

const maxRecentProfiles = 5
const recentProfilesKey = "recent_profiles"

func NewProfileDialog(w fyne.Window, a fyne.App, callback func(string, error)) *ProfileDialog {
	pd := &ProfileDialog{
		window:   w,
		app:      a,
		callback: callback,
		files:    make([]string, 0),
		filePath: "",
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	pd.homeDir = homeDir
	pd.currentPath = homeDir

	// Load recent profiles
	pd.loadRecentProfiles()

	return pd
}

// loadRecentProfiles loads the list of recently selected profiles from preferences
func (pd *ProfileDialog) loadRecentProfiles() {
	recentJSON := pd.app.Preferences().StringWithFallback(recentProfilesKey, "[]")
	pd.recentProfiles = make([]string, 0)
	err := json.Unmarshal([]byte(recentJSON), &pd.recentProfiles)
	if err != nil {
		// Silently ignore error and start with empty recent profiles list
		pd.recentProfiles = make([]string, 0)
	}
}

// saveRecentProfiles saves the list of recently selected profiles to preferences
func (pd *ProfileDialog) saveRecentProfiles() {
	recentJSON, _ := json.Marshal(pd.recentProfiles)
	pd.app.Preferences().SetString(recentProfilesKey, string(recentJSON))
}

// addRecentProfile adds a profile path to the recent profiles list
func (pd *ProfileDialog) addRecentProfile(profilePath string) {
	// Remove if already exists
	for i, path := range pd.recentProfiles {
		if path == profilePath {
			pd.recentProfiles = append(pd.recentProfiles[:i], pd.recentProfiles[i+1:]...)
			break
		}
	}

	// Add to front
	pd.recentProfiles = append([]string{profilePath}, pd.recentProfiles...)

	// Keep only last 5
	if len(pd.recentProfiles) > maxRecentProfiles {
		pd.recentProfiles = pd.recentProfiles[:maxRecentProfiles]
	}

	pd.saveRecentProfiles()
}

func (pd *ProfileDialog) Show() {
	// Create path label showing current directory
	pd.pathLabel = widget.NewLabel(pd.currentPath)
	pd.pathLabel.Wrapping = fyne.TextTruncate
	pd.pathLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create recent profiles list
	pd.recentList = widget.NewList(
		func() int {
			return len(pd.recentProfiles)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.HistoryIcon())
			label := widget.NewLabel("template")
			label.Truncation = fyne.TextTruncateEllipsis
			return container.NewHBox(icon, label)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			cont := obj.(*fyne.Container)
			label := cont.Objects[1].(*widget.Label)
			label.SetText(pd.recentProfiles[id])
		},
	)

	// Handle recent profile selection
	pd.recentList.OnSelected = func(id widget.ListItemID) {
		profilePath := pd.recentProfiles[id]

		// Check if file still exists
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			dialog.ShowError(err, pd.window)
			pd.recentList.UnselectAll()
			return
		}

		// Read and return
		content, err := os.ReadFile(profilePath)
		if err != nil {
			pd.callback("", err)
			pd.dialog.Hide()
			return
		}

		// Update recent profiles (move to front)
		pd.addRecentProfile(profilePath)

		// Store file path for external access
		pd.filePath = profilePath

		pd.callback(string(content), nil)
		pd.dialog.Hide()
	}

	// Create file list
	pd.fileList = widget.NewList(
		func() int {
			return len(pd.files)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			label := widget.NewLabel("template")
			return container.NewHBox(icon, label)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			cont := obj.(*fyne.Container)
			icon := cont.Objects[0].(*widget.Icon)
			label := cont.Objects[1].(*widget.Label)

			fileName := pd.files[id]
			label.SetText(fileName)

			fullPath := filepath.Join(pd.currentPath, fileName)
			fileInfo, err := os.Stat(fullPath)
			if err == nil && fileInfo.IsDir() {
				icon.SetResource(theme.FolderIcon())
			} else if strings.HasSuffix(fileName, ".share") || strings.HasSuffix(fileName, ".json") ||
				strings.HasSuffix(fileName, ".txt") || strings.HasSuffix(fileName, ".csv") ||
				strings.HasSuffix(fileName, ".parquet") {
				icon.SetResource(theme.DocumentIcon())
			} else {
				icon.SetResource(theme.FileIcon())
			}
		},
	)

	// Handle file selection
	pd.fileList.OnSelected = func(id widget.ListItemID) {
		fileName := pd.files[id]
		fullPath := filepath.Join(pd.currentPath, fileName)

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return
		}

		if fileInfo.IsDir() {
			// Navigate into directory
			pd.currentPath = fullPath
			pd.loadDirectory()
			pd.fileList.UnselectAll()
		} else {
			// File selected - read and return
			content, err := os.ReadFile(fullPath)
			if err != nil {
				pd.callback("", err)
				pd.dialog.Hide()
				return
			}

			// Add to recent profiles
			pd.addRecentProfile(fullPath)

			// Store file path for external access
			pd.filePath = fullPath

			pd.callback(string(content), nil)
			pd.dialog.Hide()
		}
	}

	// Create navigation buttons
	homeButton := widget.NewButtonWithIcon("Home", theme.HomeIcon(), func() {
		pd.currentPath = pd.homeDir
		pd.loadDirectory()
	})

	upButton := widget.NewButtonWithIcon("Up", theme.NavigateBackIcon(), func() {
		parent := filepath.Dir(pd.currentPath)
		if parent != pd.currentPath {
			pd.currentPath = parent
			pd.loadDirectory()
		}
	})

	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		pd.loadDirectory()
	})

	// Create filter info
	filterInfo := widget.NewLabel("Showing: .share, .json, .txt, .csv, and .parquet files, and directories")
	filterInfo.TextStyle = fyne.TextStyle{Italic: true}

	// Navigation toolbar
	navToolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(homeButton, upButton, refreshButton),
		nil,
		pd.pathLabel,
	)

	// Instructions
	instructions := widget.NewRichTextFromMarkdown("**Select a Delta Sharing profile or data file**\n\nSupported formats:\n- Delta Sharing profiles: .share, .json, .txt\n- Data files: .csv, .parquet, .json\n\nDouble-click a folder to navigate, or click a file to select it.")
	instructions.Wrapping = fyne.TextWrapWord

	// Create recent profiles card - always use the list
	// The list will show nothing if empty, or show items if populated
	recentCard := widget.NewCard("", "Recent Profiles", pd.recentList)

	// Create browser section
	browserCard := widget.NewCard("", "Browse Files", pd.fileList)

	// Split view with recent profiles on left and file browser on right
	splitContent := container.NewHSplit(recentCard, browserCard)
	splitContent.SetOffset(0.3) // 30% for recent profiles, 70% for file browser

	// Main content with better spacing
	content := container.NewBorder(
		container.NewVBox(
			instructions,
			widget.NewSeparator(),
			navToolbar,
			widget.NewSeparator(),
			filterInfo,
		),
		nil, nil, nil,
		splitContent,
	)

	// Create the custom dialog
	pd.dialog = dialog.NewCustom("Select Delta Sharing Profile", "Close", content, pd.window)

	// Make it much larger
	pd.dialog.Resize(fyne.NewSize(800, 600))

	// Load initial directory
	pd.loadDirectory()

	pd.dialog.Show()
}

func (pd *ProfileDialog) loadDirectory() {
	entries, err := os.ReadDir(pd.currentPath)
	if err != nil {
		dialog.ShowError(err, pd.window)
		return
	}

	pd.files = make([]string, 0)

	// Add directories first
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			pd.files = append(pd.files, entry.Name())
		}
	}

	// Add .share, .json, .txt, .csv, and .parquet files
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, ".share") || strings.HasSuffix(name, ".json") ||
				strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".csv") ||
				strings.HasSuffix(name, ".parquet") {
				pd.files = append(pd.files, name)
			}
		}
	}

	pd.pathLabel.SetText(pd.currentPath)
	pd.fileList.Refresh()
}
