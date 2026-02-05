#!/bin/bash

# Database Connection Debugging Script

echo "=========================================="
echo "Database Connection Diagnostics"
echo "=========================================="
echo ""

# Load environment variables if .env exists
if [ -f .env ]; then
    echo "✓ Found .env file"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠ No .env file found, using defaults from config.go"
fi

# Display current configuration
echo ""
echo "Current Configuration:"
echo "-------------------------------------------"
echo "DB_TYPE:     ${DB_TYPE:-mysql}"
echo "DB_HOST:     ${DB_HOST:-localhost}"
echo "DB_PORT:     ${DB_PORT:-3309}"
echo "DB_NAME:     ${DB_NAME:-test}"
echo "DB_USER:     ${DB_USER:-root}"
echo "DB_PASSWORD: ${DB_PASSWORD:+[SET - hidden]}"
echo "-------------------------------------------"
echo ""

# Check if MySQL is installed
echo "1. Checking MySQL client installation..."
if command -v mysql &> /dev/null; then
    echo "✓ MySQL client is installed"
    mysql --version
else
    echo "✗ MySQL client not found. Install it to test connections manually."
fi
echo ""

# Test MySQL connection manually
echo "2. Testing MySQL connection..."
echo "   Attempting: mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3309} -u ${DB_USER:-root} -p${DB_PASSWORD:-123456}"
echo ""

if command -v mysql &> /dev/null; then
    # Try to connect and show databases
    mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3309} -u ${DB_USER:-root} -p${DB_PASSWORD:-123456} -e "SHOW DATABASES;" 2>&1

    if [ $? -eq 0 ]; then
        echo ""
        echo "✓ MySQL connection successful!"
        echo ""

        # Check if the target database exists
        echo "3. Checking if database '${DB_NAME:-test}' exists..."
        DB_EXISTS=$(mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3309} -u ${DB_USER:-root} -p${DB_PASSWORD:-123456} -e "SHOW DATABASES LIKE '${DB_NAME:-test}';" 2>&1 | grep -c "${DB_NAME:-test}")

        if [ "$DB_EXISTS" -gt 0 ]; then
            echo "✓ Database '${DB_NAME:-test}' exists"
            echo ""
            echo "4. Listing tables in '${DB_NAME:-test}'..."
            mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3309} -u ${DB_USER:-root} -p${DB_PASSWORD:-123456} ${DB_NAME:-test} -e "SHOW TABLES;" 2>&1
        else
            echo "✗ Database '${DB_NAME:-test}' does NOT exist"
            echo ""
            echo "Available databases:"
            mysql -h ${DB_HOST:-localhost} -P ${DB_PORT:-3309} -u ${DB_USER:-root} -p${DB_PASSWORD:-123456} -e "SHOW DATABASES;" 2>&1
        fi
    else
        echo ""
        echo "✗ MySQL connection FAILED"
        echo ""
        echo "Common issues:"
        echo "  1. MySQL server not running"
        echo "  2. Wrong port (you're using 3309, standard is 3306)"
        echo "  3. Wrong credentials"
        echo "  4. Firewall blocking connection"
        echo "  5. MySQL not listening on the specified host/port"
    fi
fi

echo ""
echo "=========================================="
echo "Next Steps:"
echo "=========================================="
echo ""
echo "If connection failed:"
echo "  1. Check MySQL is running: sudo systemctl status mysql (Linux) or check Services (Windows)"
echo "  2. Verify port: netstat -an | grep 3309"
echo "  3. Update .env with correct values"
echo "  4. Create database if needed: mysql -u root -p -e \"CREATE DATABASE test;\""
echo ""
echo "To test the MCP server:"
echo "  export \$(cat .env | xargs) && ./dbhub-mcp-server"
echo ""
