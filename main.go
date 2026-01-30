package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Read API key once at startup
	apiKey := os.Getenv("API_KEY")
	if apiKey != "" {
		log.Printf("API key validation enabled")
	} else {
		log.Printf("API key validation disabled (API_KEY not set)")
	}

	// Create server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes
	mux := http.NewServeMux()

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/models", server.HandleModels)
	mux.HandleFunc("/v1/chat/completions", server.HandleChatCompletions)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Middleware chain: logging -> CORS -> API key validation -> handlers
	handler := loggingMiddleware(corsMiddleware(apiKeyMiddleware(mux, apiKey)))

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")

		// Shutdown HTTP server with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)

		// Stop copilot client with timeout
		done := make(chan struct{})
		go func() {
			server.Close()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Server stopped gracefully")
		case <-time.After(3 * time.Second):
			log.Println("Timeout waiting for copilot client, forcing exit")
		}

		os.Exit(0)
	}()

	log.Printf("Starting OpenAI-compatible Copilot server on http://localhost%s", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  /v1/models")
	log.Printf("  POST /v1/chat/completions")

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// corsMiddleware adds CORS headers for Open WebUI compatibility
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// apiKeyMiddleware validates the API key if one is configured
func apiKeyMiddleware(next http.Handler, expectedKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip validation for health check endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// If no API_KEY is configured, allow any request (current behavior)
		if expectedKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		// API_KEY is configured, so we need to validate
		authHeader := r.Header.Get("Authorization")

		// Extract the key from "Bearer <key>" format
		var providedKey string
		if strings.HasPrefix(authHeader, "Bearer ") {
			providedKey = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		} else {
			providedKey = strings.TrimSpace(authHeader)
		}

		// Reject empty keys
		if providedKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`))
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(expectedKey)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"Invalid API key","type":"invalid_request_error"}}`))
			return
		}

		// Key is valid, proceed
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.size += len(b)
	rw.body.Write(b) // Capture response body
	return rw.ResponseWriter.Write(b)
}

// Flush implements http.Flusher for streaming support
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// loggingMiddleware logs request and response details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and log request body for POST requests
		var requestBody string
		if r.Method == http.MethodPost && r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore the body so it can be read again
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Log incoming request
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		if requestBody != "" {
			log.Printf("  Request Body: %s", truncateBody(requestBody, 10000))
		}

		// Wrap response writer to capture status and body
		wrapped := newResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log response
		log.Printf("← %s %s [%d] %d bytes in %v",
			r.Method, r.URL.Path,
			wrapped.statusCode, wrapped.size, duration)

		// Log response body for non-streaming responses (limited size)
		if wrapped.body.Len() > 0 && wrapped.body.Len() < 5000 {
			log.Printf("  Response Body: %s", truncateBody(wrapped.body.String(), 500))
		}
	})
}

// truncateBody truncates a body string to maxLen characters
func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "... (truncated)"
}
