package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Using the defaults from config.go
	host := "localhost"
	port := 3309
	dbName := "test"
	user := "root"
	password := "123456"
	timeout := 10 * time.Second

	fmt.Println("========================================")
	fmt.Println("MySQL Connection Test")
	fmt.Println("========================================")
	fmt.Printf("Host:     %s\n", host)
	fmt.Printf("Port:     %d\n", port)
	fmt.Printf("Database: %s\n", dbName)
	fmt.Printf("User:     %s\n", user)
	fmt.Printf("Password: %s\n", "[hidden]")
	fmt.Println("========================================")
	fmt.Println()

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=%s",
		user, password, host, port, dbName, timeout)

	fmt.Println("1. Testing connection to MySQL server...")

	// Try to connect
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("✗ Failed to open connection: %v\n\n", err)
		fmt.Println("This usually means driver issue or DSN format problem.")
		return
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection with ping
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Println("2. Pinging database...")
	err = db.PingContext(ctx)
	if err != nil {
		fmt.Printf("✗ Failed to ping database: %v\n\n", err)

		fmt.Println("Common causes:")
		fmt.Println("  1. Database 'test' does not exist")
		fmt.Println("  2. Wrong password for user 'root'")
		fmt.Println("  3. User 'root' doesn't have access from 'localhost'")
		fmt.Println("  4. MySQL server not running")
		fmt.Println()
		fmt.Println("Solutions:")
		fmt.Println("  • Try connecting without database name first:")
		fmt.Printf("      mysql -h %s -P %d -u %s -p\n", host, port, user)
		fmt.Println("  • Create the database:")
		fmt.Printf("      mysql -u root -p -e \"CREATE DATABASE test;\"\n")
		fmt.Println("  • Or update config.go to use an existing database")
		fmt.Println()

		// Try connecting without database name
		fmt.Println("3. Trying to connect without database name...")
		dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&timeout=%s",
			user, password, host, port, timeout)
		dbNoDB, err2 := sql.Open("mysql", dsnNoDB)
		if err2 == nil {
			defer dbNoDB.Close()
			ctx2, cancel2 := context.WithTimeout(context.Background(), timeout)
			defer cancel2()

			err2 = dbNoDB.PingContext(ctx2)
			if err2 == nil {
				fmt.Println("✓ Successfully connected to MySQL server!")
				fmt.Println()
				fmt.Println("4. Checking available databases...")

				rows, err3 := dbNoDB.QueryContext(ctx2, "SHOW DATABASES")
				if err3 == nil {
					defer rows.Close()
					fmt.Println("Available databases:")
					for rows.Next() {
						var dbName string
						if err := rows.Scan(&dbName); err == nil {
							fmt.Printf("  • %s\n", dbName)
						}
					}
					fmt.Println()
					fmt.Println("ACTION REQUIRED:")
					fmt.Println("  Either create the 'test' database or update DB_NAME in config.go")
					fmt.Println("  to use one of the databases listed above.")
				}
			} else {
				fmt.Printf("✗ Still failed: %v\n", err2)
				fmt.Println()
				fmt.Println("This indicates a credential issue.")
				fmt.Println("Check your MySQL username and password.")
			}
		}
		return
	}

	fmt.Println("✓ Successfully connected to MySQL!")
	fmt.Println()

	// List tables
	fmt.Println("3. Listing tables in database 'test'...")
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		fmt.Printf("✗ Failed to list tables: %v\n", err)
		return
	}
	defer rows.Close()

	tableCount := 0
	fmt.Println("Tables:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			fmt.Printf("  • %s\n", tableName)
			tableCount++
		}
	}

	if tableCount == 0 {
		fmt.Println("  (no tables found)")
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("✓ Connection test SUCCESSFUL!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("The MCP server should work now.")
	fmt.Println("Run: ./dbhub-mcp-server")
}
