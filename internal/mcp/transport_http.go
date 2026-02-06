package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// HTTPTransportConfig holds configuration for HTTP transport
type HTTPTransportConfig struct {
	Addr        string   // Server address (e.g., ":8080")
	CORSOrigins []string // Allowed CORS origins (e.g., ["*"] or ["https://example.com"])
	APIKey      string   // Optional API key for authentication
}

// HTTPTransport handles HTTP-based communication
type HTTPTransport struct {
	server       *http.Server
	addr         string
	corsOrigins  []string
	apiKey       string
	requestChan  chan *httpRequest
	responseChan map[string]chan *Response
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// httpRequest wraps a request with its response channel
type httpRequest struct {
	req      *Request
	respChan chan *Response
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(config HTTPTransportConfig) *HTTPTransport {
	ctx, cancel := context.WithCancel(context.Background())

	t := &HTTPTransport{
		addr:         config.Addr,
		corsOrigins:  config.CORSOrigins,
		apiKey:       config.APIKey,
		requestChan:  make(chan *httpRequest, 10), // Buffered channel for concurrent requests
		responseChan: make(map[string]chan *Response),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", t.handleMCPRequest)
	mux.HandleFunc("/health", t.handleHealthCheck)

	t.server = &http.Server{
		Addr:         config.Addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return t
}

// GetType returns the transport type
func (t *HTTPTransport) GetType() TransportType {
	return TransportHTTP
}

// Start initializes the HTTP server
func (t *HTTPTransport) Start(ctx context.Context) error {
	// Start HTTP server in background
	go func() {
		log.Printf("[INFO] HTTP server listening on %s", t.addr)
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] HTTP server error: %v", err)
		}
	}()

	// Start response router in background
	go t.routeResponses()

	return nil
}

// ReadRequest reads the next request from the channel
func (t *HTTPTransport) ReadRequest() (*Request, error) {
	select {
	case httpReq := <-t.requestChan:
		// Store response channel for this request
		reqID := fmt.Sprintf("%v", httpReq.req.ID)
		t.mu.Lock()
		t.responseChan[reqID] = httpReq.respChan
		t.mu.Unlock()
		return httpReq.req, nil
	case <-t.ctx.Done():
		return nil, fmt.Errorf("transport closed")
	}
}

// WriteResponse writes a response to the appropriate channel
func (t *HTTPTransport) WriteResponse(resp *Response) error {
	if resp == nil {
		// This is a notification (no response needed)
		return nil
	}

	reqID := fmt.Sprintf("%v", resp.ID)
	t.mu.RLock()
	respChan, ok := t.responseChan[reqID]
	t.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no response channel found for request ID: %v", resp.ID)
	}

	// Send response with timeout
	select {
	case respChan <- resp:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout writing response for request ID: %v", resp.ID)
	}
}

// Close shuts down the HTTP server
func (t *HTTPTransport) Close() error {
	log.Printf("[INFO] Shutting down HTTP server...")
	t.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := t.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	log.Printf("[INFO] HTTP server shutdown complete")
	return nil
}

// handleMCPRequest handles incoming MCP requests via HTTP
func (t *HTTPTransport) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	t.setCORSHeaders(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check API key if configured
	if t.apiKey != "" {
		providedKey := r.Header.Get("X-API-Key")
		if providedKey != t.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse request body
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] HTTP request: method=%s id=%v", req.Method, req.ID)

	// Create response channel for this request
	respChan := make(chan *Response, 1)

	// Send request to main processing loop
	httpReq := &httpRequest{
		req:      &req,
		respChan: respChan,
	}

	select {
	case t.requestChan <- httpReq:
		// Request queued successfully
	case <-time.After(5 * time.Second):
		http.Error(w, "Server busy", http.StatusServiceUnavailable)
		return
	}

	// Wait for response with timeout
	select {
	case resp := <-respChan:
		// Clean up response channel
		reqID := fmt.Sprintf("%v", req.ID)
		t.mu.Lock()
		delete(t.responseChan, reqID)
		t.mu.Unlock()

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("[ERROR] Failed to encode response: %v", err)
		}
		log.Printf("[DEBUG] HTTP response sent: id=%v", resp.ID)

	case <-time.After(60 * time.Second):
		// Timeout waiting for response
		reqID := fmt.Sprintf("%v", req.ID)
		t.mu.Lock()
		delete(t.responseChan, reqID)
		t.mu.Unlock()

		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
	}
}

// handleHealthCheck handles health check requests
func (t *HTTPTransport) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	t.setCORSHeaders(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// setCORSHeaders sets CORS headers based on configuration
func (t *HTTPTransport) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range t.corsOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			if allowedOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			break
		}
	}

	if !allowed && origin != "" {
		// If no wildcard and origin doesn't match, don't set CORS headers
		return
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// routeResponses handles routing responses back to HTTP handlers
func (t *HTTPTransport) routeResponses() {
	<-t.ctx.Done()
	// Clean up any pending response channels
	t.mu.Lock()
	for _, ch := range t.responseChan {
		close(ch)
	}
	t.responseChan = make(map[string]chan *Response)
	t.mu.Unlock()
}
