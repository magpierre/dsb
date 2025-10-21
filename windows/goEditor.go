package windows

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// GoEditor manages the Go code editor and output pane
type GoEditor struct {
	w                fyne.Window
	codeEditor       *widget.Entry
	syntaxEditor     *SyntaxEditor
	outputText       *widget.RichText
	executeButton    *widget.Button
	clearButton      *widget.Button
	saveButton       *widget.Button
	container        *fyne.Container
	interpreter      *interp.Interpreter
	editorScroll     *container.Scroll
	previewScroll    *container.Scroll
	stopScrollSync   chan bool
	lastCursorRow    int // Track last cursor row for change detection

	// Debouncing for syntax highlighting
	syntaxUpdateChan chan string
	stopSyntaxUpdate chan bool
	lastText         string // Track last processed text for incremental updates

	// Async code execution
	executionDone    chan bool
}

// NewGoEditor creates a new Go editor instance
func NewGoEditor(w fyne.Window) *GoEditor {
	ge := &GoEditor{
		w:                w,
		stopScrollSync:   make(chan bool),
		syntaxUpdateChan: make(chan string, 10),    // Buffered channel for text updates
		stopSyntaxUpdate: make(chan bool),
		executionDone:    make(chan bool),
	}
	ge.createUI()
	ge.startScrollSync()
	ge.startSyntaxHighlighter()
	return ge
}

// createUI builds the Go editor interface
func (ge *GoEditor) createUI() {
	// Create code editor (multiline text entry)
	ge.codeEditor = widget.NewMultiLineEntry()
	ge.codeEditor.SetPlaceHolder("// Enter your Go code here...\n// Example:\n// fmt.Println(\"Hello, World!\")\n// x := 42\n// fmt.Printf(\"Answer: %d\\n\", x)")
	ge.codeEditor.Wrapping = fyne.TextWrapOff

	// Create syntax editor for highlighted preview
	ge.syntaxEditor = NewSyntaxEditor()

	// Sync code editor changes to syntax editor (debounced via channel)
	ge.codeEditor.OnChanged = func(text string) {
		// Send update to debounced syntax highlighter goroutine
		select {
		case ge.syntaxUpdateChan <- text:
			// Successfully queued the update
		default:
			// Channel full, skip this update (will be superseded by next one)
		}
	}

	// Create output text area (read-only) with bold, colored text
	ge.outputText = widget.NewRichText()
	ge.outputText.Wrapping = fyne.TextWrapWord
	// Set initial placeholder text
	ge.setOutput("Output will appear here...")

	// Note: Buttons are now in the main window toolbar and only visible when Go Editor is selected
	// These button fields are kept for compatibility but no longer used in the UI
	ge.executeButton = widget.NewButtonWithIcon("Execute (interpreter mode)", theme.MediaPlayIcon(), func() {
		ge.executeCode()
	})

	ge.clearButton = widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		ge.clearOutput()
	})

	ge.saveButton = widget.NewButtonWithIcon("Save Code", theme.DocumentSaveIcon(), func() {
		ge.saveCode()
	})

	// Create scroll containers and store references for sync
	ge.editorScroll = container.NewScroll(ge.codeEditor)
	ge.previewScroll = container.NewScroll(ge.syntaxEditor)

	// Create vertical split for editor: input (top) and syntax preview (bottom)
	editorSplit := container.NewVSplit(
		container.NewBorder(
			widget.NewLabel("Input Editor (type here):"),
			nil, nil, nil,
			ge.editorScroll,
		),
		container.NewBorder(
			widget.NewLabel("Syntax Highlighted Preview:"),
			nil, nil, nil,
			ge.previewScroll,
		),
	)
	editorSplit.SetOffset(0.5) // 50/50 split

	// Wrap editor split in card
	editorCard := widget.NewCard("", "", editorSplit)

	// Create output section with label
	outputCard := widget.NewCard("", "Result of code execution:", container.NewBorder(
		nil, nil, nil, nil,
		container.NewScroll(ge.outputText),
	))

	// Create split container - editor on left, output on right
	splitContent := container.NewHSplit(
		editorCard,
		outputCard,
	)
	splitContent.SetOffset(0.5) // 50/50 split

	// Main container - toolbar buttons are now in the main window toolbar
	// Wrap splitContent in a border container to match the expected type
	ge.container = container.NewBorder(nil, nil, nil, nil, splitContent)
}

