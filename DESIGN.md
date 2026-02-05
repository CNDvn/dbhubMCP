# DBHub MCP Server - Design Documentation

## Design Overview

This document explains the architecture, implementation decisions, and rationale behind the DBHub MCP Server.

## Architecture Decisions

### 1. Layered Architecture

**Decision:** Implement a strict layered architecture with clear boundaries.

**Rationale:**
- **Maintainability**: Each layer has a single responsibility
- **Testability**: Layers can be tested in isolation with mocks
- **Extensibility**: New databases can be added without modifying upper layers
- **Security**: Security layer acts as a centralized gatekeeper

**Trade-offs:**
- More abstraction adds slight complexity
- More files to navigate
- But: Significantly better long-term maintainability

### 2. Interface-Based Database Adapters

**Decision:** Define a `database.Adapter` interface with all database operations.

**Rationale:**
- **Polymorphism**: Server code doesn't care about database type
- **Easy Extension**: Adding a new database is just implementing the interface
- **Consistent API**: All databases expose identical operations
- **Testability**: Easy to create mock adapters for testing

**Implementation:**
```go
type Adapter interface {
    Connect(ctx context.Context) error
    ListTables(ctx context.Context) ([]TableInfo, error)
    // ... other methods
}
```

### 3. STDIO Transport (Not HTTP)

**Decision:** Use STDIO for MCP protocol transport.

**Rationale:**
- **MCP Specification**: STDIO is the standard for MCP servers
- **Security**: No network ports exposed
- **Integration**: Works seamlessly with Claude Desktop
- **Simplicity**: No TLS, authentication, or routing complexity

**Implementation:**
- Read JSON-RPC messages from `stdin`
- Write responses to `stdout`
- All logs go to `stderr`

### 4. Multi-Layer Read-Only Enforcement

**Decision:** Implement read-only enforcement at three levels.

**Layer 1 - SQL Validation:**
```go
// In security/validator.go
func (v *Validator) ValidateReadOnlyQuery(query string) error {
    // Block write keywords: INSERT, UPDATE, DELETE, etc.
    // Enforce SELECT/EXPLAIN/SHOW/DESCRIBE start
    // Detect SQL injection patterns
}
```

**Layer 2 - Database Permissions:**
```sql
-- Use read-only database users
GRANT SELECT ON database.* TO 'readonly_user'@'%';
```

**Layer 3 - Query Limits:**
```go
// Timeouts via context.WithTimeout()
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

// Row limits via maxRows parameter
result, err := adapter.ExecuteQuery(ctx, query, maxRows)
```

**Rationale:**
- **Defense in Depth**: Multiple independent security layers
- **Fail-Safe**: If one layer fails, others still protect
- **Clear Errors**: Each layer provides specific error messages

### 5. Context-Based Timeout Management

**Decision:** Use `context.Context` for all database operations.

**Rationale:**
- **Resource Control**: Prevents runaway queries
- **Graceful Cancellation**: Cancel in-progress queries on shutdown
- **Standard Pattern**: Idiomatic Go for managing request lifecycles

**Implementation:**
```go
ctx, cancel := context.WithTimeout(ctx, cfg.QueryTimeout)
defer cancel()
result, err := db.QueryContext(ctx, query)
```

### 6. Connection Pooling

**Decision:** Configure connection pools on all adapters.

**Rationale:**
- **Performance**: Reuse connections instead of creating new ones
- **Resource Limits**: Prevent overwhelming the database
- **Latency**: Eliminate connection establishment overhead

**Configuration:**
```go
db.SetMaxOpenConns(maxConns)        // Total connections
db.SetMaxIdleConns(maxIdleConns)    // Keep-alive pool
db.SetConnMaxLifetime(time.Hour)    // Refresh connections
```

### 7. Structured Error Messages

**Decision:** Return detailed, structured errors to LLMs.

**Rationale:**
- **LLM Understanding**: LLMs need clear error messages to adjust behavior
- **Debugging**: Developers can quickly identify issues
- **Security**: Don't leak sensitive information in errors

**Example:**
```go
return nil, fmt.Errorf("query validation failed: write operation detected: %s is not allowed", keyword)
```

## Implementation Patterns

### Pattern 1: Tool Registration

Tools are registered in `registerTools()` with a declarative schema:

```go
s.RegisterTool(Tool{
    Name:        "execute_readonly_query",
    Description: "Executes a read-only SQL query...",
    InputSchema: InputSchema{
        Type: "object",
        Properties: map[string]Property{
            "query": {
                Type:        "string",
                Description: "The SQL SELECT query to execute",
            },
        },
        Required: []string{"query"},
    },
}, s.handleExecuteQuery)
```

This makes tools self-documenting and easy to add.

### Pattern 2: Row Conversion

The `rowsToResult()` function provides consistent conversion from `sql.Rows` to JSON-serializable results:

```go
func rowsToResult(rows *sql.Rows, maxRows int) (*QueryResult, error) {
    // 1. Get column names
    // 2. Create value holders
    // 3. Scan rows into map[string]interface{}
    // 4. Convert []byte to string for JSON
    // 5. Enforce maxRows limit
}
```

This shared function ensures consistent behavior across all databases.

### Pattern 3: Configuration via Environment

All configuration is environment-based for 12-factor app compliance:

```go
cfg, err := config.LoadFromEnv()
// Validates required fields
// Provides sensible defaults
// Returns structured Config object
```

### Pattern 4: Graceful Shutdown

