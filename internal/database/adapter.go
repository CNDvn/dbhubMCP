package database

import (
	"context"
	"database/sql"
)

// TableInfo represents metadata about a database table
type TableInfo struct {
	TableName   string `json:"table_name"`
	TableSchema string `json:"table_schema,omitempty"`
	TableType   string `json:"table_type,omitempty"`
}

// ColumnInfo represents metadata about a table column
type ColumnInfo struct {
	ColumnName    string `json:"column_name"`
	DataType      string `json:"data_type"`
	IsNullable    string `json:"is_nullable"`
	ColumnDefault string `json:"column_default,omitempty"`
	ColumnKey     string `json:"column_key,omitempty"`
	Extra         string `json:"extra,omitempty"`
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	RowCount int                     `json:"row_count"`
}

// Adapter defines the interface for database operations
type Adapter interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close() error

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error

	// ListTables returns a list of all tables in the database
	ListTables(ctx context.Context) ([]TableInfo, error)

	// DescribeTable returns column information for a specific table
	DescribeTable(ctx context.Context, tableName string) ([]ColumnInfo, error)

	// ExecuteQuery executes a read-only query and returns results
	ExecuteQuery(ctx context.Context, query string, maxRows int) (*QueryResult, error)

	// ExplainQuery returns the query execution plan
	ExplainQuery(ctx context.Context, query string) (*QueryResult, error)

	// GetDBType returns the database type (mysql, postgres, etc.)
	GetDBType() string
}

// rowsToResult converts sql.Rows to QueryResult
func rowsToResult(rows *sql.Rows, maxRows int) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	// Create a slice to hold column values
	columnCount := len(columns)
	values := make([]interface{}, columnCount)
	valuePtrs := make([]interface{}, columnCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		if rowCount >= maxRows {
			break
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		result.Rows = append(result.Rows, row)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result.RowCount = rowCount
	return result, nil
}
