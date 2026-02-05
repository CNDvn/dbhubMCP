# Quick Start Guide

Get the DBHub MCP Server up and running in 5 minutes.

## Prerequisites

- Go 1.21 or later
- MySQL or PostgreSQL database
- Read-only database user (recommended)

## Step 1: Setup Database

### For MySQL:

```sql
-- Create a test database
CREATE DATABASE testdb;

-- Create a read-only user
CREATE USER 'readonly'@'%' IDENTIFIED BY 'password123';
GRANT SELECT ON testdb.* TO 'readonly'@'%';
FLUSH PRIVILEGES;

-- Create a sample table
USE testdb;
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO users (username, email) VALUES
    ('alice', 'alice@example.com'),
    ('bob', 'bob@example.com'),
    ('charlie', 'charlie@example.com');
```

### For PostgreSQL:

```sql
-- Create a test database
CREATE DATABASE testdb;

-- Connect to testdb
\c testdb

-- Create a read-only user
CREATE USER readonly WITH PASSWORD 'password123';
GRANT CONNECT ON DATABASE testdb TO readonly;
GRANT USAGE ON SCHEMA public TO readonly;

-- Create a sample table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grant SELECT permission
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly;

-- Insert sample data
INSERT INTO users (username, email) VALUES
    ('alice', 'alice@example.com'),
    ('bob', 'bob@example.com'),
    ('charlie', 'charlie@example.com');
```

## Step 2: Configure Environment

```bash
# Copy example configuration
cp .env.example .env

# Edit .env with your database credentials
nano .env
```

Example `.env` for MySQL:
```bash
DB_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=testdb
DB_USER=readonly
DB_PASSWORD=password123
```

## Step 3: Build and Run

```bash
# Install dependencies
go mod download

# Build the server
go build -o dbhub-mcp-server ./cmd/server

# Run the server
export $(cat .env | xargs) && ./dbhub-mcp-server
```

Or use the Makefile:
```bash
make build
make run
```

You should see:
```
╔═══════════════════════════════════════╗
║     DBHub MCP Server v1.0.0           ║
║  Database Operations via MCP Protocol ║
╚═══════════════════════════════════════╝

[INFO] Starting MCP Server for mysql database
[INFO] Connected to mysql database
[INFO] Registered 4 tools
[INFO] Server ready, listening on STDIO...
```

## Step 4: Test the Server

The server communicates via JSON-RPC over STDIO. You can test it manually:

### Test 1: Initialize

Send this JSON to stdin:
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}
```

Expected response:
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{"listChanged":false}},"serverInfo":{"name":"dbhub-mcp-server","version":"1.0.0"}}}
```

### Test 2: List Tools

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

### Test 3: List Tables

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_tables","arguments":{}}}
```

### Test 4: Execute Query

```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"execute_readonly_query","arguments":{"query":"SELECT * FROM users LIMIT 5"}}}
```

### Test 5: Try Write Operation (Should Fail)

```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"execute_readonly_query","arguments":{"query":"DELETE FROM users WHERE id = 1"}}}
```

Expected error:
```json
{"jsonrpc":"2.0","id":5,"result":{"content":[{"type":"text","text":"Error: query validation failed: write operation detected: DELETE is not allowed"}],"isError":true}}
```

## Step 5: Use with Claude Desktop

1. Find your Claude Desktop config file:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

2. Add the MCP server:

```json
{
  "mcpServers": {
    "dbhub": {
      "command": "/path/to/dbhub-mcp-server",
      "env": {
        "DB_TYPE": "mysql",
        "DB_HOST": "localhost",
        "DB_PORT": "3306",
        "DB_NAME": "testdb",
        "DB_USER": "readonly",
        "DB_PASSWORD": "password123"
      }
    }
  }
}
```

3. Restart Claude Desktop

4. Ask Claude:
   - "What tables are in my database?"
   - "Show me the schema of the users table"
   - "Query the users table and show me all records"

## Troubleshooting

### Connection Refused

**Problem:** Can't connect to database.

**Solutions:**
- Check database is running: `mysql -u readonly -p` or `psql -U readonly -d testdb`
- Verify host/port in `.env`
- Check firewall rules

### Permission Denied

**Problem:** Database user lacks permissions.

**Solutions:**
- Verify user has SELECT grants: `SHOW GRANTS FOR 'readonly'@'%';` (MySQL)
- Re-run the GRANT commands from Step 1

### Invalid Query

**Problem:** Query is blocked as unsafe.

**Solutions:**
- Ensure query starts with SELECT, EXPLAIN, DESCRIBE, or SHOW
- Remove write operations (INSERT, UPDATE, DELETE)
- Remove SQL comments (`--` or `/* */`)

### Timeout Error

**Problem:** Query takes too long.

**Solutions:**
- Increase `QUERY_TIMEOUT_SEC` in `.env`
- Optimize query with indexes
- Add LIMIT clause to reduce result size

## Next Steps

- Read [DESIGN.md](DESIGN.md) to understand the architecture
- Read [CLAUDE.md](CLAUDE.md) for development guidance
- Check [examples/](examples/) for more JSON-RPC examples
- Explore the codebase in [internal/](internal/)

## Production Checklist

Before deploying to production:

- [ ] Use a dedicated read-only database user
- [ ] Set strong password in secure environment variable
- [ ] Configure appropriate `QUERY_TIMEOUT_SEC` (default: 30s)
- [ ] Set `MAX_ROWS` based on your use case (default: 1000)
- [ ] Tune connection pool: `DB_MAX_CONNS`, `DB_MAX_IDLE_CONNS`
- [ ] Test with your actual database schema
- [ ] Monitor server logs for errors
- [ ] Set up graceful shutdown handling
- [ ] Consider using SSL/TLS for database connections

## Getting Help

- Check the [README.md](README.md) for full documentation
- Review [examples/](examples/) for MCP request/response samples
- Open an issue on GitHub
