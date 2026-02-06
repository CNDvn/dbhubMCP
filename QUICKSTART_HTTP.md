# Quick Start Guide - HTTP Mode

## Prerequisites
- Built `dbhub-mcp-server.exe` in the project directory
- MySQL or PostgreSQL database running
- curl (for testing)

## Step 1: Start the Server

### Option A: Using Batch File (Windows)
```bash
# Edit start_http.bat to configure your database
start_http.bat
```

### Option B: Manual Start
```bash
set TRANSPORT_TYPE=http
set HTTP_ADDR=:8080
set HTTP_CORS_ORIGINS=*
set HTTP_API_KEY=test-secret-key

set DB_TYPE=mysql
set DB_NAME=test
set DB_USER=root
set DB_PASSWORD=123456
set DB_HOST=localhost
set DB_PORT=3309

dbhub-mcp-server.exe
```

You should see:
```
[INFO] Starting MCP Server for mysql database
[INFO] MCP Server starting with http transport...
[INFO] Connected to mysql database
[INFO] Registered 4 tools
[INFO] HTTP server will listen on :8080
[INFO] HTTP server listening on :8080
[INFO] Server ready
```

## Step 2: Test the Server

### Quick Test (Health Check)
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

### Run All Tests
```bash
test_http_simple.bat
```

## Step 3: Use the MCP Server

### List Available Tools
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}"
```

### List Database Tables
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/call\",\"params\":{\"name\":\"list_tables\",\"arguments\":{}}}"
```

### Describe a Table
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"describe_table\",\"arguments\":{\"table_name\":\"users\"}}}"
```

### Execute a Query
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"execute_readonly_query\",\"arguments\":{\"query\":\"SELECT * FROM users LIMIT 5\"}}}"
```

## Step 4: Connect from Your Application

### JavaScript/TypeScript Example

```javascript
// mcp-client.js
const MCP_URL = 'http://localhost:8080/mcp';
const API_KEY = 'test-secret-key';

async function callMCPTool(toolName, args) {
  const response = await fetch(MCP_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: Date.now(),
      method: 'tools/call',
      params: {
        name: toolName,
        arguments: args,
      },
    }),
  });

  const data = await response.json();
  return data.result;
}

// Usage examples
async function main() {
  // List all tables
  const tables = await callMCPTool('list_tables', {});
  console.log('Tables:', tables);

  // Describe a table
  const schema = await callMCPTool('describe_table', {
    table_name: 'users'
  });
  console.log('Schema:', schema);

  // Execute a query
  const results = await callMCPTool('execute_readonly_query', {
    query: 'SELECT * FROM users LIMIT 10'
  });
  console.log('Results:', results);
}

main();
```

### Python Example

```python
# mcp_client.py
import requests
import json

MCP_URL = 'http://localhost:8080/mcp'
API_KEY = 'test-secret-key'

def call_mcp_tool(tool_name, args):
    response = requests.post(
        MCP_URL,
        headers={
            'Content-Type': 'application/json',
            'X-API-Key': API_KEY,
        },
        json={
            'jsonrpc': '2.0',
            'id': 1,
            'method': 'tools/call',
            'params': {
                'name': tool_name,
                'arguments': args,
            },
        },
    )
    return response.json()['result']

# Usage examples
if __name__ == '__main__':
    # List all tables
    tables = call_mcp_tool('list_tables', {})
    print('Tables:', json.dumps(tables, indent=2))

    # Describe a table
    schema = call_mcp_tool('describe_table', {'table_name': 'users'})
    print('Schema:', json.dumps(schema, indent=2))

    # Execute a query
    results = call_mcp_tool('execute_readonly_query', {
        'query': 'SELECT * FROM users LIMIT 10'
    })
    print('Results:', json.dumps(results, indent=2))
```

## Available Tools

1. **list_tables** - Lists all tables in the database
   - Arguments: None

2. **describe_table** - Describes a table schema
   - Arguments: `table_name` (string)

3. **execute_readonly_query** - Executes a read-only SQL query
   - Arguments: `query` (string)

4. **explain_query** - Returns query execution plan
   - Arguments: `query` (string)

## Security Notes

### For Development
```bash
HTTP_API_KEY=test-secret-key
HTTP_CORS_ORIGINS=*
```

### For Production
```bash
# Use a strong random key
HTTP_API_KEY=$(openssl rand -hex 32)

# Restrict CORS to your domain
HTTP_CORS_ORIGINS=https://yourdomain.com

# Use reverse proxy (nginx/Caddy) for:
# - TLS/HTTPS
# - Rate limiting
# - IP filtering
```

## Troubleshooting

### Server won't start
- Check if port 8080 is already in use
- Try different port: `HTTP_ADDR=:8081`
- Check database connection settings

### 401 Unauthorized
- Verify API key matches in request header
- Check `X-API-Key` header is set correctly

### CORS errors
- Add your origin to `HTTP_CORS_ORIGINS`
- Use comma-separated list for multiple origins

### Connection refused
- Ensure server is running
- Check firewall settings
- Verify correct URL and port

## Next Steps

- See [README_HTTP.md](README_HTTP.md) for complete documentation
- Configure reverse proxy for production (nginx/Caddy)
- Set up monitoring and logging
- Configure database read-only user
