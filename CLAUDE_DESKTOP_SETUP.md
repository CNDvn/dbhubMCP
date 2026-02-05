# Claude Desktop Setup Instructions

## Your MCP Server is Working! ✓

The connection test succeeded. Now you just need to configure Claude Desktop properly.

## Step-by-Step Setup

### 1. Locate Claude Desktop Config File

Open File Explorer and paste this path in the address bar:
```
%APPDATA%\Claude
```

You should see a file called `claude_desktop_config.json`

### 2. Edit the Config File

Open `claude_desktop_config.json` with a text editor (Notepad, VS Code, etc.)

**If the file is empty or only has `{}`**, replace the entire content with:

```json
{
  "mcpServers": {
    "dbhub": {
      "command": "C:\\Users\\hieu.banh\\learn\\dbhubMCP\\dbhub-mcp-server.exe",
      "args": [],
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

**If the file already has other MCP servers**, add the `dbhub` entry inside the `mcpServers` object:

```json
{
  "mcpServers": {
    "existing-server": {
      ...existing config...
    },
    "dbhub": {
      "command": "C:\\Users\\hieu.banh\\learn\\dbhubMCP\\dbhub-mcp-server.exe",
      "args": [],
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

### 3. Important Notes

**Path Format:**
- Use double backslashes `\\` in Windows paths
- The full path must point to your `dbhub-mcp-server.exe` file
- Your path: `C:\\Users\\hieu.banh\\learn\\dbhubMCP\\dbhub-mcp-server.exe`

**Password Security:**
- The password is visible in the config file
- Make sure the file has appropriate permissions
- Consider using a read-only database user instead of root

### 4. Restart Claude Desktop

After saving the config file:
1. Completely quit Claude Desktop (not just close the window)
2. Start Claude Desktop again
3. The MCP server should now connect successfully

### 5. Verify Connection

Once Claude Desktop restarts, you should see:
- A connection indicator for the MCP server
- The server should show as "connected"

You can then ask Claude:
- "What tables are in my database?"
- "Show me the schema of the account_balance table"
- "Query the test table and show me the data"

## Troubleshooting

### Error: "Command not found" or "Cannot execute"

**Solution:** Check the path is correct
```bash
# Verify the file exists
ls -la "C:\Users\hieu.banh\learn\dbhubMCP\dbhub-mcp-server.exe"
```

### Error: "Failed to connect to database"

**Solution:** Verify credentials in the `env` section match your MySQL setup

### Error: "MCP server crashed"

**Solution:** Check Claude Desktop logs at:
```
%APPDATA%\Claude\logs
```

## Testing Without Claude Desktop

You can test the server manually:

```bash
cd C:\Users\hieu.banh\learn\dbhubMCP

# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./dbhub-mcp-server.exe

# Test listing tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./dbhub-mcp-server.exe

# Test listing tables
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_tables","arguments":{}}}' | ./dbhub-mcp-server.exe
```

## Quick Reference

**Your Database:**
- Type: MySQL
- Host: localhost
- Port: 3309
- Database: test
- Tables: account_balance, test

**Your MCP Server:**
- Location: `C:\Users\hieu.banh\learn\dbhubMCP\dbhub-mcp-server.exe`
- Status: ✓ Working (connection test passed)

**Tools Available:**
1. `list_tables` - List all tables
2. `describe_table` - Show table schema
3. `execute_readonly_query` - Run SELECT queries
4. `explain_query` - Show query execution plan

## Support

If you're still having issues:
1. Check the server logs in Claude Desktop logs folder
2. Run the connection test: `go run test_connection.go`
3. Verify the exe file exists and is executable
4. Make sure MySQL is running on port 3309
