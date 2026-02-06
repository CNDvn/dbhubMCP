package mcp

import "context"

// TransportType represents the type of transport being used
type TransportType string

const (
	TransportSTDIO TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// MessageTransport is the interface that all transports must implement
type MessageTransport interface {
	// GetType returns the transport type
	GetType() TransportType

	// Start initializes the transport (e.g., start HTTP server)
	Start(ctx context.Context) error

	// ReadRequest reads the next request from the transport
	ReadRequest() (*Request, error)

	// WriteResponse writes a response to the transport
	WriteResponse(resp *Response) error

	// Close cleans up transport resources
	Close() error
}
