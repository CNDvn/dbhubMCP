package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTransport_GetType(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":8080",
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	if transport.GetType() != TransportHTTP {
		t.Errorf("Expected transport type %s, got %s", TransportHTTP, transport.GetType())
	}
}

func TestHTTPTransport_StartStop(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":18080", // Use different port for testing
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	ctx := context.Background()
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	if err := transport.Close(); err != nil {
		t.Errorf("Failed to close transport: %v", err)
	}
}

func TestHTTPTransport_HealthCheck(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":8080",
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	transport.handleHealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestHTTPTransport_CORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		corsOrigins    []string
		requestOrigin  string
		expectedHeader string
	}{
		{
			name:           "Wildcard CORS",
			corsOrigins:    []string{"*"},
			requestOrigin:  "http://example.com",
			expectedHeader: "*",
		},
		{
			name:           "Specific origin match",
			corsOrigins:    []string{"http://example.com"},
			requestOrigin:  "http://example.com",
			expectedHeader: "http://example.com",
		},
		{
			name:           "No origin match",
			corsOrigins:    []string{"http://example.com"},
			requestOrigin:  "http://evil.com",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewHTTPTransport(HTTPTransportConfig{
				Addr:        ":8080",
				CORSOrigins: tt.corsOrigins,
				APIKey:      "",
			})

			req := httptest.NewRequest("OPTIONS", "/mcp", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()

			transport.handleMCPRequest(w, req)

			corsHeader := w.Header().Get("Access-Control-Allow-Origin")
			if corsHeader != tt.expectedHeader {
				t.Errorf("Expected CORS header '%s', got '%s'", tt.expectedHeader, corsHeader)
			}
		})
	}
}

func TestHTTPTransport_Authentication(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		providedKey    string
		expectedStatus int
	}{
		{
			name:           "Incorrect API key",
			apiKey:         "secret-key",
			providedKey:    "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing API key",
			apiKey:         "secret-key",
			providedKey:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewHTTPTransport(HTTPTransportConfig{
				Addr:        ":8080",
				CORSOrigins: []string{"*"},
				APIKey:      tt.apiKey,
			})

			reqBody := Request{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "ping",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(body))
			if tt.providedKey != "" {
				req.Header.Set("X-API-Key", tt.providedKey)
			}
			w := httptest.NewRecorder()

			transport.handleMCPRequest(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHTTPTransport_MethodNotAllowed(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":8080",
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	transport.handleMCPRequest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHTTPTransport_InvalidJSON(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":8080",
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	transport.handleMCPRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTPTransport_ConcurrentRequests(t *testing.T) {
	transport := NewHTTPTransport(HTTPTransportConfig{
		Addr:        ":18081",
		CORSOrigins: []string{"*"},
		APIKey:      "",
	})

	ctx := context.Background()
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("Failed to start transport: %v", err)
	}
	defer transport.Close()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Simulate concurrent requests
	numRequests := 5
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			reqBody := Request{
				JSONRPC: "2.0",
				ID:      id,
				Method:  "ping",
			}
			body, _ := json.Marshal(reqBody)

			resp, err := http.Post("http://localhost:18081/health", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Errorf("Request %d failed: %v", id, err)
			} else {
				resp.Body.Close()
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			// Request completed
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}
