package main

import (
	"net/http"
	"net/http/httptest"
	"os"
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
		envAPIKey      string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No API_KEY env var set - allows any key",
			envAPIKey:      "",
			authHeader:     "Bearer random-key",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "No API_KEY env var set - allows no auth header",
			envAPIKey:      "",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API_KEY set - valid key with Bearer prefix",
			envAPIKey:      "secret-key-123",
			authHeader:     "Bearer secret-key-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API_KEY set - valid key without Bearer prefix",
			envAPIKey:      "secret-key-123",
			authHeader:     "secret-key-123",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "API_KEY set - invalid key",
			envAPIKey:      "secret-key-123",
			authHeader:     "Bearer wrong-key",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
		{
			name:           "API_KEY set - no auth header",
			envAPIKey:      "secret-key-123",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			if tt.envAPIKey != "" {
				os.Setenv("API_KEY", tt.envAPIKey)
				defer os.Unsetenv("API_KEY")
			} else {
				os.Unsetenv("API_KEY")
			}

			// Create a request
			req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Wrap the handler with the middleware
			handler := apiKeyMiddleware(nextHandler)
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
