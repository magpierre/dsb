package windows

import (
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
	dialog   dialog.Dialog
	window   fyne.Window
	callback func(string, error)
	fileList *widget.List
	files    []string
	homeDir  string
	currentPath string
	pathLabel *widget.Label
}

func NewProfileDialog(w fyne.Window, callback func(string, error)) *ProfileDialog {
	pd := &ProfileDialog{
		window:   w,
		callback: callback,
		files:    make([]string, 0),
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	pd.homeDir = homeDir
	pd.currentPath = homeDir

	return pd
}

func (pd *ProfileDialog) Show() {
	// Create path label showing current directory
	pd.pathLabel = widget.NewLabel(pd.currentPath)
	pd.pathLabel.Wrapping = fyne.TextTruncate
	pd.pathLabel.TextStyle = fyne.TextStyle{Bold: true}

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
			} else if strings.HasSuffix(fileName, ".share") || strings.HasSuffix(fileName, ".json") || strings.HasSuffix(fileName, ".txt") {
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
	filterInfo := widget.NewLabel("Showing: .share, .json, and .txt files, and directories")
	filterInfo.TextStyle = fyne.TextStyle{Italic: true}

	// Navigation toolbar
	navToolbar := container.NewBorder(
		nil, nil,
		container.NewHBox(homeButton, upButton, refreshButton),
		nil,
		pd.pathLabel,
	)

	// Instructions
	instructions := widget.NewRichTextFromMarkdown("**Select a Delta Sharing profile file (.share, .json, or .txt)**\n\nDouble-click a folder to navigate, or click a file to select it.")
	instructions.Wrapping = fyne.TextWrapWord

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
		pd.fileList,
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

	// Add .share, .json, and .txt files
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, ".share") || strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".txt") {
				pd.files = append(pd.files, name)
			}
		}
	}

	pd.pathLabel.SetText(pd.currentPath)
	pd.fileList.Refresh()
}