// GetContainer returns the main container for the Go editor
func (ge *GoEditor) GetContainer() *fyne.Container {
	return ge.container
}

// executeCode executes the Go code using yaegi interpreter asynchronously
func (ge *GoEditor) executeCode() {
	code := ge.codeEditor.Text

	if code == "" {
		ge.setOutput("Error: No code to execute\n")
		return
	}

	// Clear previous output and show executing message
	ge.setOutput("Executing Go code...\n")
	ge.appendOutput("----------------------------------------\n")

	// Disable execute button during execution to prevent concurrent runs
	ge.executeButton.Disable()

	// Run the code execution in a goroutine to avoid blocking the UI
	go func() {
		defer func() {
			// Re-enable execute button when done
			ge.executeButton.Enable()
			// Signal completion
			select {
			case ge.executionDone <- true:
			default:
			}
		}()

		// Create a buffer to capture output
		var outputBuffer bytes.Buffer

		// Create interpreter with custom stdout
		i := interp.New(interp.Options{
			Stdout: &outputBuffer,
			Stderr: &outputBuffer,
		})

		// Use the standard library
		if err := i.Use(stdlib.Symbols); err != nil {
			ge.appendOutput(fmt.Sprintf("Error loading stdlib: %v\n", err))
			return
		}

		// Use delta_sharing symbols
		if err := i.Use(DeltaSharingSymbols); err != nil {
			ge.appendOutput(fmt.Sprintf("Error loading delta_sharing: %v\n", err))
			return
		}

		// Wrap code in a main function context if it's not already
		wrappedCode := fmt.Sprintf(`package main
import (
	"fmt"
	"context"
	"encoding/json"
	delta_sharing "github.com/magpierre/go_delta_sharing_client"
)

func main() {
	%s
}
`, code)

		// Try to evaluate the code
		_, execError := i.Eval(wrappedCode)

		// Display captured output (code output should be bold)
		capturedOutput := outputBuffer.String()
		if capturedOutput != "" {
			ge.appendOutputBold(capturedOutput)
		}

		// Display any execution errors (normal text)
		if execError != nil {
			ge.appendOutput(fmt.Sprintf("\nExecution error: %v\n", execError))
		}

		ge.appendOutput("----------------------------------------\n")
		ge.appendOutput("Execution completed.\n")
	}()
}

// setOutput replaces the output window content with normal text
func (ge *GoEditor) setOutput(text string) {
	segment := &widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			TextStyle: fyne.TextStyle{Bold: false},
			ColorName: theme.ColorNameForeground,
		},
	}

	ge.outputText.Segments = []widget.RichTextSegment{segment}
	ge.outputText.Refresh()
}

// appendOutput adds text to the output window with normal styling
func (ge *GoEditor) appendOutput(text string) {
	ge.appendOutputStyled(text, false)
}

// appendOutputBold adds text to the output window with bold styling
func (ge *GoEditor) appendOutputBold(text string) {
	ge.appendOutputStyled(text, true)
}

// appendOutputStyled adds text to the output window with specified styling
func (ge *GoEditor) appendOutputStyled(text string, bold bool) {
	// Append new segment to existing segments
	segment := &widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			TextStyle: fyne.TextStyle{Bold: bold},
			ColorName: theme.ColorNameForeground,
		},
	}

	ge.outputText.Segments = append(ge.outputText.Segments, segment)
	ge.outputText.Refresh()
}

// clearOutput clears the output window
func (ge *GoEditor) clearOutput() {
	ge.outputText.Segments = []widget.RichTextSegment{}
	ge.outputText.Refresh()
}

// SetCode sets the code in the editor (useful for loading templates or examples)
func (ge *GoEditor) SetCode(code string) {
	ge.codeEditor.SetText(code)
}

// GetCode returns the current code in the editor
func (ge *GoEditor) GetCode() string {
	return ge.codeEditor.Text
}

