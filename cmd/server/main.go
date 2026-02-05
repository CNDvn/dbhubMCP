package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hieubanhh/dbhubMCP/internal/config"
	"github.com/hieubanhh/dbhubMCP/internal/database"
	"github.com/hieubanhh/dbhubMCP/internal/mcp"
	"github.com/hieubanhh/dbhubMCP/internal/security"
)

func main() {
	// Setup logging to stderr (stdout is used for MCP protocol)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	log.Printf("[INFO] Starting MCP Server for %s database", cfg.DBType)
	log.Printf("[INFO] Database: %s@%s:%d/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("[INFO] Max connections: %d, Max rows: %d, Query timeout: %v",
		cfg.DBMaxConns, cfg.MaxRows, cfg.QueryTimeout)

	// Create database adapter based on type
	var adapter database.Adapter
	switch cfg.DBType {
	case "mysql":
		adapter = database.NewMySQLAdapter(
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBMaxConns,
			cfg.DBMaxIdleConns,
			cfg.DBConnTimeout,
		)
	case "postgres":
		adapter = database.NewPostgresAdapter(
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBMaxConns,
			cfg.DBMaxIdleConns,
			cfg.DBConnTimeout,
		)
	default:
		log.Fatalf("[FATAL] Unsupported database type: %s", cfg.DBType)
	}

	// Create SQL validator
	validator := security.NewValidator(10000) // 10KB max query length

	// Create MCP server
	server := mcp.NewServer(adapter, validator, cfg.MaxRows)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("[INFO] Received shutdown signal")
		cancel()
	}()

	// Run server
	if err := server.Run(ctx); err != nil {
		log.Fatalf("[FATAL] Server error: %v", err)
	}

	log.Printf("[INFO] Server shutdown complete")
}

func init() {
	// Print startup banner to stderr
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔═══════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║     DBHub MCP Server v1.0.0           ║")
	fmt.Fprintln(os.Stderr, "║  Database Operations via MCP Protocol ║")
	fmt.Fprintln(os.Stderr, "╚═══════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
}
