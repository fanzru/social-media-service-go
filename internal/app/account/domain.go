package account

import (
	"time"
)

// Account represents the account domain model
type Account struct {
	ID        int64      `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"` // Hidden from JSON response
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// RegisterRequest represents the request payload for account registration
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the request payload for account login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response payload for successful login
type LoginResponse struct {
	Account     Account `json:"account"`
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"`
	ExpiresIn   int64   `json:"expires_in"` // seconds
}

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Errors     []string    `json:"errors,omitempty"`
	ServerTime string      `json:"serverTime"`
	RequestID  string      `json:"requestId"`
	Data       interface{} `json:"data,omitempty"`
}
