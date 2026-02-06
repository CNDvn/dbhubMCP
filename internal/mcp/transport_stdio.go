package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// StdioTransport handles STDIO-based communication
type StdioTransport struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewStdioTransport creates a new STDIO transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// GetType returns the transport type
func (t *StdioTransport) GetType() TransportType {
	return TransportSTDIO
}

// Start initializes the transport (no-op for STDIO)
func (t *StdioTransport) Start(ctx context.Context) error {
	return nil
}

// ReadRequest reads and parses a JSON-RPC request from stdin
func (t *StdioTransport) ReadRequest() (*Request, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	log.Printf("[DEBUG] Received request: method=%s id=%v", req.Method, req.ID)
	return &req, nil
}

// WriteResponse writes a JSON-RPC response to stdout
func (t *StdioTransport) WriteResponse(resp *Response) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write the JSON followed by a newline
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	if _, err := t.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	log.Printf("[DEBUG] Sent response: id=%v hasError=%v", resp.ID, resp.Error != nil)
	return nil
}

// Close cleans up transport resources (no-op for STDIO)
func (t *StdioTransport) Close() error {
	return nil
}

// WriteError writes an error response
func (t *StdioTransport) WriteError(id interface{}, code int, message string, data interface{}) error {
	resp := &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrorObj{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return t.WriteResponse(resp)
}
