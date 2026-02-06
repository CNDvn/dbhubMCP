# Changelog

All notable changes to the DBHub MCP Server will be documented in this file.

## [Unreleased]

### Added
- HTTP transport support alongside existing STDIO transport
- Transport abstraction layer with `MessageTransport` interface
- HTTP endpoints:
  - `POST /mcp` - Main JSON-RPC endpoint
  - `GET /health` - Health check endpoint
- Optional API key authentication for HTTP mode via `X-API-Key` header
- Configurable CORS support with wildcard or specific origins
- Concurrent HTTP request handling with buffered channels
- HTTP-specific configuration environment variables:
  - `TRANSPORT_TYPE` - Select "stdio" or "http" mode
  - `HTTP_ADDR` - Configure server listen address
  - `HTTP_CORS_ORIGINS` - Configure CORS origins
  - `HTTP_API_KEY` - Optional API key authentication
- Comprehensive test suite for HTTP transport:
  - Lifecycle management tests
  - Authentication tests
  - CORS header tests
  - Concurrent request handling tests
  - Error handling tests
- HTTP transport documentation in `README_HTTP.md`
- Example HTTP configuration file `.env.http.example`

### Changed
- Refactored transport layer to use interface-based design
- Renamed `Transport` to `StdioTransport` for clarity
- Updated `NewServer()` to accept injected transport
- Modified server startup to call `transport.Start()` and `transport.Close()`
- Updated main README with HTTP transport information
- Enhanced architecture diagram to show dual transport support

### Technical Details
- Zero breaking changes - STDIO mode remains default
- All existing functionality preserved
- No new external dependencies (uses Go standard library)
- Thread-safe HTTP request/response routing
- Request timeout protection (60 seconds per HTTP request)
- Server-level timeouts (30s read/write)
- Graceful shutdown support for HTTP server

### Files Added
- `internal/mcp/transport_interface.go` - Transport abstraction
- `internal/mcp/transport_stdio.go` - STDIO transport (refactored)
- `internal/mcp/transport_http.go` - HTTP transport implementation
- `internal/mcp/transport_http_test.go` - HTTP transport tests
- `README_HTTP.md` - HTTP mode documentation
- `.env.http.example` - HTTP configuration example
- `test_http.sh` - HTTP testing script
- `CHANGELOG.md` - This file

### Files Modified
- `internal/mcp/server.go` - Updated to use transport interface
- `internal/config/config.go` - Added HTTP configuration fields
- `cmd/server/main.go` - Added transport factory and selection
- `README.md` - Updated with HTTP transport information
- `CLAUDE.md` - Updated with HTTP implementation details

### Files Removed
- `internal/mcp/transport.go` - Renamed to `transport_stdio.go`

## [1.0.0] - Previous Release

### Features
- MCP Protocol compliance with STDIO transport
- MySQL and PostgreSQL support
- Read-only query enforcement
- SQL injection prevention
- Connection pooling
- Query timeouts and row limits
- Four MCP tools: list_tables, describe_table, execute_readonly_query, explain_query

---

## Migration Guide

### From STDIO-only to Dual Transport

**No migration needed for existing deployments!** The server continues to use STDIO by default.

**To enable HTTP mode:**

1. Update your `.env` file or environment variables:
   ```bash
   export TRANSPORT_TYPE=http
   export HTTP_ADDR=:8080
   export HTTP_CORS_ORIGINS=*
   export HTTP_API_KEY=your-secret-key
   ```

2. Restart the server:
   ```bash
   ./dbhub-mcp-server
   ```

3. Test the HTTP endpoint:
   ```bash
   curl http://localhost:8080/health
   ```

**To revert to STDIO mode:**

Simply remove the `TRANSPORT_TYPE` environment variable or set it to "stdio".

---

## Security Considerations

When using HTTP mode in production:

1. **Always set an API key:**
   ```bash
   HTTP_API_KEY=$(openssl rand -hex 32)
   ```

2. **Restrict CORS origins:**
   ```bash
   HTTP_CORS_ORIGINS=https://app.example.com,https://admin.example.com
   ```

3. **Use a reverse proxy** (nginx/Caddy) for:
   - TLS/HTTPS termination
   - Rate limiting
   - Access logging
   - IP filtering

4. **Firewall rules:**
   - Restrict port access to known IPs
   - Use VPN or private networks when possible

5. **Monitor logs:**
   - Watch for authentication failures
   - Track unusual query patterns
   - Alert on high error rates

---

## Future Enhancements

Planned features for future releases:

- WebSocket transport for real-time bidirectional communication
- Built-in TLS/HTTPS support with certificate configuration
- Rate limiting per IP address or API key
- JWT token authentication
- Prometheus metrics endpoint for monitoring
- Structured request/response logging
- gRPC transport option
- Multi-tenant support with per-client database configurations

---

## Contributors

- Implementation based on the MCP protocol specification
- HTTP transport design follows industry best practices
- Security features aligned with OWASP guidelines
