package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hieubanhh/dbhubMCP/internal/security"
)

// handleListTables handles the list_tables tool
func (s *Server) handleListTables(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tables, err := s.adapter.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	// Format result as JSON
	resultJSON, err := json.MarshalIndent(tables, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format result: %w", err)
	}

	return &CallToolResult{
		Content: []Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Found %d tables:\n\n%s", len(tables), string(resultJSON)),
			},
		},
	}, nil
}

// handleDescribeTable handles the describe_table tool
func (s *Server) handleDescribeTable(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	// Extract table name
	tableName, ok := args["table_name"].(string)
	if !ok || tableName == "" {
		return nil, fmt.Errorf("table_name is required and must be a string")
	}

	// Validate table name
	if err := security.SanitizeTableName(tableName); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	columns, err := s.adapter.DescribeTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}

	// Format result as JSON
	resultJSON, err := json.MarshalIndent(columns, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format result: %w", err)
	}

	return &CallToolResult{
		Content: []Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Table '%s' has %d columns:\n\n%s", tableName, len(columns), string(resultJSON)),
			},
		},
	}, nil
}

// handleExecuteQuery handles the execute_readonly_query tool
func (s *Server) handleExecuteQuery(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	// Extract query
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required and must be a string")
	}

	// Validate query (read-only check)
	if err := s.validator.ValidateReadOnlyQuery(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute query
	result, err := s.adapter.ExecuteQuery(ctx, query, s.maxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Format result
	var resultText string
	if result.RowCount == 0 {
		resultText = "Query returned no rows."
	} else {
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to format result: %w", err)
		}

		limitNote := ""
		if result.RowCount >= s.maxRows {
			limitNote = fmt.Sprintf("\n\n⚠️  Result limited to %d rows (MAX_ROWS setting)", s.maxRows)
		}

		resultText = fmt.Sprintf("Query executed successfully. Returned %d rows:\n\n%s%s",
			result.RowCount, string(resultJSON), limitNote)
	}

	return &CallToolResult{
		Content: []Content{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleExplainQuery handles the explain_query tool
func (s *Server) handleExplainQuery(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	// Extract query
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required and must be a string")
	}

	// Validate query (read-only check)
	if err := s.validator.ValidateReadOnlyQuery(query); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get query execution plan
	result, err := s.adapter.ExplainQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to explain query: %w", err)
	}

	// Format result as JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format result: %w", err)
	}

	return &CallToolResult{
		Content: []Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Query execution plan:\n\n%s", string(resultJSON)),
			},
		},
	}, nil
}
