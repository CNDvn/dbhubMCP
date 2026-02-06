# DBHub MCP Server

A Model Context Protocol (MCP) server written in Go that provides safe, read-only database operations for MySQL and PostgreSQL.

## Features

- ✅ **MCP Protocol Compliance**: Full implementation of the MCP specification
- ✅ **Dual Transport Support**: STDIO (for Claude Desktop) and HTTP (for web/network clients)
- ✅ **Multi-Database Support**: MySQL and PostgreSQL via unified interface
- ✅ **Read-Only Enforcement**: Multi-layer security to prevent write operations
- ✅ **Connection Pooling**: Efficient connection management
- ✅ **Query Validation**: SQL injection prevention and read-only query enforcement
- ✅ **HTTP Security**: Optional API key authentication and CORS support
- ✅ **Extensible Architecture**: Easy to add support for more databases and transports

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

#### Database Configuration
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

#### Transport Configuration (Optional)
| Variable | Description | Default |
|----------|-------------|---------|
| `TRANSPORT_TYPE` | Transport mode: "stdio" or "http" | stdio |
| `HTTP_ADDR` | HTTP server address (HTTP mode only) | :8080 |
| `HTTP_CORS_ORIGINS` | Comma-separated CORS origins (HTTP mode) | * |
| `HTTP_API_KEY` | Optional API key for authentication | (none) |

## Usage

### Running the Server

**STDIO Mode (Default - for Claude Desktop):**
```bash
# Load environment variables and run
export $(cat .env | xargs) && ./dbhub-mcp-server
```

**HTTP Mode (for web/network clients):**
```bash
export TRANSPORT_TYPE=http
export HTTP_ADDR=:8080
export HTTP_CORS_ORIGINS=*
export HTTP_API_KEY=your-secret-key
export $(cat .env | xargs) && ./dbhub-mcp-server
```

For complete HTTP mode documentation, see [README_HTTP.md](README_HTTP.md).

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
│  MCP Clients                    │
│  ├─ Claude Desktop (STDIO)      │
│  ├─ Web Browser (HTTP)          │
│  └─ Custom Client (HTTP)        │
└────────────┬────────────────────┘
             │
     ┌───────┴───────┐
     ↓               ↓
┌─────────┐    ┌──────────┐
│ STDIO   │    │ HTTP     │
│ (stdin/ │    │ (POST    │
│ stdout) │    │ /mcp)    │
└────┬────┘    └────┬─────┘
     │              │
     └──────┬───────┘
            ↓
┌────────────────────────────┐
│ MessageTransport Interface │
└────────────┬───────────────┘
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
│       └── main.go                  # Entry point
├── internal/
│   ├── mcp/
│   │   ├── server.go                # MCP server implementation
│   │   ├── protocol.go              # MCP protocol types
│   │   ├── transport_interface.go   # Transport abstraction
│   │   ├── transport_stdio.go       # STDIO transport
│   │   ├── transport_http.go        # HTTP transport
│   │   ├── transport_http_test.go   # HTTP transport tests
│   │   └── handlers.go              # Tool handlers
│   ├── database/
│   │   ├── adapter.go               # Database interface
│   │   ├── mysql.go                 # MySQL implementation
│   │   └── postgres.go              # PostgreSQL implementation
│   ├── security/
│   │   └── validator.go             # SQL validation
│   └── config/
│       └── config.go                # Configuration
├── go.mod
├── README.md
└── README_HTTP.md                   # HTTP transport documentation
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
