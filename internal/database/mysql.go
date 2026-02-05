package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLAdapter implements the Adapter interface for MySQL
type MySQLAdapter struct {
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

// NewMySQLAdapter creates a new MySQL adapter
func NewMySQLAdapter(host string, port int, dbName, user, password string, maxConns, maxIdleConns int, connTimeout time.Duration) *MySQLAdapter {
	return &MySQLAdapter{
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

// Connect establishes a connection to MySQL
func (a *MySQLAdapter) Connect(ctx context.Context) error {
	// Build DSN (Data Source Name)
	// format: user:password@tcp(host:port)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=%s",
		a.user, a.password, a.host, a.port, a.dbName, a.connTimeout)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
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
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	a.db = db
	return nil
}

// Close closes the MySQL connection
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (a *MySQLAdapter) Ping(ctx context.Context) error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}
	return a.db.PingContext(ctx)
}

// ListTables returns all tables in the MySQL database
func (a *MySQLAdapter) ListTables(ctx context.Context) ([]TableInfo, error) {
	query := `
		SELECT
			TABLE_NAME as table_name,
			TABLE_SCHEMA as table_schema,
			TABLE_TYPE as table_type
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := a.db.QueryContext(ctx, query, a.dbName)
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

// DescribeTable returns column information for a MySQL table
func (a *MySQLAdapter) DescribeTable(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT
			COLUMN_NAME as column_name,
			DATA_TYPE as data_type,
			IS_NULLABLE as is_nullable,
			COALESCE(COLUMN_DEFAULT, '') as column_default,
			COALESCE(COLUMN_KEY, '') as column_key,
			COALESCE(EXTRA, '') as extra
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := a.db.QueryContext(ctx, query, a.dbName, tableName)
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

// ExecuteQuery executes a read-only query on MySQL
func (a *MySQLAdapter) ExecuteQuery(ctx context.Context, query string, maxRows int) (*QueryResult, error) {
	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	return rowsToResult(rows, maxRows)
}

// ExplainQuery returns the execution plan for a MySQL query
func (a *MySQLAdapter) ExplainQuery(ctx context.Context, query string) (*QueryResult, error) {
	explainQuery := fmt.Sprintf("EXPLAIN %s", query)

	rows, err := a.db.QueryContext(ctx, explainQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to explain query: %w", err)
	}
	defer rows.Close()

	return rowsToResult(rows, 1000) // EXPLAIN results are typically small
}

// GetDBType returns the database type
func (a *MySQLAdapter) GetDBType() string {
	return "mysql"
}
