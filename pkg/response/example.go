package response

import (
	"context"
	"net/http"
)

// Example usage of the response package
func ExampleUsage(ctx context.Context, w http.ResponseWriter) {
	// This is just an example - not actual code that runs

	// Success response
	Success(ctx, "Operation completed", map[string]string{"id": "123"}).Send(w, http.StatusOK)

	// Error responses
	BadRequest(ctx, "Invalid input", []string{"name is required"}).Send(w, http.StatusBadRequest)
	Unauthorized(ctx, "Invalid credentials", []string{"token expired"}).Send(w, http.StatusUnauthorized)
	NotFound(ctx, "Resource not found", []string{"user not found"}).Send(w, http.StatusNotFound)
	Conflict(ctx, "Resource already exists", []string{"email already taken"}).Send(w, http.StatusConflict)
	ValidationError(ctx, "Validation failed", []string{"email format invalid"}).Send(w, http.StatusBadRequest)
	InternalServerError(ctx, "Something went wrong", []string{"database connection failed"}).Send(w, http.StatusInternalServerError)

	// Custom response with builder pattern
	New(ctx).
		WithCode("CUSTOM_CODE").
		WithMessage("Custom message").
		WithData(map[string]interface{}{"custom": "data"}).
		WithErrors([]string{"custom error"}).
		Send(w, http.StatusOK)
}
