package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/fanzru/social-media-service-go/pkg/jwt"
	"github.com/fanzru/social-media-service-go/pkg/logger"
	"github.com/fanzru/social-media-service-go/pkg/reqctx"
	"github.com/fanzru/social-media-service-go/pkg/response"
)

// AuthMiddleware handles authentication based on OpenAPI spec security requirements
type AuthMiddleware struct {
	jwtService *jwt.Service
	// Map of path patterns to their security requirements
	// Key: HTTP method + path pattern (e.g., "GET /api/account/profile")
	// Value: whether authentication is required
	securityMap map[string]bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *jwt.Service) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		securityMap: make(map[string]bool),
	}
}

// AddSecurityRequirement adds a security requirement for a specific endpoint
func (m *AuthMiddleware) AddSecurityRequirement(method, path string, requiresAuth bool) {
	key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	m.securityMap[key] = requiresAuth
}

// Middleware returns the authentication middleware function
func (m *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := reqctx.GetRequestID(ctx)

			// Always allow CORS preflight without auth
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check if this endpoint requires authentication
			requiresAuth := m.requiresAuthFor(r.Method, r.URL.Path)

			// If no auth required, proceed directly
			if !requiresAuth {
				logger.GetGlobal().Info("No authentication required",
					"requestId", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.GetGlobal().Warn("Missing authorization header",
					"requestId", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)
				response.Unauthorized(ctx, "Authorization header required", []string{"Missing Authorization header"}).Send(w, http.StatusUnauthorized)
				return
			}

			// Check if it's Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.GetGlobal().Warn("Invalid authorization header format",
					"requestId", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"authHeader", "[REDACTED]",
				)
				response.Unauthorized(ctx, "Invalid authorization header format", []string{"Authorization header must start with 'Bearer '"}).Send(w, http.StatusUnauthorized)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				logger.GetGlobal().Warn("Empty token",
					"requestId", requestID,
					"method", r.Method,
					"path", r.URL.Path,
				)
				response.Unauthorized(ctx, "Token required", []string{"Bearer token cannot be empty"}).Send(w, http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := m.jwtService.ValidateToken(token)
			if err != nil {
				logger.GetGlobal().Warn("Invalid token",
					"requestId", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"error", err.Error(),
				)
				response.Unauthorized(ctx, "Invalid token", []string{err.Error()}).Send(w, http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx = context.WithValue(ctx, "user_id", claims.AccountID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_name", claims.Name)

			logger.GetGlobal().Info("Authentication successful",
				"requestId", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"user_id", claims.AccountID,
				"user_email", claims.Email,
			)

			// Proceed with authenticated request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requiresAuthFor determines whether auth is required for a given method and path.
// It first tries an exact match, then falls back to prefix-based matching to support
// dynamic path segments like "/api/comments/by-post/{postId}".
func (m *AuthMiddleware) requiresAuthFor(method, path string) bool {
	// 1) Exact match
	exactKey := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	if v, ok := m.securityMap[exactKey]; ok {
		return v
	}

	// 2) Prefix match against registered patterns
	// Example: ruleKey "GET /api/comments/by-post" matches
	//          request path "/api/comments/by-post/5"
	method = strings.ToUpper(method)
	for k, v := range m.securityMap {
		// Expect keys in format: "METHOD /path"
		if !strings.HasPrefix(k, method+" ") {
			continue
		}
		rulePath := strings.TrimPrefix(k, method+" ")

		if rulePath == path {
			return v
		}
		// Normalize: ensure rulePath without trailing slash compares to path segments
		if strings.HasSuffix(rulePath, "/") {
			rulePath = strings.TrimSuffix(rulePath, "/")
		}

		// If request path starts with rulePath followed by a slash, consider it a match
		if rulePath != "" && strings.HasPrefix(path, rulePath+"/") {
			return v
		}
	}
	// Default: no auth required if not specified
	return false
}

// Helper functions to get user info from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value("user_id").(int64)
	return userID, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value("user_email").(string)
	return email, ok
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value("user_name").(string)
	return name, ok
}
