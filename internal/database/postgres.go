package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresAdapter implements the Adapter interface for PostgreSQL
type PostgresAdapter struct {
	db             *sql.DB
	host           string
	port           int
	dbName         string
	user           string
	password       string
	maxConns       int
	maxIdleConns   int
	connTimeout    time.Duration
}

// NewPostgresAdapter creates a new PostgreSQL adapter
func NewPostgresAdapter(host string, port int, dbName, user, password string, maxConns, maxIdleConns int, connTimeout time.Duration) *PostgresAdapter {
	return &PostgresAdapter{
		host:         host,
		port:         port,
		dbName:       dbName,
		user:         user,
		password:     password,
		maxConns:     maxConns,
		maxIdleConns: maxIdleConns,
		connTimeout:  connTimeout,
	}
}

// Connect establishes a connection to PostgreSQL
func (a *PostgresAdapter) Connect(ctx context.Context) error {
	// Build connection string
	// format: host=localhost port=5432 user=myuser password=mypass dbname=mydb sslmode=disable
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		a.host, a.port, a.user, a.password, a.dbName, int(a.connTimeout.Seconds()))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(a.maxConns)
	db.SetMaxIdleConns(a.maxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, a.connTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	a.db = db
	return nil
}

// Close closes the PostgreSQL connection
func (a *PostgresAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (a *PostgresAdapter) Ping(ctx context.Context) error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}
	return a.db.PingContext(ctx)
}

// ListTables returns all tables in the PostgreSQL database
func (a *PostgresAdapter) ListTables(ctx context.Context) ([]TableInfo, error) {
	query := `
		SELECT
			table_name,
			table_schema,
			table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_name
	`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.TableName, &table.TableSchema, &table.TableType); err != nil {
			return nil, fmt.Errorf("failed to scan table info: %w", err)
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

// DescribeTable returns column information for a PostgreSQL table
func (a *PostgresAdapter) DescribeTable(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT
			column_name,
			data_type,
			is_nullable,
			COALESCE(column_default, '') as column_default,
			'' as column_key,
			'' as extra
		FROM information_schema.columns
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
			AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := a.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable, &col.ColumnDefault, &col.ColumnKey, &col.Extra); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table not found: %s", tableName)
	}

	return columns, nil
}

// ExecuteQuery executes a read-only query on PostgreSQL
func (a *PostgresAdapter) ExecuteQuery(ctx context.Context, query string, maxRows int) (*QueryResult, error) {
	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return rowsToResult(rows, maxRows)
}

// ExplainQuery returns the execution plan for a PostgreSQL query
func (a *PostgresAdapter) ExplainQuery(ctx context.Context, query string) (*QueryResult, error) {
	explainQuery := fmt.Sprintf("EXPLAIN %s", query)

	rows, err := a.db.QueryContext(ctx, explainQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to explain query: %w", err)
	}
	defer rows.Close()

	return rowsToResult(rows, 1000) // EXPLAIN results are typically small
}

// GetDBType returns the database type
func (a *PostgresAdapter) GetDBType() string {
	return "postgres"
}
