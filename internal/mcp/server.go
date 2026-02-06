package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hieubanhh/dbhubMCP/internal/database"
	"github.com/hieubanhh/dbhubMCP/internal/security"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "dbhub-mcp-server"
	ServerVersion   = "1.0.0"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error)

// Server represents the MCP server
type Server struct {
	transport MessageTransport
	adapter   database.Adapter
	validator *security.Validator
	tools     map[string]ToolHandler
	toolDefs  []Tool
	maxRows   int
	queryTimeout context.Context
}

// NewServer creates a new MCP server
func NewServer(transport MessageTransport, adapter database.Adapter, validator *security.Validator, maxRows int) *Server {
	s := &Server{
		transport: transport,
		adapter:   adapter,
		validator: validator,
		tools:     make(map[string]ToolHandler),
		maxRows:   maxRows,
	}

	// Register tools
	s.registerTools()

	return s
}

// registerTools registers all available tools
func (s *Server) registerTools() {
	// list_tables tool
	s.RegisterTool(Tool{
		Name:        "list_tables",
		Description: "Lists all tables in the connected database. Returns table names, schemas, and types.",
		InputSchema: InputSchema{
			Type:       "object",
			Properties: map[string]Property{},
			Required:   []string{},
		},
	}, s.handleListTables)

	// describe_table tool
	s.RegisterTool(Tool{
		Name:        "describe_table",
		Description: "Describes the schema of a specific table. Returns column names, data types, nullability, defaults, and keys.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"table_name": {
					Type:        "string",
					Description: "The name of the table to describe",
				},
			},
			Required: []string{"table_name"},
		},
	}, s.handleDescribeTable)

	// execute_readonly_query tool
	s.RegisterTool(Tool{
		Name:        "execute_readonly_query",
		Description: "Executes a read-only SQL query (SELECT only). Write operations are strictly blocked. Returns column names and rows.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "The SQL SELECT query to execute",
				},
			},
			Required: []string{"query"},
		},
	}, s.handleExecuteQuery)

	// explain_query tool
	s.RegisterTool(Tool{
		Name:        "explain_query",
		Description: "Returns the execution plan for a SQL query without executing it. Useful for understanding query performance.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "The SQL query to explain",
				},
			},
			Required: []string{"query"},
		},
	}, s.handleExplainQuery)
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.toolDefs = append(s.toolDefs, tool)
	s.tools[tool.Name] = handler
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	log.Printf("[INFO] MCP Server starting with %s transport...", s.transport.GetType())

	// Connect to database
	if err := s.adapter.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer s.adapter.Close()

	log.Printf("[INFO] Connected to %s database", s.adapter.GetDBType())
	log.Printf("[INFO] Registered %d tools", len(s.toolDefs))

	// Start transport
	if err := s.transport.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}
	defer s.transport.Close()

	log.Printf("[INFO] Server ready")

	// Main message loop
	for {
		req, err := s.transport.ReadRequest()
		if err != nil {
			if err == io.EOF {
				log.Printf("[INFO] Client disconnected")
				return nil
			}
			log.Printf("[ERROR] Failed to read request: %v", err)
			continue
		}

		// Handle request
		resp := s.handleRequest(ctx, req)
		if err := s.transport.WriteResponse(resp); err != nil {
			log.Printf("[ERROR] Failed to write response: %v", err)
		}
	}
}

// handleRequest processes an incoming request
func (s *Server) handleRequest(ctx context.Context, req *Request) *Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return s.handleInitialized(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "ping":
		return s.handlePing(req)
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &ErrorObj{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *Request) *Response {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized(req *Request) *Response {
	// This is a notification, no response needed
	log.Printf("[INFO] Client initialized")
	return nil
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *Request) *Response {
	result := ListToolsResult{
		Tools: s.toolDefs,
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(ctx context.Context, req *Request) *Response {
	// Parse params
	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &ErrorObj{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	var params CallToolParams
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &ErrorObj{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Find tool handler
	handler, ok := s.tools[params.Name]
	if !ok {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &ErrorObj{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", params.Name),
			},
		}
	}

	// Execute tool
	result, err := handler(ctx, params.Arguments)
	if err != nil {
		log.Printf("[ERROR] Tool execution failed: %v", err)
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: &CallToolResult{
				Content: []Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Error: %v", err),
					},
				},
				IsError: true,
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handlePing handles the ping request
func (s *Server) handlePing(req *Request) *Response {
	if err := s.adapter.Ping(context.Background()); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &ErrorObj{
				Code:    -32603,
				Message: "Database not available",
				Data:    err.Error(),
			},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]string{"status": "ok"},
	}
}
