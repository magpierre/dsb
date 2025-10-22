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
	"bytes"
	"fmt"

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
	w             fyne.Window
	codeEditor    *CursorTrackingEntry
	syntaxEditor  *SyntaxEditor
	outputText    *widget.RichText
	executeButton *widget.Button
	clearButton   *widget.Button
	saveButton    *widget.Button
	container     *fyne.Container
	interpreter   *interp.Interpreter
}

// CursorTrackingEntry extends widget.Entry to track cursor movements
type CursorTrackingEntry struct {
	widget.Entry
	onCursorChanged func()
}

// NewCursorTrackingEntry creates a new cursor tracking entry
func NewCursorTrackingEntry() *CursorTrackingEntry {
	entry := &CursorTrackingEntry{}
	entry.ExtendBaseWidget(entry)
	entry.MultiLine = true
	return entry
}

// TypedKey handles keyboard events and tracks cursor movements
func (e *CursorTrackingEntry) TypedKey(key *fyne.KeyEvent) {
	e.Entry.TypedKey(key)
	// Trigger cursor changed callback for cursor movement keys
	if e.onCursorChanged != nil {
		switch key.Name {
		case fyne.KeyUp, fyne.KeyDown, fyne.KeyLeft, fyne.KeyRight,
			fyne.KeyHome, fyne.KeyEnd, fyne.KeyPageUp, fyne.KeyPageDown:
			e.onCursorChanged()
		}
	}
}

// Tapped handles mouse clicks which can also move the cursor
func (e *CursorTrackingEntry) Tapped(ev *fyne.PointEvent) {
	e.Entry.Tapped(ev)
	if e.onCursorChanged != nil {
		e.onCursorChanged()
	}
}

// SetOnCursorChanged sets the callback for cursor position changes
func (e *CursorTrackingEntry) SetOnCursorChanged(callback func()) {
	e.onCursorChanged = callback
}

// NewGoEditor creates a new Go editor instance
func NewGoEditor(w fyne.Window) *GoEditor {
	ge := &GoEditor{
		w: w,
	}
	ge.createUI()
	return ge
}

// createUI builds the Go editor interface
func (ge *GoEditor) createUI() {
	// Create code editor (custom multiline text entry with cursor tracking)
	ge.codeEditor = NewCursorTrackingEntry()
	ge.codeEditor.SetPlaceHolder("// Enter your Go code here...\n// Example:\n// fmt.Println(\"Hello, World!\")\n// x := 42\n// fmt.Printf(\"Answer: %d\\n\", x)")
	ge.codeEditor.Wrapping = fyne.TextWrapOff

	// Create syntax editor for highlighted preview
	ge.syntaxEditor = NewSyntaxEditor()

	// Sync code editor changes to syntax editor (direct, synchronous)
	ge.codeEditor.OnChanged = func(text string) {
		// Update syntax highlighting directly (synchronous)
		ge.syntaxEditor.SetText(text)
		// Update highlighted line based on cursor position
		ge.updateHighlightedLine()
	}

	// Track cursor movements (for arrow keys and mouse clicks)
	ge.codeEditor.SetOnCursorChanged(func() {
		ge.updateHighlightedLine()
	})

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

	// Create scroll containers (no sync needed)
	editorScroll := container.NewScroll(ge.codeEditor)
	previewScroll := container.NewScroll(ge.syntaxEditor)

	// Create vertical split for editor: input (top) and syntax preview (bottom)
	editorSplit := container.NewVSplit(
		container.NewBorder(
			widget.NewLabel("Input Editor (type here):"),
			nil, nil, nil,
			editorScroll,
		),
		container.NewBorder(
			widget.NewLabel("Syntax Highlighted Preview:"),
			nil, nil, nil,
			previewScroll,
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

// executeCode executes the Go code using yaegi interpreter
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
	defer ge.executeButton.Enable()

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

// updateHighlightedLine updates the highlighted line in the syntax editor based on cursor position
func (ge *GoEditor) updateHighlightedLine() {
	// Get the cursor row (0-indexed)
	cursorRow := ge.codeEditor.CursorRow
	// Convert to 1-indexed for the syntax editor
	ge.syntaxEditor.SetHighlightedLine(cursorRow + 1)
}

// saveCode opens a file dialog and saves the code to the selected file
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

		defer writer.Close()

		// Write the file (synchronous)
		_, err = writer.Write([]byte(code))
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save file: %v", err), ge.w)
			return
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Code saved to %s (%d bytes)", writer.URI().Name(), len(code)), ge.w)
	}, ge.w)

	// Set default filename and filter for Go files
	saveDialog.SetFileName("code.go")
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".go"}))

	// Show the dialog
	saveDialog.Show()
}

// loadCode opens a file dialog and loads code from the selected file
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

		defer reader.Close()

		// Read the file (synchronous)
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

		// Update the editor
		ge.codeEditor.SetText(string(fileContent))
		ge.setOutput(fmt.Sprintf("Loaded code from %s (%d bytes)\n", reader.URI().Name(), len(fileContent)))
	}, ge.w)

	// Set filter for Go files
	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".go"}))

	// Show the dialog
	openDialog.Show()
}