// saveCode opens a file dialog and saves the code to the selected file asynchronously
func (ge *GoEditor) saveCode() {
	code := ge.codeEditor.Text

	if code == "" {
		dialog.ShowInformation("Nothing to Save", "The editor is empty. Please write some code before saving.", ge.w)
		return
	}

	// Create a save file dialog
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ge.w)
			return
		}
		if writer == nil {
			// User cancelled the dialog
			return
		}

		// Show progress dialog
		progress := dialog.NewInformation("Saving", "Saving file...", ge.w)
		progress.Show()

		// Write the file in a goroutine to avoid blocking the UI
		go func() {
			defer writer.Close()
			defer progress.Hide()

			_, err = writer.Write([]byte(code))
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to save file: %v", err), ge.w)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Code saved to %s (%d bytes)", writer.URI().Name(), len(code)), ge.w)
		}()
	}, ge.w)

	// Set default filename and filter for Go files
	saveDialog.SetFileName("code.go")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".go"}))

	// Show the dialog
	saveDialog.Show()
}

// loadCode opens a file dialog and loads code from the selected file asynchronously
func (ge *GoEditor) loadCode() {
	// Create an open file dialog
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ge.w)
			return
		}
		if reader == nil {
			// User cancelled the dialog
			return
		}

		// Show progress dialog
		progress := dialog.NewInformation("Loading", "Loading file...", ge.w)
		progress.Show()

		// Read the file in a goroutine to avoid blocking the UI
		go func() {
			defer reader.Close()
			defer progress.Hide()

			fileContent := make([]byte, 0)
			buffer := make([]byte, 4096) // Larger buffer for better performance

			for {
				n, err := reader.Read(buffer)
				if n > 0 {
					fileContent = append(fileContent, buffer[:n]...)
				}
				if err != nil {
					break
				}
			}

			// Update the UI on the main thread
			ge.codeEditor.SetText(string(fileContent))
			ge.setOutput(fmt.Sprintf("Loaded code from %s (%d bytes)\n", reader.URI().Name(), len(fileContent)))
		}()
	}, ge.w)

	// Set filter for Go files
	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".go"}))

	// Show the dialog
	openDialog.Show()
}

// startScrollSync starts a goroutine that synchronizes scroll positions
func (ge *GoEditor) startScrollSync() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms (optimized from 50ms)
		defer ticker.Stop()

		lastUpdateTime := time.Now()
		updateCooldown := 50 * time.Millisecond // Minimum time between updates

		for {
			select {
			case <-ge.stopScrollSync:
				return
			case <-ticker.C:
				// Only update if enough time has passed since last update
				if time.Since(lastUpdateTime) >= updateCooldown {
					if ge.syncScrollPosition() {
						lastUpdateTime = time.Now()
					}
				}
			}
		}
	}()
}

// syncScrollPosition synchronizes the preview to center on the current cursor line
// Returns true if an update was performed, false otherwise
func (ge *GoEditor) syncScrollPosition() bool {
	if ge.editorScroll == nil || ge.previewScroll == nil || ge.codeEditor == nil || ge.syntaxEditor == nil {
		return false
	}

	// Get the current cursor row (0-indexed)
	cursorRow := ge.codeEditor.CursorRow

	// Only update if cursor row has changed (avoid unnecessary refreshes)
	if cursorRow == ge.lastCursorRow {
		return false
	}
	ge.lastCursorRow = cursorRow

	// Get the TextGrid from syntax editor
	textGrid := ge.syntaxEditor.GetTextGrid()
	if textGrid == nil {
		return false
	}

	// Get viewport height in pixels
	viewportHeight := ge.previewScroll.Size().Height

	// Get row height - calculate from TextGrid dimensions
	// Each row in TextGrid has a consistent height
	totalRows := len(textGrid.Rows)
	if totalRows <= 0 {
		return false
	}
	contentHeight := textGrid.MinSize().Height
	rowHeight := contentHeight / float32(totalRows)
	if rowHeight <= 0 {
		return false
	}

	// Calculate how many rows fit in the viewport
	rowsInView := int(viewportHeight / rowHeight)
	if rowsInView <= 0 {
		rowsInView = 1
	}

	// Calculate the offset to center the cursor row
	// We want the cursor row to be in the middle of the viewport
	rowsAboveCenter := rowsInView / 2

	// Calculate the target row to be at the top of the viewport
	topRow := cursorRow - rowsAboveCenter
	if topRow < 0 {
		topRow = 0
	}

	// Calculate the Y offset in pixels
	targetOffsetY := float32(topRow) * rowHeight

	// Check bounds (contentHeight already calculated above)
	maxOffsetY := contentHeight - viewportHeight
	if maxOffsetY < 0 {
		maxOffsetY = 0
	}

	// Clamp the offset to valid range
	if targetOffsetY > maxOffsetY {
		targetOffsetY = maxOffsetY
	}
	if targetOffsetY < 0 {
		targetOffsetY = 0
	}

	// Update the scroll offset
	ge.previewScroll.Offset.Y = targetOffsetY
	ge.previewScroll.Refresh()

	return true
}

