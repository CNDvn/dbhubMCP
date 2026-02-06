# HTTP Transport for DBHub MCP Server

This document describes the HTTP transport feature added to the DBHub MCP Server.

## Overview

The DBHub MCP Server now supports two transport modes:
- **STDIO mode** (default): For Claude Desktop integration
- **HTTP mode** (new): For network-based clients and web applications

## Quick Start

### STDIO Mode (Default)
No changes needed. The server works exactly as before:

```bash
export DB_TYPE=mysql
export DB_NAME=mydb
export DB_USER=root
export DB_PASSWORD=secret
./dbhub-mcp-server
```

### HTTP Mode

Start the server in HTTP mode:

```bash
export TRANSPORT_TYPE=http
export HTTP_ADDR=:8080
export HTTP_CORS_ORIGINS=*
export HTTP_API_KEY=your-secret-key

export DB_TYPE=mysql
export DB_NAME=mydb
export DB_USER=root
export DB_PASSWORD=secret

./dbhub-mcp-server
```

## Configuration

### Environment Variables

#### Transport Configuration
- `TRANSPORT_TYPE` - Transport mode: "stdio" (default) or "http"
- `HTTP_ADDR` - Server address (default: ":8080")
- `HTTP_CORS_ORIGINS` - Comma-separated CORS origins (default: "*")
- `HTTP_API_KEY` - Optional API key for authentication (default: none)

#### Database Configuration (same as before)
- `DB_TYPE` - Database type: "mysql" or "postgres"
- `DB_HOST` - Database host (default: "localhost")
- `DB_PORT` - Database port (default: 3306 for MySQL, 5432 for PostgreSQL)
- `DB_NAME` - Database name (required)
- `DB_USER` - Database user (required)
- `DB_PASSWORD` - Database password
- `DB_MAX_CONNS` - Maximum connections (default: 10)
- `DB_MAX_IDLE_CONNS` - Maximum idle connections (default: 5)
- `DB_CONN_TIMEOUT_SEC` - Connection timeout in seconds (default: 10)
- `QUERY_TIMEOUT_SEC` - Query timeout in seconds (default: 30)
- `MAX_ROWS` - Maximum rows per query (default: 1000)

## HTTP API

### Endpoints

#### POST /mcp
Main JSON-RPC endpoint for MCP protocol messages.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "list_tables",
        "description": "Lists all tables in the connected database",
        "inputSchema": {...}
      }
    ]
  }
}
```

**Authentication:**
If `HTTP_API_KEY` is set, include the `X-API-Key` header:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

#### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

### CORS Support

The server supports CORS for cross-origin requests. By default, all origins are allowed (`*`).

**Restrict to specific origins:**
```bash
export HTTP_CORS_ORIGINS=https://app.example.com,https://admin.example.com
```

**Preflight requests** (OPTIONS) are automatically handled.

## Security

### API Key Authentication

Set `HTTP_API_KEY` to require authentication:

```bash
export HTTP_API_KEY=my-secret-key-12345
```

All requests to `/mcp` must include the `X-API-Key` header:

```bash
curl -H "X-API-Key: my-secret-key-12345" http://localhost:8080/mcp
```

**Note:** If `HTTP_API_KEY` is not set, no authentication is required (useful for development).

### CORS Configuration

**Development (allow all):**
```bash
export HTTP_CORS_ORIGINS=*
```

**Production (restrict origins):**
```bash
export HTTP_CORS_ORIGINS=https://app.example.com,https://admin.example.com
```

### Best Practices

1. **Always set an API key in production:**
   ```bash
   export HTTP_API_KEY=$(openssl rand -hex 32)
   ```

2. **Restrict CORS origins:**
   Use specific domains instead of wildcard `*`

3. **Use a reverse proxy:**
   Place nginx or Caddy in front for TLS, rate limiting, and logging

4. **Firewall rules:**
   Restrict port access to known IP addresses

5. **Monitor logs:**
   Watch for suspicious activity in stderr logs

## Example Usage

### JavaScript/TypeScript

```javascript
async function listTables() {
  const response = await fetch('http://localhost:8080/mcp', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your-secret-key',
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/call',
      params: {
        name: 'list_tables',
        arguments: {},
      },
    }),
  });

  const data = await response.json();
  console.log(data.result);
}
```

### Python

```python
import requests

