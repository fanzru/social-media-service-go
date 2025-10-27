package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fanzru/social-media-service-go/internal/app/account"
	"github.com/fanzru/social-media-service-go/internal/app/account/app"
	"github.com/fanzru/social-media-service-go/pkg/middleware"
	"github.com/fanzru/social-media-service-go/pkg/response"
)

// Handler handles HTTP requests for account operations
// Implements genhttp.ServerInterface
type Handler struct {
	service app.Service
}

// NewHandler creates a new account handler
func NewHandler(service app.Service) *Handler {
	return &Handler{service: service}
}

// PostApiAccountRegister implements genhttp.ServerInterface
func (h *Handler) PostApiAccountRegister(w http.ResponseWriter, r *http.Request) {
	h.Register(w, r)
}

// PostApiAccountLogin implements genhttp.ServerInterface
func (h *Handler) PostApiAccountLogin(w http.ResponseWriter, r *http.Request) {
	h.Login(w, r)
}

// GetApiAccountProfile implements genhttp.ServerInterface
func (h *Handler) GetApiAccountProfile(w http.ResponseWriter, r *http.Request) {
	h.GetProfile(w, r)
}

// Register handles account registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req account.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateRegisterRequest(&req); err != nil {
		response.ValidationError(ctx, "Validation failed", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	// Register account
	acc, err := h.service.Register(ctx, &req)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Conflict(ctx, "Email already exists", []string{err.Error()}).Send(w, http.StatusConflict)
			return
		}
		response.InternalServerError(ctx, "Failed to register account", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	// Send success response
	response.Success(ctx, "Account registered successfully", acc).Send(w, http.StatusCreated)
}

// Login handles account login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req account.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(ctx, "Invalid request body", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateLoginRequest(&req); err != nil {
		response.ValidationError(ctx, "Validation failed", []string{err.Error()}).Send(w, http.StatusBadRequest)
		return
	}

	// Login account
	loginResp, err := h.service.Login(ctx, &req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			response.Unauthorized(ctx, "Invalid credentials", []string{err.Error()}).Send(w, http.StatusUnauthorized)
			return
		}
		response.InternalServerError(ctx, "Failed to login", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	// Send success response
	response.Success(ctx, "Login successful", loginResp).Send(w, http.StatusOK)
}

// GetProfile handles getting account profile (requires authentication)
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Unauthorized(ctx, "User not authenticated", []string{"Missing user ID in context"}).Send(w, http.StatusUnauthorized)
		return
	}

	// Get account by ID
	acc, err := h.service.GetAccountByID(ctx, userID)
	if err != nil {
		response.InternalServerError(ctx, "Failed to get account profile", []string{err.Error()}).Send(w, http.StatusInternalServerError)
		return
	}

	// Send success response
	response.Success(ctx, "Profile retrieved successfully", acc).Send(w, http.StatusOK)
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response.Success(ctx, "Service is healthy", map[string]string{
		"status": "ok",
	}).Send(w, http.StatusOK)
}

// validateRegisterRequest validates the register request
func validateRegisterRequest(req *account.RegisterRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(req.Name) > 100 {
		return fmt.Errorf("name must be at most 100 characters")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

// validateLoginRequest validates the login request
func validateLoginRequest(req *account.LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Simple email validation - in production use a proper email validation library
	return len(email) > 0 && len(email) < 255 &&
		contains(email, "@") &&
		contains(email, ".") &&
		!startsWith(email, "@") &&
		!endsWith(email, "@")
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					contains(s[1:], substr))))
}

// startsWith checks if string starts with prefix
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// endsWith checks if string ends with suffix
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
