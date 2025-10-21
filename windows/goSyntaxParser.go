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
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TokenType represents the type of a syntax token
type TokenType int

const (
	TokenKeyword     TokenType = iota // if, for, func, return, etc.
	TokenString                       // "...", `...`
	TokenComment                      // //, /* */
	TokenNumber                       // 123, 3.14, 0x1A
	TokenOperator                     // +, -, *, :=, ==
	TokenIdentifier                   // variable names
	TokenBuiltinType                  // int, string, bool, etc.
	TokenFunction                     // function names (after func keyword)
)

// StyledCell represents a single character with its style
type StyledCell struct {
	Rune  rune
	Style widget.TextGridStyle
}

// SyntaxStyles defines the color scheme for different token types
var SyntaxStyles = map[TokenType]widget.TextGridStyle{
	TokenKeyword: &widget.CustomTextGridStyle{
		FGColor:   color.NRGBA{R: 255, G: 20, B: 147, A: 255}, // Bright magenta/pink - very visible
		BGColor:   color.NRGBA{R: 50, G: 0, B: 50, A: 40},     // Subtle purple background
		TextStyle: fyne.TextStyle{Bold: true, Italic: false},
	},
	TokenString: &widget.CustomTextGridStyle{
		FGColor: color.NRGBA{R: 0, G: 180, B: 0, A: 255}, // Brighter green
	},
	TokenComment: &widget.CustomTextGridStyle{
		FGColor:   color.NRGBA{R: 128, G: 128, B: 128, A: 255}, // Gray
		TextStyle: fyne.TextStyle{Italic: true},
	},
	TokenNumber: &widget.CustomTextGridStyle{
		FGColor: color.NRGBA{R: 0, G: 150, B: 255, A: 255}, // Brighter blue
	},
	TokenOperator: &widget.CustomTextGridStyle{
		FGColor: color.NRGBA{R: 120, G: 120, B: 120, A: 255}, // Lighter gray
	},
	TokenBuiltinType: &widget.CustomTextGridStyle{
		FGColor:   color.NRGBA{R: 0, G: 180, B: 180, A: 255}, // Brighter teal
		TextStyle: fyne.TextStyle{Bold: true},
	},
	TokenFunction: &widget.CustomTextGridStyle{
		FGColor: color.NRGBA{R: 255, G: 140, B: 0, A: 255}, // Brighter orange
	},
	TokenIdentifier: nil, // Use default theme color
}

// Go keywords registry
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true,
	"continue": true, "default": true, "defer": true, "else": true,
	"fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true,
	"map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true,
	"var": true,
}

// Go built-in types
var goBuiltinTypes = map[string]bool{
	"bool": true, "byte": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true, "int": true,
	"int8": true, "int16": true, "int32": true, "int64": true,
	"rune": true, "string": true, "uint": true, "uint8": true,
	"uint16": true, "uint32": true, "uint64": true, "uintptr": true, "any": true, "comparable": true, "nil": true, "true": true, "false": true,
}

// ParseGoLine parses a single line and returns styled cells
// This is more efficient than parsing the entire document
func ParseGoLine(line string) []StyledCell {
	cells := make([]StyledCell, 0, len(line))
	runes := []rune(line)
	pos := 0

	for pos < len(runes) {
		// Skip whitespace - keep as default style
		if syntaxIsWhitespace(runes[pos]) {
			cells = append(cells, StyledCell{Rune: runes[pos], Style: nil})
			pos++
			continue
		}

		// Check for line comments (//)
		if pos+1 < len(runes) && runes[pos] == '/' && runes[pos+1] == '/' {
			// Rest of line is a comment
			for pos < len(runes) {
				cells = append(cells, StyledCell{
					Rune:  runes[pos],
					Style: SyntaxStyles[TokenComment],
				})
				pos++
			}
			break
		}

		// Check for strings (double quote or backtick)
		if runes[pos] == '"' || runes[pos] == '`' {
			endPos := syntaxParseString(runes, pos)
			for i := pos; i < endPos; i++ {
				cells = append(cells, StyledCell{
					Rune:  runes[i],
					Style: SyntaxStyles[TokenString],
				})
			}
			pos = endPos
			continue
		}

		// Check for numbers
		if syntaxIsDigit(runes[pos]) {
			endPos := syntaxParseNumber(runes, pos)
			for i := pos; i < endPos; i++ {
				cells = append(cells, StyledCell{
					Rune:  runes[i],
					Style: SyntaxStyles[TokenNumber],
				})
			}
			pos = endPos
			continue
		}

		// Check for identifiers/keywords
		if syntaxIsLetter(runes[pos]) || runes[pos] == '_' {
			endPos := syntaxParseIdentifier(runes, pos)
			word := string(runes[pos:endPos])

			var style widget.TextGridStyle
			if goKeywords[word] {
				style = SyntaxStyles[TokenKeyword]
			} else if goBuiltinTypes[word] {
				style = SyntaxStyles[TokenBuiltinType]
			} else {
				style = SyntaxStyles[TokenIdentifier]
			}

			for i := pos; i < endPos; i++ {
				cells = append(cells, StyledCell{
					Rune:  runes[i],
					Style: style,
				})
			}
			pos = endPos
			continue
		}

		// Operators and other characters - use operator style
		if syntaxIsOperator(runes[pos]) {
			cells = append(cells, StyledCell{
				Rune:  runes[pos],
				Style: SyntaxStyles[TokenOperator],
			})
		} else {
			// Default - no special style
			cells = append(cells, StyledCell{
				Rune:  runes[pos],
				Style: nil,
			})
		}
		pos++
	}

	return cells
}

// syntaxParseString parses a string literal starting at position start
func syntaxParseString(runes []rune, start int) int {
	quote := runes[start]
	pos := start + 1

	if quote == '`' {
		// Raw string - backticks
		for pos < len(runes) {
			if runes[pos] == '`' {
				return pos + 1
			}
			pos++
		}
		return pos // Unclosed string
	}

	// Double-quoted string with escapes
	for pos < len(runes) {
		if runes[pos] == '\\' && pos+1 < len(runes) {
			pos += 2 // Skip escaped character
			continue
		}
		if runes[pos] == '"' {
			return pos + 1
		}
		pos++
	}
	return pos // Unclosed string
}

// syntaxParseNumber parses a number literal starting at position start
func syntaxParseNumber(runes []rune, start int) int {
	pos := start
	// Simple number parsing: digits, dots, and e/E for scientific notation
	for pos < len(runes) {
		r := runes[pos]
		if !syntaxIsDigit(r) && r != '.' && r != 'e' && r != 'E' && r != 'x' && r != 'X' {
			break
		}
		pos++
	}
	return pos
}

// syntaxParseIdentifier parses an identifier starting at position start
func syntaxParseIdentifier(runes []rune, start int) int {
	pos := start
	for pos < len(runes) {
		r := runes[pos]
		if !syntaxIsLetter(r) && !syntaxIsDigit(r) && r != '_' {
			break
		}
		pos++
	}
	return pos
}

// syntaxIsWhitespace checks if a rune is whitespace
func syntaxIsWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// syntaxIsDigit checks if a rune is a digit
func syntaxIsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// syntaxIsLetter checks if a rune is a letter
func syntaxIsLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// syntaxIsOperator checks if a rune is an operator
func syntaxIsOperator(r rune) bool {
	operators := "+-*/%&|^<>=!:;,.()[]{}~"
	return strings.ContainsRune(operators, r)
}
