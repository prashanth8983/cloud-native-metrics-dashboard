package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"metrics-api/internal/api/handlers"
	"metrics-api/pkg/logger"

	"github.com/dgrijalva/jwt-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger is a mock implementation of the logger.Logger interface
type MockLogger struct {
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
	DebugMessages []string
	Fields        map[string]interface{}
}

func NewMockLogger() *MockLogger {
    return &MockLogger{
        InfoMessages:  make([]string, 0),
        WarnMessages:  make([]string, 0),
        ErrorMessages: make([]string, 0),
        DebugMessages: make([]string, 0),
        Fields:        make(map[string]interface{}),
    }
}

func (m *MockLogger) Sync() error {
    return nil
}

func (m *MockLogger) Info(args ...interface{}) {
    m.InfoMessages = append(m.InfoMessages, fmt.Sprint(args...))
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.InfoMessages = append(m.InfoMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.WarnMessages = append(m.WarnMessages, fmt.Sprint(args...))
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.WarnMessages = append(m.WarnMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Error(args ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, fmt.Sprint(args...))
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.ErrorMessages = append(m.ErrorMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, fmt.Sprint(args...))
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.DebugMessages = append(m.DebugMessages, fmt.Sprintf(format, args...))
}

func (m *MockLogger) Fatal(args ...interface{}) {
	panic(fmt.Sprint(args...))
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
    for k, v := range fields {
        m.Fields[k] = v
    }
    return m
}

func (m *MockLogger) With(fields map[string]interface{}) logger.Logger {
    // Create a new logger instance to avoid modifying the original
    newLogger := NewMockLogger()
    // Copy existing fields
    for k, v := range m.Fields {
        newLogger.Fields[k] = v
    }
    // Add new fields
    for k, v := range fields {
        newLogger.Fields[k] = v
    }
    // Copy message history if needed
    newLogger.InfoMessages = append([]string{}, m.InfoMessages...)
    newLogger.WarnMessages = append([]string{}, m.WarnMessages...)
    newLogger.ErrorMessages = append([]string{}, m.ErrorMessages...)
    newLogger.DebugMessages = append([]string{}, m.DebugMessages...)
    return newLogger
}

// Helper function to create a test request
func createTestRequest(method, path string, headers map[string]string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

// Test RequestID middleware
func TestRequestIDMiddleware(t *testing.T) {
	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		fmt.Fprint(w, requestID)
	})

	// Create middleware
	middleware := RequestID(handler)

	// Test with no existing request ID
	t.Run("No Existing Request ID", func(t *testing.T) {
		req := createTestRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Body.String())
		assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
	})

	// Test with existing request ID
	t.Run("Existing Request ID", func(t *testing.T) {
		existingID := "test-request-id-123"
		req := createTestRequest("GET", "/test", map[string]string{
			"X-Request-ID": existingID,
		})
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, existingID, rr.Body.String())
		assert.Equal(t, existingID, rr.Header().Get("X-Request-ID"))
	})
}

// Test LoggingMiddleware
func TestLoggingMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := NewMockLogger()

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := LoggingMiddleware(mockLogger)(handler)

	// Create request
	req := createTestRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Call the middleware
	middleware.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())

	// Check logs
	assert.GreaterOrEqual(t, len(mockLogger.InfoMessages), 2)
	assert.Contains(t, mockLogger.InfoMessages[0], "Request started")
	assert.Contains(t, mockLogger.InfoMessages[1], "Request completed")
}

// Test RecoveryMiddleware
func TestRecoveryMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := NewMockLogger()

	// Create a handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create middleware
	middleware := RecoveryMiddleware(mockLogger)(handler)

	// Create request
	req := createTestRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// This should not panic
	middleware.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Check logs
	assert.GreaterOrEqual(t, len(mockLogger.ErrorMessages), 1)
	assert.Contains(t, mockLogger.ErrorMessages[0], "Panic recovered")
	assert.Contains(t, mockLogger.ErrorMessages[0], "test panic")
}