Signal handling enables clean shutdown:

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    cancel() // Cancel context
}()
```

## Security Implementation

### SQL Injection Prevention

**1. Keyword Blocking:**
```go
writeKeywords := []string{
    "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
    "TRUNCATE", "REPLACE", "MERGE", "GRANT", "REVOKE",
}
```

**2. Pattern Detection:**
```go
sqlInjectionPatterns := []*regexp.Regexp{
    regexp.MustCompile(`(?i);\s*(DROP|DELETE|...)`), // Multiple statements
    regexp.MustCompile(`(?i)--`),                     // SQL comments
    regexp.MustCompile(`(?i)/\*.*\*/`),               // Multi-line comments
}
```

**3. Word Boundary Matching:**
```go
pattern := regexp.MustCompile(`\b` + keyword + `\b`)
```
This prevents false positives like "SELECTION" matching "SELECT".

### Table Name Validation

```go
func SanitizeTableName(tableName string) error {
    // 1. Non-empty check
    // 2. Allow: alphanumeric, underscore, dot, quotes
    // 3. Block: SQL keywords, comments, semicolons
}
```

## Database-Specific Considerations

### MySQL Implementation

**DSN Format:**
```
user:password@tcp(host:port)/dbname?parseTime=true&timeout=10s
```

**Information Schema Queries:**
```sql
SELECT TABLE_NAME, TABLE_SCHEMA, TABLE_TYPE
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = ?
```

### PostgreSQL Implementation

**Connection String:**
```
host=localhost port=5432 user=user password=pass dbname=db sslmode=disable
```

**Schema Filtering:**
```sql
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
```

**Parameterization:**
PostgreSQL uses `$1, $2, ...` placeholders instead of `?`.

## Performance Considerations

### Connection Pooling

- **MaxOpenConns**: Limits total connections to prevent database overload
- **MaxIdleConns**: Keeps connections alive for reuse
- **ConnMaxLifetime**: Refreshes connections to prevent stale connections

### Query Limits

- **MAX_ROWS**: Prevents memory exhaustion on large result sets
- **QUERY_TIMEOUT_SEC**: Prevents long-running queries from blocking

### Efficient Result Serialization

Results are streamed and limited:
```go
for rows.Next() {
    if rowCount >= maxRows {
        break // Stop fetching
    }
    // Process row
}
```

## Future Extensibility

### Adding New Databases

1. Create `internal/database/newdb.go`
2. Implement `Adapter` interface
3. Add driver import: `_ "github.com/newdb/driver"`
4. Register in `main.go`:
```go
case "newdb":
    adapter = database.NewNewDBAdapter(...)
```

### Adding New Tools

1. Define tool schema in `registerTools()`
2. Implement handler in `handlers.go`
3. Follow pattern: validate args → query DB → format result

### Adding New Security Rules

1. Add keywords to `writeKeywords`
2. Add patterns to `sqlInjectionPatterns`
3. Update `ValidateReadOnlyQuery()` logic

## Testing Strategy

### Unit Tests
- Mock `database.Adapter` interface
- Test SQL validation with edge cases
- Test tool handlers with various inputs

### Integration Tests
- Test against real MySQL/PostgreSQL instances
- Use Docker containers for test databases
- Test full MCP protocol flow

### Security Tests
- Test SQL injection attempts
- Test write operation blocking
- Test query timeout enforcement

## Common Pitfalls and Solutions

### Pitfall 1: False Positives in SQL Validation

**Problem:** "SELECTION" matches "SELECT" keyword.

**Solution:** Use word boundary regex: `\bSELECT\b`

### Pitfall 2: Stdout/Stderr Confusion

**Problem:** Logs appear in MCP protocol stream.

**Solution:** All logs to stderr, MCP protocol to stdout.

### Pitfall 3: Context Not Propagated

**Problem:** Queries don't respect timeouts.

**Solution:** Always use `QueryContext()`, not `Query()`.

### Pitfall 4: []byte in JSON

**Problem:** Binary data in JSON responses.

**Solution:** Convert `[]byte` to `string` in `rowsToResult()`.

## Design Trade-offs

### Trade-off 1: Regex vs. SQL Parser

**Chosen:** Regex validation

**Pros:**
- Simple, fast, no dependencies
- Covers 99% of cases

**Cons:**
- May have edge cases
- Not as accurate as full SQL parser

**Justification:** Simplicity wins for this use case. Combined with DB-level permissions, security is adequate.

### Trade-off 2: lib/pq vs. pgx

**Chosen:** lib/pq (pure Go)

**Pros:**
- Pure Go, easier to compile
- Stable, well-tested
- Works with database/sql interface

**Cons:**
- pgx has better performance
- pgx has more features

**Justification:** Consistency with database/sql pattern. Performance difference negligible for read-only queries.

### Trade-off 3: STDIO vs. HTTP

**Chosen:** STDIO (per MCP spec)

**Pros:**
- MCP standard
- No network exposure
- Simpler implementation

**Cons:**
- Can't use curl/Postman for testing
- Requires MCP client

**Justification:** Following MCP specification ensures compatibility with Claude Desktop and other MCP clients.

## Conclusion

This design prioritizes:
1. **Security** - Multi-layer read-only enforcement
2. **Extensibility** - Easy to add databases/tools
3. **Reliability** - Timeouts, limits, error handling
4. **Simplicity** - Clear layers, minimal dependencies
5. **Standards** - Follows MCP specification

The architecture supports the core requirements while remaining maintainable and extensible for future enhancements.
