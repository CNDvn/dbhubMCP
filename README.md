# DBHub MCP Server

A Model Context Protocol (MCP) server written in Go that provides safe, read-only database operations for MySQL and PostgreSQL.

## Features

- ✅ **MCP Protocol Compliance**: Full implementation of the MCP specification with STDIO transport
- ✅ **Multi-Database Support**: MySQL and PostgreSQL via unified interface
- ✅ **Read-Only Enforcement**: Multi-layer security to prevent write operations
- ✅ **Connection Pooling**: Efficient connection management
- ✅ **Query Validation**: SQL injection prevention and read-only query enforcement
- ✅ **Extensible Architecture**: Easy to add support for more databases

## MCP Tools

The server exposes the following MCP tools:

1. **list_tables** - Lists all tables in the database
2. **describe_table** - Returns schema information for a specific table
3. **execute_readonly_query** - Executes SELECT queries (write operations blocked)
4. **explain_query** - Returns query execution plans without executing

## Installation

### Prerequisites

- Go 1.21 or later
- MySQL or PostgreSQL database
- Read-only database user (recommended)

### Build

```bash
# Clone the repository
git clone <repository-url>
cd dbhubMCP

# Install dependencies
go mod download

# Build the server
go build -o dbhub-mcp-server ./cmd/server
```

## Configuration

Configure the server using environment variables:

```bash
# Copy example configuration
cp .env.example .env

# Edit configuration
nano .env
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_TYPE` | Database type: "mysql" or "postgres" | mysql |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 3306 |
| `DB_NAME` | Database name | (required) |
| `DB_USER` | Database username | (required) |
| `DB_PASSWORD` | Database password | (required) |
| `DB_MAX_CONNS` | Maximum open connections | 10 |
| `DB_MAX_IDLE_CONNS` | Maximum idle connections | 5 |
| `DB_CONN_TIMEOUT_SEC` | Connection timeout in seconds | 10 |
| `QUERY_TIMEOUT_SEC` | Query execution timeout | 30 |
| `MAX_ROWS` | Maximum rows to return | 1000 |
| `LOG_LEVEL` | Logging level | info |

## Usage

### Running the Server

```bash
# Load environment variables and run
export $(cat .env | xargs) && ./dbhub-mcp-server
```

### Using with Claude Desktop

Add to your Claude Desktop MCP configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "dbhub": {
      "command": "/path/to/dbhub-mcp-server",
      "env": {
        "DB_TYPE": "mysql",
        "DB_HOST": "localhost",
        "DB_PORT": "3306",
        "DB_NAME": "mydb",
        "DB_USER": "readonly_user",
        "DB_PASSWORD": "password"
      }
    }
  }
}
```

## Security

### Multi-Layer Read-Only Enforcement

1. **SQL Validation Layer**
   - Blocks INSERT, UPDATE, DELETE, DROP, ALTER, TRUNCATE
   - Validates queries start with SELECT, EXPLAIN, DESCRIBE, or SHOW
   - Detects SQL injection patterns

2. **Database Permission Layer**
   - Use a read-only database user
   - Grant only SELECT privileges

3. **Query Limits**
   - Configurable timeout per query
   - Maximum row limits to prevent memory issues

### Creating Read-Only Users

**MySQL:**
```sql
CREATE USER 'readonly_user'@'%' IDENTIFIED BY 'secure_password';
GRANT SELECT ON mydb.* TO 'readonly_user'@'%';
FLUSH PRIVILEGES;
```

**PostgreSQL:**
```sql
CREATE USER readonly_user WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE mydb TO readonly_user;
GRANT USAGE ON SCHEMA public TO readonly_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly_user;
```

## Architecture

```
┌─────────────────────────────────┐
│   MCP Client (Claude/LLM)       │
└────────────┬────────────────────┘
             │ STDIO (JSON-RPC)
┌────────────▼────────────────────┐
│   Transport Layer               │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│   MCP Server                    │
│   - Protocol handling           │
│   - Tool routing                │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│   Security/Validation Layer     │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│   Database Adapter Interface    │
└──────┬──────────────┬───────────┘
       │              │
   ┌───▼────┐    ┌───▼─────┐
   │ MySQL  │    │ Postgres│
   └────────┘    └─────────┘
```

## Development

### Project Structure

```
dbhubMCP/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── mcp/
│   │   ├── server.go            # MCP server implementation
│   │   ├── protocol.go          # MCP protocol types
│   │   ├── transport.go         # STDIO transport
│   │   └── handlers.go          # Tool handlers
│   ├── database/
│   │   ├── adapter.go           # Database interface
│   │   ├── mysql.go             # MySQL implementation
│   │   └── postgres.go          # PostgreSQL implementation
│   ├── security/
│   │   └── validator.go         # SQL validation
│   └── config/
│       └── config.go            # Configuration
├── go.mod
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Adding a New Database

1. Implement the `database.Adapter` interface in `internal/database/`
2. Add database-specific driver import
3. Register in `main.go` switch statement

## Example MCP Interactions

See `examples/` directory for complete JSON-RPC request/response examples:
- `initialize.json` - Server initialization
- `list_tools.json` - Tool discovery
- `call_list_tables.json` - List tables example
- `call_execute_query.json` - Execute query example

## License

MIT License

## Contributing

Contributions welcome! Please ensure:
- Code follows Go conventions
- Security best practices maintained
- Tests included for new features
