package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	AccountID int64  `json:"account_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	jwt.RegisteredClaims
}

// Service handles JWT operations
type Service struct {
	secretKey []byte
	expiresIn time.Duration
}

// NewService creates a new JWT service
func NewService(secretKey string, expiresIn time.Duration) *Service {
	return &Service{
		secretKey: []byte(secretKey),
		expiresIn: expiresIn,
	}
}

// GenerateToken creates a new JWT token for the given account
func (s *Service) GenerateToken(accountID int64, email, name string) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		Email:     email,
		Name:      name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "social-media-service",
			Subject:   fmt.Sprintf("%d", accountID),
			Audience:  []string{"social-media-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d-%d", accountID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates and parses a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetExpiresInSeconds returns the expiration time in seconds
func (s *Service) GetExpiresInSeconds() int64 {
	return int64(s.expiresIn.Seconds())
}
