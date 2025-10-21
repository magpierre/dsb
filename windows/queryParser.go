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
	"fmt"
	"strconv"
	"strings"
)

// QueryParser handles parsing and evaluation of search expressions
type QueryParser struct {
	columnMap map[string]int // Maps column names to indices
}

// Comparison operators
type CompOp int

const (
	OpEqual CompOp = iota
	OpNotEqual
	OpGreater
	OpLess
	OpGreaterEqual
	OpLessEqual
	OpContains
)

// Expression represents a single comparison
type Expression struct {
	ColumnName string
	Operator   CompOp
	Value      string
}

// LogicalOp represents AND/OR operations
type LogicalOp int

const (
	LogicAND LogicalOp = iota
	LogicOR
)

// Query represents a complete query with multiple expressions
type Query struct {
	Expressions []Expression
	LogicOps    []LogicalOp // Operations between expressions
}

// NewQueryParser creates a new query parser with column name mapping
func NewQueryParser(headers []string) *QueryParser {
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.ToLower(header)] = i
	}
	return &QueryParser{columnMap: columnMap}
}

// ParseQuery parses a query string into a Query structure
func (qp *QueryParser) ParseQuery(queryStr string) (*Query, error) {
	if strings.TrimSpace(queryStr) == "" {
		return nil, nil
	}

	query := &Query{
		Expressions: make([]Expression, 0),
		LogicOps:    make([]LogicalOp, 0),
	}

	// Split by AND/OR (case-insensitive)
	parts := qp.splitByLogicOps(queryStr)

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty query")
	}

	// Parse each expression
	for _, part := range parts {
		if part.isOperator {
			if strings.ToUpper(part.text) == "AND" {
				query.LogicOps = append(query.LogicOps, LogicAND)
			} else if strings.ToUpper(part.text) == "OR" {
				query.LogicOps = append(query.LogicOps, LogicOR)
			}
		} else {
			expr, err := qp.parseExpression(part.text)
			if err != nil {
				return nil, err
			}
			query.Expressions = append(query.Expressions, expr)
		}
	}

	// Validate: should have N expressions and N-1 operators
	if len(query.LogicOps) != len(query.Expressions)-1 {
		return nil, fmt.Errorf("invalid query: mismatched expressions and operators")
	}

	return query, nil
}

type queryPart struct {
	text       string
	isOperator bool
}

