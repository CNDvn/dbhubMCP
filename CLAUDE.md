# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DBHub MCP Server is a Model Context Protocol (MCP) server written in Go that provides safe, read-only database operations for MySQL and PostgreSQL. It enables LLMs to interact with databases through a standardized protocol while enforcing strict read-only access.

**Key Technologies:**
- Go 1.21+
- MCP Protocol (STDIO transport)
- MySQL (`go-sql-driver/mysql`)
- PostgreSQL (`lib/pq`)
- JSON-RPC 2.0

## Build and Run Commands

```bash
# Download dependencies
go mod download

# Build the server
go build -o dbhub-mcp-server ./cmd/server

# Run with environment variables
export $(cat .env | xargs) && ./dbhub-mcp-server

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./internal/security

# Check for race conditions
go test -race ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o dbhub-mcp-server-linux ./cmd/server
GOOS=windows GOARCH=amd64 go build -o dbhub-mcp-server.exe ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o dbhub-mcp-server-mac ./cmd/server
```

## Architecture Overview

The codebase follows a layered architecture with clear separation of concerns:

### Layer Flow
```
MCP Client (Claude/LLM)
    ↓ STDIO (JSON-RPC)
Transport Layer (internal/mcp/transport.go)
    ↓
MCP Server (internal/mcp/server.go)
    ↓
Tool Handlers (internal/mcp/handlers.go)
    ↓
Security Layer (internal/security/validator.go)
    ↓
Database Adapter Interface (internal/database/adapter.go)
    ↓
Database Implementations (mysql.go / postgres.go)
```

### Key Components

**1. MCP Protocol Layer** (`internal/mcp/`)
- `protocol.go` - MCP protocol structures (Request, Response, Tool definitions)
- `transport.go` - STDIO-based JSON-RPC communication
- `server.go` - Main server orchestration and request routing
- `handlers.go` - Individual tool implementations

**2. Database Layer** (`internal/database/`)
- `adapter.go` - Database adapter interface definition
- `mysql.go` - MySQL-specific implementation
- `postgres.go` - PostgreSQL-specific implementation
- Shared `rowsToResult()` function for consistent query result formatting

**3. Security Layer** (`internal/security/`)
- `validator.go` - SQL validation and read-only enforcement
- Blocks write operations (INSERT, UPDATE, DELETE, DROP, etc.)
- SQL injection pattern detection
- Table name sanitization

**4. Configuration** (`internal/config/`)
- `config.go` - Environment variable loading and validation

**5. Entry Point** (`cmd/server/`)
- `main.go` - Server initialization, graceful shutdown, signal handling

### Design Patterns

**Interface-Based Adapters**: The `database.Adapter` interface allows easy addition of new databases. To add a new database:
1. Create a new file in `internal/database/` (e.g., `oracle.go`)
2. Implement all `Adapter` interface methods
3. Add a case in `main.go` to instantiate the new adapter

**Multi-Layer Security**:
- Layer 1: SQL validation (regex + keyword blocking)
- Layer 2: Database-level read-only user permissions
- Layer 3: Query timeouts and row limits

**Context Propagation**: All database operations accept `context.Context` for proper cancellation and timeout handling.

**Connection Pooling**: Each adapter configures connection pools via `SetMaxOpenConns()` and `SetMaxIdleConns()`.

## MCP Protocol Implementation

The server implements MCP protocol methods:
- `initialize` - Server capability negotiation
- `initialized` - Notification of client initialization
- `tools/list` - Returns available tools
- `tools/call` - Executes a specific tool
- `ping` - Health check

All communication follows JSON-RPC 2.0 specification over STDIO.

## Security Considerations

**Read-Only Enforcement:**
- SQL validator blocks all write keywords
- Queries must start with SELECT, EXPLAIN, DESCRIBE, SHOW, or WITH
- SQL comment patterns are blocked (`--`, `/* */`)

**When modifying security layer:**
- Update `writeKeywords` slice for new dangerous keywords
- Add new injection patterns to `sqlInjectionPatterns`
- Maintain word boundary matching to avoid false positives

**Query Limits:**
- `QUERY_TIMEOUT_SEC` enforces maximum query execution time
- `MAX_ROWS` limits result set size to prevent memory issues

## Adding New Tools

To add a new MCP tool:

1. Define the tool in `server.go` `registerTools()`:
```go
s.RegisterTool(Tool{
    Name:        "your_tool_name",
    Description: "Tool description",
    InputSchema: InputSchema{
        Type: "object",
        Properties: map[string]Property{
            "param_name": {
                Type:        "string",
                Description: "Parameter description",
            },
        },
        Required: []string{"param_name"},
    },
}, s.handleYourTool)
```

2. Implement handler in `handlers.go`:
```go
func (s *Server) handleYourTool(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
    // Extract and validate arguments
    // Add timeout to context
    // Execute database operation via adapter
    // Format and return result
}
```

## Database Adapter Interface

When adding a new database, implement these methods:
- `Connect(ctx)` - Establish connection with pool configuration
- `Close()` - Clean up resources
- `Ping(ctx)` - Health check
- `ListTables(ctx)` - Return all tables as `[]TableInfo`
- `DescribeTable(ctx, tableName)` - Return columns as `[]ColumnInfo`
- `ExecuteQuery(ctx, query, maxRows)` - Execute and return `*QueryResult`
- `ExplainQuery(ctx, query)` - Return execution plan as `*QueryResult`
- `GetDBType()` - Return database type string

Use the `rowsToResult()` helper function to convert `sql.Rows` to `QueryResult`.

## Configuration Management

All configuration is loaded from environment variables via `config.LoadFromEnv()`. Required variables: `DB_NAME`, `DB_USER`. The `DB_TYPE` must be "mysql" or "postgres".

Connection pooling is critical for performance - defaults are `DB_MAX_CONNS=10` and `DB_MAX_IDLE_CONNS=5`.

## Error Handling

- Database errors are wrapped with context using `fmt.Errorf("...: %w", err)`
- Tool handlers return errors that are automatically formatted as MCP error responses
- Logs go to stderr (stdout is reserved for MCP protocol)
- Log format: `[LEVEL] Message` (e.g., `[ERROR]`, `[INFO]`, `[DEBUG]`)

## Testing Strategy

When writing tests:
- Mock the `database.Adapter` interface for unit tests
- Test SQL validation thoroughly with edge cases
- Test both valid and malicious SQL patterns
- Use table-driven tests for multiple scenarios
- Test context cancellation and timeouts

## Logging

All logs write to `stderr` (stdout is used for MCP JSON-RPC protocol). Log levels:
- `[FATAL]` - Unrecoverable errors that stop the server
- `[ERROR]` - Operation failures
- `[INFO]` - Important events (startup, shutdown, connections)
- `[DEBUG]` - Protocol-level debugging (requests, responses)
