package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"metrics-api/pkg/logger"

	"github.com/dgrijalva/jwt-go"
)

// AuthConfig holds configuration for authentication middleware
type AuthConfig struct {
	JWTSecret      string   // Secret key for JWT validation
	TokenExpiry    int      // Token expiry in minutes
	AllowedOrigins []string // CORS allowed origins
	DisableAuth    bool     // Flag to disable auth (for development)
}

// UserClaims represents the claims in a JWT token
type UserClaims struct {
	UserID string   `json:"userId"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.StandardClaims
}

const (
	userClaimsKey contextKey = "userClaims"
)

// JWTAuth middleware validates JWT tokens
func JWTAuth(config AuthConfig, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if disabled
			if config.DisableAuth {
				log.Debug("Authentication disabled, skipping token validation")
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from request
			tokenString := extractToken(r)
			if tokenString == "" {
				log.Warn("No authentication token provided")
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			// Parse and validate token
			claims, err := validateToken(tokenString, config.JWTSecret)
			if err != nil {
				log.Warnf("Invalid authentication token: %v", err)
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			// Check if token is expired
			if claims.ExpiresAt < time.Now().Unix() {
				log.Warn("Token has expired")
				http.Error(w, "Unauthorized: Token expired", http.StatusUnauthorized)
				return
			}

			// Token is valid, store claims in context
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			
			// Add auth-related headers for downstream services
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
			
			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleAuth middleware verifies that the user has the required role(s)
func RoleAuth(requiredRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(userClaimsKey).(*UserClaims)
			if !ok {
				http.Error(w, "Forbidden: Authentication required", http.StatusForbidden)
				return
			}

			// If no specific roles are required, allow access
			if len(requiredRoles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has any of the required roles
			for _, requiredRole := range requiredRoles {
				for _, userRole := range claims.Roles {
					if requiredRole == userRole {
						// User has the required role, proceed
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// User doesn't have any of the required roles
			http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
		})
	}
}

// ExtractToken gets the JWT token from the Authorization header
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// ValidateToken validates a JWT token and returns the claims
func validateToken(tokenString, secret string) (*UserClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Extract and return the claims
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, email string, roles []string, secret string, expiryMinutes int) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	// Create claims
	claims := &UserClaims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "metrics-api",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserFromContext retrieves user claims from the context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	return claims, ok
}

// HasRole checks if a user has a specific role
func HasRole(ctx context.Context, role string) bool {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}

	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}

	return false
}

// IsAdmin checks if a user has the admin role
func IsAdmin(ctx context.Context) bool {
	return HasRole(ctx, "admin")
}