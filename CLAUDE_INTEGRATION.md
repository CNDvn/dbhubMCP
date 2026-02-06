# Adding DBHub MCP Server to Claude

## Important: STDIO vs HTTP

- **Claude Desktop**: Uses STDIO transport (stdin/stdout) - This is the DEFAULT mode
- **HTTP Transport**: For web apps and custom applications - NOT for Claude Desktop

## ✅ Option 1: Claude Desktop Integration (Recommended)

### Step 1: Locate Claude Desktop Config

The config file location depends on your OS:

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

**macOS:**
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Linux:**
```
~/.config/Claude/claude_desktop_config.json
```

### Step 2: Add MCP Server Configuration

Open the config file and add your server:

```json
{
  "mcpServers": {
    "dbhub": {
      "command": "C:\\Users\\hieu.banh\\learn\\dbhubMCP\\dbhub-mcp-server.exe",
      "env": {
        "DB_TYPE": "mysql",
        "DB_HOST": "localhost",
        "DB_PORT": "3309",
        "DB_NAME": "test",
        "DB_USER": "root",
        "DB_PASSWORD": "123456"
      }
    }
  }
}
```

**Important Notes:**
- Use double backslashes `\\` in Windows paths
- Use absolute path to the executable
- Do NOT set `TRANSPORT_TYPE` (defaults to STDIO)
- The server will communicate via stdin/stdout automatically

### Step 3: Restart Claude Desktop

Close and reopen Claude Desktop to load the new MCP server.

### Step 4: Verify Connection

In Claude Desktop, you should see:
- MCP icon or indicator showing the server is connected
- You can now ask Claude to query your database

**Example prompts:**
- "List all tables in the database"
- "Show me the schema for the users table"
- "Query the database for all active users"

---

## 🌐 Option 2: HTTP Mode for Custom Applications

If you want to integrate with your own web app or custom client (NOT Claude Desktop):

### Step 1: Start Server in HTTP Mode

```bash
set TRANSPORT_TYPE=http
set HTTP_ADDR=:8080
set HTTP_API_KEY=your-secret-key
set DB_TYPE=mysql
set DB_NAME=test
set DB_USER=root
set DB_PASSWORD=123456

dbhub-mcp-server.exe
```

### Step 2: Connect Your Application

Use HTTP requests to interact with the MCP server:

```javascript
// Example: JavaScript client
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

See `README_HTTP.md` for complete HTTP documentation.

---

## Troubleshooting Claude Desktop Integration

### Server Not Showing Up

1. **Check config file syntax:**
   - Valid JSON (no trailing commas)
   - Correct path to executable
   - Double backslashes in Windows paths

2. **Verify executable works:**
   ```bash
   # Test in command line
   dbhub-mcp-server.exe
   ```
   Should start and wait for input.

3. **Check Claude Desktop logs:**
   - Look for MCP-related errors
   - Verify server startup messages

### Permission Issues

If you get permission errors:
- Ensure the executable has execute permissions
- Check database credentials are correct
- Verify network access to database

### Connection Timeouts

If Claude Desktop can't connect:
- Increase connection timeout in config
- Check database is running
- Verify firewall settings

---

## Example Configurations

### MySQL (Local)
```json
{
  "mcpServers": {
    "mysql-local": {
      "command": "C:\\path\\to\\dbhub-mcp-server.exe",
      "env": {
        "DB_TYPE": "mysql",
        "DB_HOST": "localhost",
        "DB_PORT": "3306",
        "DB_NAME": "myapp",
        "DB_USER": "readonly_user",
        "DB_PASSWORD": "secure_password"
      }
    }
  }
}
```

### PostgreSQL (Remote)
```json
{
  "mcpServers": {
    "postgres-prod": {
      "command": "C:\\path\\to\\dbhub-mcp-server.exe",
      "env": {
        "DB_TYPE": "postgres",
        "DB_HOST": "db.example.com",
        "DB_PORT": "5432",
        "DB_NAME": "production",
        "DB_USER": "readonly_user",
        "DB_PASSWORD": "secure_password",
        "QUERY_TIMEOUT_SEC": "60",
        "MAX_ROWS": "500"
      }
    }
  }
}
```

### Multiple Databases
```json
{
  "mcpServers": {
    "mysql-dev": {
      "command": "C:\\path\\to\\dbhub-mcp-server.exe",
      "env": {
        "DB_TYPE": "mysql",
        "DB_NAME": "dev_database",
        "DB_USER": "dev_user",
        "DB_PASSWORD": "dev_pass"
      }
    },
    "postgres-analytics": {
      "command": "C:\\path\\to\\dbhub-mcp-server.exe",
      "env": {
        "DB_TYPE": "postgres",
        "DB_NAME": "analytics",
        "DB_USER": "analytics_user",
        "DB_PASSWORD": "analytics_pass"
      }
    }
  }
}
```

---

## Security Best Practices

### For Claude Desktop (STDIO Mode)

1. **Use Read-Only Database User:**
   ```sql
   -- MySQL
   CREATE USER 'readonly'@'%' IDENTIFIED BY 'secure_password';
   GRANT SELECT ON mydb.* TO 'readonly'@'%';
   FLUSH PRIVILEGES;

   -- PostgreSQL
   CREATE USER readonly WITH PASSWORD 'secure_password';
   GRANT CONNECT ON DATABASE mydb TO readonly;
   GRANT USAGE ON SCHEMA public TO readonly;
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
   ```

2. **Limit Query Results:**
   ```json
   "env": {
     "MAX_ROWS": "1000",
     "QUERY_TIMEOUT_SEC": "30"
   }
   ```

3. **Connection Limits:**
   ```json
   "env": {
     "DB_MAX_CONNS": "5",
     "DB_MAX_IDLE_CONNS": "2"
   }
   ```

### For HTTP Mode

1. **Always Use API Key:**
   ```bash
   HTTP_API_KEY=$(openssl rand -hex 32)
   ```

2. **Restrict CORS:**
   ```bash
   HTTP_CORS_ORIGINS=https://yourdomain.com
   ```

3. **Use Reverse Proxy:**
   - nginx or Caddy for TLS
   - Rate limiting
   - IP filtering

---

## Next Steps

### After Adding to Claude Desktop:

1. **Test Basic Queries:**
   - "List all tables"
   - "Show schema for users table"

2. **Try Complex Operations:**
   - "Find all users created in the last week"
   - "Show me the top 10 products by sales"

3. **Use in Workflows:**
   - Data analysis
   - Report generation
   - Schema exploration

### For HTTP Integration:

1. **Build Custom UI:** See `demo.html` for example
2. **Create Client Library:** Wrap API calls
3. **Add to Your App:** Integrate database queries
4. **Monitor Usage:** Track queries and performance

---

## Summary

| Use Case | Transport Mode | Config Location |
|----------|---------------|-----------------|
| Claude Desktop | STDIO (default) | `claude_desktop_config.json` |
| Web Applications | HTTP | Environment variables |
| Custom Clients | HTTP | Environment variables |

**For Claude Desktop:** Use STDIO mode (default) - just add to config file
**For Your Apps:** Use HTTP mode - start server with `TRANSPORT_TYPE=http`

See also:
- `claude_desktop_example.json` - Example config for Claude Desktop
- `README_HTTP.md` - Complete HTTP mode documentation
- `QUICKSTART_HTTP.md` - Quick start guide for HTTP mode