// splitByLogicOps splits query by AND/OR while preserving the operators
func (qp *QueryParser) splitByLogicOps(query string) []queryPart {
	parts := make([]queryPart, 0)
	current := ""
	i := 0

	for i < len(query) {
		// Check for AND
		if i+3 <= len(query) && strings.ToUpper(query[i:i+3]) == "AND" {
			// Check if it's a word boundary
			if (i == 0 || isWhitespace(query[i-1])) && (i+3 >= len(query) || isWhitespace(query[i+3])) {
				if strings.TrimSpace(current) != "" {
					parts = append(parts, queryPart{text: strings.TrimSpace(current), isOperator: false})
					current = ""
				}
				parts = append(parts, queryPart{text: "AND", isOperator: true})
				i += 3
				continue
			}
		}

		// Check for OR
		if i+2 <= len(query) && strings.ToUpper(query[i:i+2]) == "OR" {
			// Check if it's a word boundary
			if (i == 0 || isWhitespace(query[i-1])) && (i+2 >= len(query) || isWhitespace(query[i+2])) {
				if strings.TrimSpace(current) != "" {
					parts = append(parts, queryPart{text: strings.TrimSpace(current), isOperator: false})
					current = ""
				}
				parts = append(parts, queryPart{text: "OR", isOperator: true})
				i += 2
				continue
			}
		}

		current += string(query[i])
		i++
	}

	if strings.TrimSpace(current) != "" {
		parts = append(parts, queryPart{text: strings.TrimSpace(current), isOperator: false})
	}

	return parts
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// parseExpression parses a single expression like "column = value"
func (qp *QueryParser) parseExpression(exprStr string) (Expression, error) {
	expr := Expression{}
	exprStr = strings.TrimSpace(exprStr)

	// Try to find operators (in order of length to match >= before =)
	operators := []struct {
		op     CompOp
		symbol string
	}{
		{OpGreaterEqual, ">="},
		{OpLessEqual, "<="},
		{OpNotEqual, "!="},
		{OpEqual, "="},
		{OpGreater, ">"},
		{OpLess, "<"},
		{OpContains, "~"}, // Use ~ for contains
	}

	for _, opInfo := range operators {
		idx := strings.Index(exprStr, opInfo.symbol)
		if idx > 0 {
			columnName := strings.TrimSpace(exprStr[:idx])
			value := strings.TrimSpace(exprStr[idx+len(opInfo.symbol):])

			// Remove quotes from value if present
			value = strings.Trim(value, "\"'")

			expr.ColumnName = columnName
			expr.Operator = opInfo.op
			expr.Value = value

			// Validate column exists
			if _, exists := qp.columnMap[strings.ToLower(columnName)]; !exists {
				return expr, fmt.Errorf("unknown column: %s", columnName)
			}

			return expr, nil
		}
	}

	// If no operator found, treat as contains search on all columns
	return Expression{
		ColumnName: "",
		Operator:   OpContains,
		Value:      exprStr,
	}, nil
}

// EvaluateRow evaluates a query against a data row
func (qp *QueryParser) EvaluateRow(query *Query, row []string, headers []string) bool {
	if query == nil || len(query.Expressions) == 0 {
		return true // Empty query matches all
	}

	// If only one expression, evaluate it
	if len(query.Expressions) == 1 {
		return qp.evaluateExpression(query.Expressions[0], row, headers)
	}

	// Evaluate first expression
	result := qp.evaluateExpression(query.Expressions[0], row, headers)

	// Apply logical operators
	for i := 0; i < len(query.LogicOps); i++ {
		nextResult := qp.evaluateExpression(query.Expressions[i+1], row, headers)

		switch query.LogicOps[i] {
		case LogicAND:
			result = result && nextResult
		case LogicOR:
			result = result || nextResult
		}
	}

	return result
}

// evaluateExpression evaluates a single expression against a row
func (qp *QueryParser) evaluateExpression(expr Expression, row []string, headers []string) bool {
	// If no column name, search all columns (contains)
	if expr.ColumnName == "" && expr.Operator == OpContains {
		searchTerm := strings.ToLower(expr.Value)
		for _, cell := range row {
			cellLower := strings.ToLower(cell)
			if strings.Contains(cellLower, searchTerm) {
				return true
			}
		}
		return false
	}

	// Get column index
	colIdx, exists := qp.columnMap[strings.ToLower(expr.ColumnName)]
	if !exists || colIdx >= len(row) {
		return false
	}

	cellValue := row[colIdx]

	// Perform comparison based on operator
	switch expr.Operator {
	case OpEqual:
		return strings.EqualFold(cellValue, expr.Value)

	case OpNotEqual:
		return !strings.EqualFold(cellValue, expr.Value)

	case OpContains:
		return strings.Contains(strings.ToLower(cellValue), strings.ToLower(expr.Value))

	case OpGreater, OpLess, OpGreaterEqual, OpLessEqual:
		return qp.compareNumeric(cellValue, expr.Value, expr.Operator)
	}

	return false
}

// compareNumeric compares two values numerically
func (qp *QueryParser) compareNumeric(cellValue, compareValue string, op CompOp) bool {
	// Try to parse as float
	cell, err1 := strconv.ParseFloat(strings.TrimSpace(cellValue), 64)
	compare, err2 := strconv.ParseFloat(strings.TrimSpace(compareValue), 64)

	if err1 != nil || err2 != nil {
		// If not numeric, do string comparison
		return qp.compareString(cellValue, compareValue, op)
	}

	switch op {
	case OpGreater:
		return cell > compare
	case OpLess:
		return cell < compare
	case OpGreaterEqual:
		return cell >= compare
	case OpLessEqual:
		return cell <= compare
	}

	return false
}

// compareString compares two strings lexicographically
func (qp *QueryParser) compareString(cellValue, compareValue string, op CompOp) bool {
	cmp := strings.Compare(strings.ToLower(cellValue), strings.ToLower(compareValue))

	switch op {
	case OpGreater:
		return cmp > 0
	case OpLess:
		return cmp < 0
	case OpGreaterEqual:
		return cmp >= 0
	case OpLessEqual:
		return cmp <= 0
	}

	return false
}