def list_tables():
    response = requests.post(
        'http://localhost:8080/mcp',
        headers={
            'Content-Type': 'application/json',
            'X-API-Key': 'your-secret-key',
        },
        json={
            'jsonrpc': '2.0',
            'id': 1,
            'method': 'tools/call',
            'params': {
                'name': 'list_tables',
                'arguments': {},
            },
        },
    )
    return response.json()
```

### curl

```bash
# List all tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'

# List tables
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "list_tables",
      "arguments": {}
    }
  }'

# Execute query
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-key" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "execute_readonly_query",
      "arguments": {
        "query": "SELECT * FROM users LIMIT 10"
      }
    }
  }'
```

## Testing

### Unit Tests

Run all tests:
```bash
go test -v ./...
```

Run tests with race detector:
```bash
go test -race ./...
```

Run HTTP transport tests only:
```bash
go test -v ./internal/mcp -run TestHTTPTransport
```

### Manual Testing

1. **Start server in HTTP mode:**
   ```bash
   export TRANSPORT_TYPE=http
   export HTTP_ADDR=:8080
   export HTTP_API_KEY=test-key
   ./dbhub-mcp-server
   ```

2. **Test health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Test authentication:**
   ```bash
   # Without API key (should fail)
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'

   # With API key (should succeed)
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -H "X-API-Key: test-key" \
     -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
   ```

4. **Test CORS:**
   ```bash
   curl -I -X OPTIONS http://localhost:8080/mcp \
     -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: POST"
   ```

## Architecture

### Transport Interface

All transports implement the `MessageTransport` interface:

```go
type MessageTransport interface {
    GetType() TransportType
    Start(ctx context.Context) error
    ReadRequest() (*Request, error)
    WriteResponse(resp *Response) error
    Close() error
}
```

### HTTP Transport Implementation

- **Concurrent request handling** - Buffered channel (capacity: 10)
- **Request timeout** - 60 seconds per request
- **Server timeouts** - 30s read/write timeouts
- **Graceful shutdown** - 5 second shutdown timeout
- **CORS support** - Configurable origins with preflight handling
- **Optional authentication** - API key via `X-API-Key` header

### Request Flow

```
HTTP POST /mcp
    ↓
Authentication (if API key configured)
    ↓
Parse JSON-RPC request
    ↓
Queue request in channel
    ↓
Wait for response (60s timeout)
    ↓
Return JSON-RPC response
```

## Backward Compatibility

- **Zero breaking changes** - STDIO mode remains default
- **Existing deployments unaffected** - No config changes needed
- **Claude Desktop integration unchanged** - Uses STDIO
- **All environment variables optional** - Existing `.env` files work

## Future Enhancements

Potential features (not yet implemented):
- WebSocket transport for bidirectional streaming
- TLS/HTTPS support with certificate configuration
- Rate limiting per IP address
- JWT token authentication
- Prometheus metrics endpoint
- Structured request logging
- gRPC transport

## Troubleshooting

### Server won't start in HTTP mode

Check that the port is not already in use:
```bash
netstat -an | grep 8080
```

Try a different port:
```bash
export HTTP_ADDR=:8081
```

### Authentication failures

Verify API key matches:
```bash
echo $HTTP_API_KEY
```

Check request header:
```bash
curl -v -H "X-API-Key: your-key" http://localhost:8080/mcp
```

### CORS errors in browser

Check CORS configuration:
```bash
echo $HTTP_CORS_ORIGINS
```

Add your origin:
```bash
export HTTP_CORS_ORIGINS=http://localhost:3000,https://app.example.com
```

### Request timeouts

Increase query timeout:
```bash
export QUERY_TIMEOUT_SEC=60
```

Check database connection:
```bash
curl -H "X-API-Key: your-key" \
  -X POST http://localhost:8080/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping"}'
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/hieubanhh/dbhubMCP/issues
- Documentation: See `CLAUDE.md` for architecture details
