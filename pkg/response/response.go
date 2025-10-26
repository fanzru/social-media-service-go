package response

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fanzru/social-media-service-go/pkg/reqctx"
)

// Response represents the standard API response format
type Response struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Errors     []string    `json:"errors,omitempty"`
	ServerTime string      `json:"serverTime"`
	RequestID  string      `json:"requestId"`
	Data       interface{} `json:"data,omitempty"`
}

// ResponseBuilder helps build standardized responses
type ResponseBuilder struct {
	response *Response
	ctx      context.Context
}

// New creates a new response builder
func New(ctx context.Context) *ResponseBuilder {
	return &ResponseBuilder{
		response: &Response{
			ServerTime: time.Now().Format(time.RFC3339),
			RequestID:  reqctx.GetRequestID(ctx),
		},
		ctx: ctx,
	}
}

// WithCode sets the response code
func (rb *ResponseBuilder) WithCode(code string) *ResponseBuilder {
	rb.response.Code = code
	return rb
}

// WithMessage sets the response message
func (rb *ResponseBuilder) WithMessage(message string) *ResponseBuilder {
	rb.response.Message = message
	return rb
}

// WithErrors sets the response errors
func (rb *ResponseBuilder) WithErrors(errors []string) *ResponseBuilder {
	rb.response.Errors = errors
	return rb
}

// WithData sets the response data
func (rb *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	return rb
}

// WithRequestID sets a custom request ID
func (rb *ResponseBuilder) WithRequestID(requestID string) *ResponseBuilder {
	rb.response.RequestID = requestID
	return rb
}

// Send sends the response with the specified status code
func (rb *ResponseBuilder) Send(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(rb.response)
}

// Success creates a success response
func Success(ctx context.Context, message string, data interface{}) *ResponseBuilder {
	return New(ctx).
		WithCode("SUCCESS").
		WithMessage(message).
		WithData(data)
}

// BadRequest creates a bad request response
func BadRequest(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("BAD_REQUEST").
		WithMessage(message).
		WithErrors(errors)
}

// Unauthorized creates an unauthorized response
func Unauthorized(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("UNAUTHORIZED").
		WithMessage(message).
		WithErrors(errors)
}

// Forbidden creates a forbidden response
func Forbidden(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("FORBIDDEN").
		WithMessage(message).
		WithErrors(errors)
}

// NotFound creates a not found response
func NotFound(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("NOT_FOUND").
		WithMessage(message).
		WithErrors(errors)
}

// Conflict creates a conflict response
func Conflict(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("CONFLICT").
		WithMessage(message).
		WithErrors(errors)
}

// ValidationError creates a validation error response
func ValidationError(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("BAD_REQUEST").
		WithMessage(message).
		WithErrors(errors)
}

// InternalServerError creates an internal server error response
func InternalServerError(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("INTERNAL_SERVER_ERROR").
		WithMessage(message).
		WithErrors(errors)
}

// Failed creates a generic failed response
func Failed(ctx context.Context, message string, errors []string) *ResponseBuilder {
	return New(ctx).
		WithCode("FAILED").
		WithMessage(message).
		WithErrors(errors)
}
