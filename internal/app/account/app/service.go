package app

import (
	"database/sql"
	"fmt"

	"github.com/fanzru/social-media-service-go/internal/app/account"
	"github.com/fanzru/social-media-service-go/internal/app/account/repo"
	"github.com/fanzru/social-media-service-go/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Service interface defines the contract for account business logic
type Service interface {
	Register(req *account.RegisterRequest) (*account.Account, error)
	Login(req *account.LoginRequest) (*account.LoginResponse, error)
	GetAccountByID(id int64) (*account.Account, error)
	UpdateAccount(acc *account.Account) error
	DeleteAccount(id int64) error
}

// service implements the Service interface
type service struct {
	repo      repo.Repository
	jwtService *jwt.Service
}

// NewService creates a new account service
func NewService(repo repo.Repository, jwtService *jwt.Service) Service {
	return &service{
		repo:      repo,
		jwtService: jwtService,
	}
}

// Register creates a new account
func (s *service) Register(req *account.RegisterRequest) (*account.Account, error) {
	// Check if email already exists
	existingAccount, err := s.repo.GetByEmail(req.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingAccount != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create account
	acc := &account.Account{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = s.repo.Create(acc)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return acc, nil
}

// Login authenticates a user
func (s *service) Login(req *account.LoginRequest) (*account.LoginResponse, error) {
	// Get account by email
	acc, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	accessToken, err := s.jwtService.GenerateToken(acc.ID, acc.Email, acc.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &account.LoginResponse{
		Account:     *acc,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtService.GetExpiresInSeconds(),
	}, nil
}

// GetAccountByID retrieves an account by ID
func (s *service) GetAccountByID(id int64) (*account.Account, error) {
	return s.repo.GetByID(id)
}

// UpdateAccount updates an existing account
func (s *service) UpdateAccount(acc *account.Account) error {
	return s.repo.Update(acc)
}

// DeleteAccount soft deletes an account
func (s *service) DeleteAccount(id int64) error {
	return s.repo.SoftDelete(id)
}