// StopScrollSync stops the scroll synchronization goroutine
func (ge *GoEditor) StopScrollSync() {
	close(ge.stopScrollSync)
}

// startSyntaxHighlighter starts a goroutine that debounces syntax highlighting updates
func (ge *GoEditor) startSyntaxHighlighter() {
	go func() {
		var debounceTimer *time.Timer
		const debounceDelay = 150 * time.Millisecond // Wait 150ms after last keystroke

		for {
			select {
			case <-ge.stopSyntaxUpdate:
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				return

			case text := <-ge.syntaxUpdateChan:
				// Reset the debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				// Create a new timer that will update syntax after delay
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					// Perform incremental syntax highlighting in this goroutine (not UI thread)
					ge.updateSyntaxIncremental(text)
				})
			}
		}
	}()
}

// updateSyntaxIncremental performs incremental syntax highlighting update
func (ge *GoEditor) updateSyntaxIncremental(newText string) {
	// Split both old and new text into lines
	oldLines := strings.Split(ge.lastText, "\n")
	newLines := strings.Split(newText, "\n")

	// If the number of lines changed significantly or it's the first update, do a full update
	if ge.lastText == "" || abs(len(newLines)-len(oldLines)) > 10 {
		ge.syntaxEditor.SetText(newText)
		ge.lastText = newText
		return
	}

	// Find the range of changed lines
	firstChanged := -1
	lastChanged := -1

	minLen := len(oldLines)
	if len(newLines) < minLen {
		minLen = len(newLines)
	}

	// Find first changed line
	for i := 0; i < minLen; i++ {
		if oldLines[i] != newLines[i] {
			firstChanged = i
			break
		}
	}

	// If no changes found in common lines, check if lines were added/removed
	if firstChanged == -1 {
		if len(oldLines) != len(newLines) {
			firstChanged = minLen
		} else {
			// No changes at all
			return
		}
	}

	// Find last changed line (scan from end)
	for i := 0; i < minLen-firstChanged; i++ {
		oldIdx := len(oldLines) - 1 - i
		newIdx := len(newLines) - 1 - i
		if oldIdx >= 0 && newIdx >= 0 && oldLines[oldIdx] != newLines[newIdx] {
			lastChanged = newIdx
			break
		}
	}

	if lastChanged == -1 {
		lastChanged = len(newLines) - 1
	}

	// If more than 20% of lines changed, do a full update instead
	changedLines := lastChanged - firstChanged + 1
	if changedLines > len(newLines)/5 {
		ge.syntaxEditor.SetText(newText)
		ge.lastText = newText
		return
	}

	// Perform incremental update
	ge.syntaxEditor.UpdateLineRange(firstChanged, lastChanged, newText)
	ge.lastText = newText
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// StopSyntaxHighlighter stops the syntax highlighting goroutine
func (ge *GoEditor) StopSyntaxHighlighter() {
	close(ge.stopSyntaxUpdate)
}

// Cleanup stops all background goroutines
// Call this when the editor is being closed to prevent resource leaks
func (ge *GoEditor) Cleanup() {
	// Stop all background goroutines
	ge.StopScrollSync()
	ge.StopSyntaxHighlighter()

	// Close execution done channel
	close(ge.executionDone)
}
