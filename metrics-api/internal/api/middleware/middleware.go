package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"metrics-api/pkg/logger"

	"github.com/google/uuid"
)

type contextKey string

const (
	requestIDKey contextKey = "requestID"
)

// RequestID middleware adds a unique identifier to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add the request ID to the response header
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID gets the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}


// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log stack trace
					stack := debug.Stack()
					log.Errorf("PANIC: %v\n%s", err, string(stack))
					
					// Return 500 error
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		// Set CORS headers for all responses
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Referer, User-Agent, sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			// Preflight request response
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set content type for regular requests
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
		}
		
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware applies a timeout to the request
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			
			// Create a channel to signal when the request is done
			done := make(chan struct{})
			
			// Process the request in a goroutine
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()
			
			// Wait for the request to finish or timeout
			select {
			case <-done:
				// Request completed successfully
				return
			case <-ctx.Done():
				// Request timed out
				http.Error(w, "Request timeout", http.StatusGatewayTimeout)
				return
			}
		})
	}
}

// WrapResponseWriter is a wrapper for http.ResponseWriter to capture status code and body
type WrapResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	body         []byte
}

// NewWrapResponseWriter creates a new WrapResponseWriter
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	return &WrapResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// Write captures the response body
func (w *WrapResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// Status returns the status code
func (w *WrapResponseWriter) Status() int {
	return w.statusCode
}

// BytesWritten returns the number of bytes written
func (w *WrapResponseWriter) BytesWritten() int {
	return w.bytesWritten
}

// Body returns the response body as a string
func (w *WrapResponseWriter) Body() string {
	return string(w.body)
}

// WriteHeader captures the status code
func (w *WrapResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
