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

package filter

import (
	"fmt"
	"strings"

	"github.com/magpierre/fyne-datatable/datatable"
)

// QueryFilter parses and evaluates SQL-like query strings.
// Supports syntax like: "age > 25 AND name = 'John' OR status = 'active'"
type QueryFilter struct {
	// Query is the SQL-like query string.
	Query string

	// Parsed query (cached after first parse).
	parsed *parsedQuery
}

// parsedQuery represents the parsed form of a query.
type parsedQuery struct {
	expressions []expression
	logicOps    []LogicOp
}

// expression represents a single comparison expression.
type expression struct {
	columnName string
	operator   CompareOp
	value      string
}

// Evaluate implements the Filter interface.
func (f *QueryFilter) Evaluate(row []datatable.Value, columnNames []string) (bool, error) {
	// Parse query if not already parsed
	if f.parsed == nil {
		parsed, err := parseQuery(f.Query, columnNames)
		if err != nil {
			return false, fmt.Errorf("failed to parse query: %w", err)
		}
		f.parsed = parsed
	}

	// Empty query matches all rows
	if f.parsed == nil || len(f.parsed.expressions) == 0 {
		return true, nil
	}

	// Build column name map for quick lookup
	columnMap := make(map[string]int, len(columnNames))
	for i, name := range columnNames {
		columnMap[strings.ToLower(name)] = i
	}

	// Evaluate first expression
	result, err := evaluateExpression(f.parsed.expressions[0], row, columnMap)
	if err != nil {
		return false, err
	}

	// Apply logical operators
	for i := 0; i < len(f.parsed.logicOps); i++ {
		nextResult, err := evaluateExpression(f.parsed.expressions[i+1], row, columnMap)
		if err != nil {
			return false, err
		}

		switch f.parsed.logicOps[i] {
		case LogicAND:
			result = result && nextResult
			if !result {
				return false, nil // Short-circuit on AND
			}
		case LogicOR:
			result = result || nextResult
			if result {
				return true, nil // Short-circuit on OR
			}
		}
	}

	return result, nil
}

// Description implements the Filter interface.
func (f *QueryFilter) Description() string {
	return f.Query
}

// parseQuery parses a query string into a structured form.
func parseQuery(queryStr string, columnNames []string) (*parsedQuery, error) {
	queryStr = strings.TrimSpace(queryStr)
	if queryStr == "" {
		return nil, nil
	}

	query := &parsedQuery{
		expressions: make([]expression, 0),
		logicOps:    make([]LogicOp, 0),
	}

	// Split by AND/OR operators
	parts := splitByLogicOps(queryStr)

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty query")
	}

	// Parse each part
	for _, part := range parts {
		if part.isOperator {
			if strings.ToUpper(part.text) == "AND" {
				query.logicOps = append(query.logicOps, LogicAND)
			} else if strings.ToUpper(part.text) == "OR" {
				query.logicOps = append(query.logicOps, LogicOR)
			}
		} else {
			expr, err := parseExpression(part.text, columnNames)
			if err != nil {
				return nil, err
			}
			query.expressions = append(query.expressions, expr)
		}
	}

	// Validate: should have N expressions and N-1 operators
	if len(query.logicOps) != len(query.expressions)-1 {
		return nil, fmt.Errorf("invalid query: mismatched expressions and operators")
	}

	return query, nil
}

type queryPart struct {
	text       string
	isOperator bool
}

// splitByLogicOps splits a query by AND/OR operators.
func splitByLogicOps(query string) []queryPart {
	parts := make([]queryPart, 0)
	current := ""
	i := 0

	for i < len(query) {
		// Check for AND
		if i+3 <= len(query) && strings.ToUpper(query[i:i+3]) == "AND" {
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

// parseExpression parses a single expression like "column = value".
func parseExpression(exprStr string, columnNames []string) (expression, error) {
	expr := expression{}
	exprStr = strings.TrimSpace(exprStr)

	// Try to find operators (in order of length to match >= before =)
	operators := []struct {
		op     CompareOp
		symbol string
	}{
		{OpGreaterOrEqual, ">="},
		{OpLessOrEqual, "<="},
		{OpNotEqual, "!="},
		{OpEqual, "="},
		{OpGreaterThan, ">"},
		{OpLessThan, "<"},
		{OpContains, "~"}, // Use ~ for contains
	}

	for _, opInfo := range operators {
		idx := strings.Index(exprStr, opInfo.symbol)
		if idx > 0 {
			columnName := strings.TrimSpace(exprStr[:idx])
			value := strings.TrimSpace(exprStr[idx+len(opInfo.symbol):])

			// Remove quotes from value if present
			value = strings.Trim(value, "\"'")

			expr.columnName = columnName
			expr.operator = opInfo.op
			expr.value = value

			// Validate column exists
			columnExists := false
			for _, col := range columnNames {
				if strings.EqualFold(col, columnName) {
					expr.columnName = col // Use exact case from schema
					columnExists = true
					break
				}
			}

			if !columnExists {
				return expr, fmt.Errorf("%w: %s", datatable.ErrColumnNotFound, columnName)
			}

			return expr, nil
		}
	}

	// If no operator found, treat as global contains search
	return expression{
		columnName: "", // Empty means search all columns
		operator:   OpContains,
		value:      exprStr,
	}, nil
}

// evaluateExpression evaluates a single expression against a row.
func evaluateExpression(expr expression, row []datatable.Value, columnMap map[string]int) (bool, error) {
	// Global search (no specific column)
	if expr.columnName == "" && expr.operator == OpContains {
		searchTerm := strings.ToLower(expr.value)
		for _, cell := range row {
			if cell.IsNull {
				continue
			}
			if strings.Contains(strings.ToLower(cell.Formatted), searchTerm) {
				return true, nil
			}
		}
		return false, nil
	}

	// Get column index
	colIdx, exists := columnMap[strings.ToLower(expr.columnName)]
	if !exists {
		return false, fmt.Errorf("%w: %s", datatable.ErrColumnNotFound, expr.columnName)
	}

	if colIdx >= len(row) {
		return false, fmt.Errorf("%w: %d", datatable.ErrInvalidColumn, colIdx)
	}

	cellValue := row[colIdx]

	// Null values don't match any comparison
	if cellValue.IsNull {
		return false, nil
	}

	// Use the compare function from SimpleFilter
	sf := &SimpleFilter{
		Column:   expr.columnName,
		Operator: expr.operator,
		Value:    expr.value,
	}

	return sf.compare(cellValue, expr.value, expr.operator)
}
