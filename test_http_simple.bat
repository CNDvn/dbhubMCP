@echo off
echo ========================================
echo Testing DBHub MCP Server - HTTP Mode
echo ========================================
echo.

set API_KEY=test-secret-key

echo Test 1: Health Check
curl -s http://localhost:8080/health
echo.
echo.

echo Test 2: List Available Tools
curl -s -X POST http://localhost:8080/mcp ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: %API_KEY%" ^
  -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}"
echo.
echo.

echo Test 3: Ping Database
curl -s -X POST http://localhost:8080/mcp ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: %API_KEY%" ^
  -d "{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"ping\"}"
echo.
echo.

echo Test 4: List Tables
curl -s -X POST http://localhost:8080/mcp ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: %API_KEY%" ^
  -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"list_tables\",\"arguments\":{}}}"
echo.
echo.

echo ========================================
echo All tests completed!
echo ========================================
