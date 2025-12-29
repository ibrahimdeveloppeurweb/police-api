package crypto

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service defines password hashing service interface
type Service interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) error
}

// service implements password hashing service
type service struct {
	logger *zap.Logger
	cost   int
}

// NewPasswordService creates a new password service
func NewPasswordService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
		cost:   bcrypt.DefaultCost, // Cost 10 by default
	}
}

// HashPassword hashes a password using bcrypt
func (s *service) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	s.logger.Debug("Password hashed successfully")
	return string(hash), nil
}

// CheckPassword verifies a password against its hash
func (s *service) CheckPassword(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			s.logger.Debug("Password verification failed: mismatch")
			return fmt.Errorf("invalid password")
		}
		s.logger.Error("Password verification failed", zap.Error(err))
		return fmt.Errorf("password verification error: %w", err)
	}

	s.logger.Debug("Password verified successfully")
	return nil
}