// Test JWTAuth middleware
func TestJWTAuthMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := NewMockLogger()

	// Create config
	authConfig := AuthConfig{
		JWTSecret:   "test-secret",
		TokenExpiry: 60,
	}

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims
		claims, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}
		// Return user ID
		fmt.Fprint(w, claims.UserID)
	})

	// Create middleware
	middleware := JWTAuth(authConfig, mockLogger)(handler)

	// Test with valid token
	t.Run("Valid Token", func(t *testing.T) {
		// Generate a valid token
		token, err := GenerateToken("test-user", "test@example.com", []string{"admin"}, authConfig.JWTSecret, authConfig.TokenExpiry)
		require.NoError(t, err)

		req := createTestRequest("GET", "/test", map[string]string{
			"Authorization": "Bearer " + token,
		})
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test-user", rr.Body.String())
	})

	// Test with invalid token
	t.Run("Invalid Token", func(t *testing.T) {
		req := createTestRequest("GET", "/test", map[string]string{
			"Authorization": "Bearer invalid-token",
		})
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
	})

	// Test with missing token
	t.Run("Missing Token", func(t *testing.T) {
		req := createTestRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "No token provided")
	})

	// Test with expired token
	t.Run("Expired Token", func(t *testing.T) {
		// Create claims with past expiry
		claims := &UserClaims{
			UserID: "test-user",
			Email:  "test@example.com",
			Roles:  []string{"admin"},
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
				IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			},
		}

		// Create token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(authConfig.JWTSecret))
		require.NoError(t, err)

		req := createTestRequest("GET", "/test", map[string]string{
			"Authorization": "Bearer " + tokenString,
		})
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Token expired")
	})

	// Test with auth disabled
	t.Run("Auth Disabled", func(t *testing.T) {
		disabledConfig := AuthConfig{
			JWTSecret:   "test-secret",
			TokenExpiry: 60,
			DisableAuth: true,
		}

		disabledMiddleware := JWTAuth(disabledConfig, mockLogger)(handler)

		req := createTestRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		// This should pass even without a token
		disabledMiddleware.ServeHTTP(rr, req)

		// The handler will fail because no user in context, but middleware passed
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "User not found in context")
	})
}

// Test RoleAuth middleware
func TestRoleAuthMiddleware(t *testing.T) {
	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	adminMiddleware := RoleAuth([]string{"admin"})(handler)
	userMiddleware := RoleAuth([]string{"user"})(handler)
	multiRoleMiddleware := RoleAuth([]string{"admin", "operator"})(handler)

	// Test with admin role
	t.Run("Has Admin Role", func(t *testing.T) {
		claims := &UserClaims{
			UserID: "test-user",
			Email:  "test@example.com",
			Roles:  []string{"admin"},
		}

		// Create request with claims in context
		req := createTestRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), userClaimsKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		adminMiddleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())
	})

	// Test with missing required role
	t.Run("Missing Required Role", func(t *testing.T) {
		claims := &UserClaims{
			UserID: "test-user",
			Email:  "test@example.com",
			Roles:  []string{"user"},
		}

		// Create request with claims in context
		req := createTestRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), userClaimsKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		adminMiddleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Insufficient permissions")
	})

	// Test with one of multiple required roles
	t.Run("Has One Of Required Roles", func(t *testing.T) {
		claims := &UserClaims{
			UserID: "test-user",
			Email:  "test@example.com",
			Roles:  []string{"operator"},
		}

		// Create request with claims in context
		req := createTestRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), userClaimsKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		multiRoleMiddleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())
	})

	// Test with missing authentication
	t.Run("Missing Authentication", func(t *testing.T) {
		req := createTestRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		userMiddleware.ServeHTTP(rr, req)

		// Check response
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication required")
	})
}

// Test token generation and validation
func TestTokenGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key"
	userID := "test-user-id"
	email := "test@example.com"
	roles := []string{"admin", "user"}
	expiryMinutes := 60

	// Generate token
	token, err := GenerateToken(userID, email, roles, secret, expiryMinutes)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := validateToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Greater(t, claims.ExpiresAt, time.Now().Unix())
}

