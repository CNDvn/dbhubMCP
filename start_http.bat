@echo off
echo Starting DBHub MCP Server in HTTP mode...
echo.

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
