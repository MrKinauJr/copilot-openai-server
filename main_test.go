package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyMiddleware(t *testing.T) {
	// Create a simple handler that just returns 200 OK
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		apiKey         string
		authHeader     string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No API key configured - allows any key",
			apiKey:         "",
			authHeader:     "Bearer random-key",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "No API key configured - allows no auth header",
			apiKey:         "",
			authHeader:     "",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API key configured - valid key with Bearer prefix",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer secret-key-123",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API key configured - valid key without Bearer prefix",
			apiKey:         "secret-key-123",
			authHeader:     "secret-key-123",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API key configured - invalid key",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer wrong-key",
			path:           "/v1/models",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
		{
			name:           "API key configured - no auth header",
			apiKey:         "secret-key-123",
			authHeader:     "",
			path:           "/v1/models",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
		{
			name:           "API key configured - Bearer with no key",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer ",
			path:           "/v1/models",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
		{
			name:           "API key configured - Bearer with whitespace only",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer   ",
			path:           "/v1/models",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
		{
			name:           "API key configured - valid key with extra whitespace",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer   secret-key-123  ",
			path:           "/v1/models",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Health check bypasses API key validation",
			apiKey:         "secret-key-123",
			authHeader:     "",
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "Health check with wrong key still works",
			apiKey:         "secret-key-123",
			authHeader:     "Bearer wrong-key",
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Wrap the handler with the middleware
			handler := apiKeyMiddleware(nextHandler, tt.apiKey)
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check body
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
