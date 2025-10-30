package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fanzru/social-media-service-go/internal/app/account"
	"github.com/fanzru/social-media-service-go/internal/app/account/repo"
	"github.com/fanzru/social-media-service-go/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Service interface defines the contract for account business logic
type Service interface {
	Register(ctx context.Context, req *account.RegisterRequest) (*account.Account, error)
	Login(ctx context.Context, req *account.LoginRequest) (*account.LoginResponse, error)
	GetAccountByID(ctx context.Context, id int64) (*account.Account, error)
	UpdateAccount(ctx context.Context, acc *account.Account) error
	DeleteAccount(ctx context.Context, id int64) error
	// GDPRDeleteAccount permanently deletes the account and all associated data
	GDPRDeleteAccount(ctx context.Context, id int64) error
}

// service implements the Service interface
type service struct {
	repo       repo.Repository
	jwtService *jwt.Service
	imageStore ImageDeleter
}

// ImageDeleter defines the capability needed to delete images
type ImageDeleter interface {
	DeleteImage(imagePath string) error
}

// NewService creates a new account service
func NewService(repo repo.Repository, jwtService *jwt.Service, imageStore ImageDeleter) Service {
	return &service{
		repo:       repo,
		jwtService: jwtService,
		imageStore: imageStore,
	}
}

// Register creates a new account
func (s *service) Register(ctx context.Context, req *account.RegisterRequest) (*account.Account, error) {
	// Check if email already exists
	existingAccount, err := s.repo.GetByEmail(ctx, req.Email)
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

	err = s.repo.Create(ctx, acc)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return acc, nil
}

// Login authenticates a user
func (s *service) Login(ctx context.Context, req *account.LoginRequest) (*account.LoginResponse, error) {
	// Get account by email
	acc, err := s.repo.GetByEmail(ctx, req.Email)
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
func (s *service) GetAccountByID(ctx context.Context, id int64) (*account.Account, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateAccount updates an existing account
func (s *service) UpdateAccount(ctx context.Context, acc *account.Account) error {
	return s.repo.Update(ctx, acc)
}

// DeleteAccount soft deletes an account
func (s *service) DeleteAccount(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

// GDPRDeleteAccount permanently deletes an account and cleans up user images
func (s *service) GDPRDeleteAccount(ctx context.Context, id int64) error {

	var err error
	// Begin DB transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Gather image paths within the transaction scope
	imagePaths, err := s.repo.ListUserPostImagePathsTx(ctx, tx, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to list user's post images: %w", err)
	}

	// Try deleting images first; if any fails, rollback to keep DB unchanged
	for _, path := range imagePaths {
		if err := s.imageStore.DeleteImage(path); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to delete image '%s': %w", path, err)
		}
	}

	// Delete account within the same transaction (CASCADE removes posts/comments)
	if err := s.repo.DeleteTx(ctx, tx, id); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete account: %w", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
