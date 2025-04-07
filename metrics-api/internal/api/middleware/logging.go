package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"metrics-api/pkg/logger"
)

// LoggingMiddleware logs incoming HTTP requests and their responses
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture the status code
			wrw := NewWrapResponseWriter(w)

			// Get the request ID from context (added by RequestID middleware)
			requestID := GetRequestID(r.Context())

			// Get user info from context if available
			var userID string
			if claims, ok := GetUserFromContext(r.Context()); ok && claims != nil {
				userID = claims.UserID
			}

			// Prepare initial log data
			logData := map[string]interface{}{
				"request_id":  requestID,
				"remote_addr": getClientIP(r),
				"user_agent":  r.UserAgent(),
				"method":      r.Method,
				"path":        r.URL.Path,
				"query":       sanitizeQuery(r.URL.RawQuery),
				"user_id":     userID,
				"referer":     r.Referer(),
			}

			// Log request start
			log.WithFields(logData).Infof("Request started: %s %s", r.Method, r.URL.Path)

			// Process request
			defer func() {
				// Recover from panic
				if err := recover(); err != nil {
					stack := string(debug.Stack())
					log.WithFields(logData).Errorf("Panic: %v\n%s", err, stack)
					http.Error(wrw, "Internal Server Error", http.StatusInternalServerError)
				}

				// Calculate request duration
				duration := time.Since(start)

				// Add response data to the log
				logData["status"] = wrw.Status()
				logData["size"] = wrw.BytesWritten()
				logData["duration"] = duration.Milliseconds()

				// Determine log level based on status code
				if wrw.Status() >= 500 {
					log.WithFields(logData).Errorf("Request completed: %s %s %d %s",
						r.Method, r.URL.Path, wrw.Status(), duration)
				} else if wrw.Status() >= 400 {
					log.WithFields(logData).Warnf("Request completed: %s %s %d %s",
						r.Method, r.URL.Path, wrw.Status(), duration)
				} else {
					log.WithFields(logData).Infof("Request completed: %s %s %d %s",
						r.Method, r.URL.Path, wrw.Status(), duration)
				}
			}()

			// Proceed with the request
			next.ServeHTTP(wrw, r)
		})
	}
}

// Ensure our wrapper implements http.ResponseWriter
var _ http.ResponseWriter = &WrapResponseWriter{}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}

	// Check for X-Real-IP header next
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.IndexByte(ip, ':'); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// sanitizeQuery removes sensitive information from query string
func sanitizeQuery(query string) string {
	if query == "" {
		return query
	}

	// List of parameter names to redact
	sensitiveParams := []string{"token", "password", "secret", "key", "auth", "credentials", "pwd"}

	parts := strings.Split(query, "&")
	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) < 2 {
			continue
		}

		key := strings.ToLower(kv[0])

		// Check if this is a sensitive parameter
		for _, param := range sensitiveParams {
			if strings.Contains(key, param) {
				parts[i] = fmt.Sprintf("%s=[REDACTED]", kv[0])
				break
			}
		}
	}

	return strings.Join(parts, "&")
}

// RequestDurationMiddleware measures request duration and logs slow requests
func RequestDurationMiddleware(log logger.Logger, slowThreshold time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer
			start := time.Now()

			// Process request
			next.ServeHTTP(w, r)

			// Calculate duration
			duration := time.Since(start)

			// Log slow requests
			if duration >= slowThreshold {
				log.WithFields(map[string]interface{}{
					"request_id":  GetRequestID(r.Context()),
					"method":      r.Method,
					"path":        r.URL.Path,
					"duration_ms": duration.Milliseconds(),
				}).Warnf("Slow request: %s %s took %s", r.Method, r.URL.Path, duration)
			}
		})
	}
}

// LogHTTPErrorMiddleware logs HTTP errors with detailed information
func LogHTTPErrorMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrw := NewWrapResponseWriter(w)
			next.ServeHTTP(wrw, r)

			// Log client and server errors
			if wrw.Status() >= 400 {
				severity := "warning"
				if wrw.Status() >= 500 {
					severity = "error"
				}

				log.WithFields(map[string]interface{}{
					"request_id":  GetRequestID(r.Context()),
					"method":      r.Method,
					"path":        r.URL.Path,
					"status_code": wrw.Status(),
					"severity":    severity,
					"user_agent":  r.UserAgent(),
					"client_ip":   getClientIP(r),
				}).Infof("HTTP %d error: %s %s", wrw.Status(), r.Method, r.URL.Path)
			}
		})
	}
}