// Test MetricsMiddleware
func TestMetricsMiddleware(t *testing.T) {
	// Create metrics middleware
	metricsMiddleware := NewMetricsMiddleware()

	// Reset Prometheus registry for testing
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Create a simple test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware
	middleware := metricsMiddleware.Middleware()(handler)

	// Create and execute request
	req := createTestRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())

	// We can't directly test the metrics values, but we can ensure middleware doesn't break
}

// Test WrapResponseWriter
func TestWrapResponseWriter(t *testing.T) {
	// Create a response recorder
	rr := httptest.NewRecorder()

	// Create a wrapped response writer
	wrw := NewWrapResponseWriter(rr)

	// Test initial state
	assert.Equal(t, http.StatusOK, wrw.Status())
	assert.Equal(t, 0, wrw.BytesWritten())

	// Test WriteHeader
	wrw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, wrw.Status())

	// Test Write
	n, err := wrw.Write([]byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, 9, wrw.BytesWritten())

	// Write more data
	n, err = wrw.Write([]byte(" more data"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, 19, wrw.BytesWritten())

	// Check final response
	assert.Equal(t, "test data more data", rr.Body.String())
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// Test RequestDurationMiddleware
func TestRequestDurationMiddleware(t *testing.T) {
	// Create mock logger
	mockLogger := NewMockLogger()

	// Create a slow handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Sleep for 10ms
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware with 5ms threshold
	middleware := RequestDurationMiddleware(mockLogger, 5*time.Millisecond)(handler)

	// Create request
	req := createTestRequest("GET", "/test", nil)
	// Add request ID to context
	ctx := context.WithValue(req.Context(), requestIDKey, "test-request-id")
	req = req.WithContext(ctx)
	
	rr := httptest.NewRecorder()

	// Call the middleware
	middleware.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check logs
	assert.GreaterOrEqual(t, len(mockLogger.WarnMessages), 1)
	assert.Contains(t, mockLogger.WarnMessages[0], "Slow request")
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	// Test getClientIP
	t.Run("getClientIP", func(t *testing.T) {
		// Test with X-Forwarded-For
		req := createTestRequest("GET", "/test", map[string]string{
			"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
		})
		ip := getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)

		// Test with X-Real-IP
		req = createTestRequest("GET", "/test", map[string]string{
			"X-Real-IP": "192.168.1.2",
		})
		ip = getClientIP(req)
		assert.Equal(t, "192.168.1.2", ip)

		// Test with RemoteAddr
		req = createTestRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"
		ip = getClientIP(req)
		assert.Equal(t, "192.168.1.3", ip)
	})

	// Test sanitizeQuery
	t.Run("sanitizeQuery", func(t *testing.T) {
		// Test with sensitive parameters
		query := "name=John&password=secret123&token=abc&key=xyz"
		sanitized := sanitizeQuery(query)
		assert.Contains(t, sanitized, "name=John")
		assert.Contains(t, sanitized, "password=[REDACTED]")
		assert.Contains(t, sanitized, "token=[REDACTED]")
		assert.Contains(t, sanitized, "key=[REDACTED]")

		// Test with no sensitive parameters
		query = "name=John&age=30&sort=desc"
		sanitized = sanitizeQuery(query)
		assert.Equal(t, query, sanitized)

		// Test with empty query
		query = ""
		sanitized = sanitizeQuery(query)
		assert.Equal(t, query, sanitized)
	})
}

// Test response helpers for completeness
func TestResponseHelpers(t *testing.T) {
	// Test respondWithJSON
	t.Run("respondWithJSON", func(t *testing.T) {
		rr := httptest.NewRecorder()
		data := map[string]interface{}{
			"message": "test message",
			"code":    200,
		}
		handlers.RespondWithJSON(rr, http.StatusOK, data)

		// Check response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test message", response["message"])
		assert.Equal(t, float64(200), response["code"])
	})

	// Test respondWithError
	t.Run("respondWithError", func(t *testing.T) {
		rr := httptest.NewRecorder()
		handlers.RespondWithError(rr, http.StatusBadRequest, "Invalid request")

		// Check response
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		// Parse response
		var response handlers.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", response.Error)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "Invalid request", response.Message)
	})
}