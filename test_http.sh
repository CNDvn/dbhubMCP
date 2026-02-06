#!/bin/bash
# Test script for HTTP transport mode

echo "=== Testing HTTP Transport Mode ==="
echo ""

# Start server in background
echo "Starting server in HTTP mode..."
export TRANSPORT_TYPE=http
export HTTP_ADDR=:8080
export HTTP_CORS_ORIGINS=*
export HTTP_API_KEY=test-secret-key

# Use existing database configuration
./dbhub-mcp-server.exe &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Server started with PID: $SERVER_PID"
echo ""

# Test 1: Health check
echo "Test 1: Health Check"
curl -s http://localhost:8080/health | jq .
echo ""

# Test 2: tools/list without API key (should fail)
echo "Test 2: tools/list without API key (should fail with 401)"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
echo ""

# Test 3: tools/list with correct API key
echo "Test 3: tools/list with correct API key"
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | jq .
echo ""

# Test 4: ping with API key
echo "Test 4: ping with API key"
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-secret-key" \
  -d '{"jsonrpc":"2.0","id":2,"method":"ping"}' | jq .
echo ""

# Test 5: CORS headers
echo "Test 5: CORS preflight request"
curl -s -I -X OPTIONS http://localhost:8080/mcp \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" | grep -i "access-control"
echo ""

# Test 6: Method not allowed
echo "Test 6: GET request to /mcp (should fail with 405)"
curl -s -w "\nHTTP Status: %{http_code}\n" -X GET http://localhost:8080/mcp
echo ""

# Cleanup
echo "Stopping server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo ""
echo "=== All tests completed ==="
