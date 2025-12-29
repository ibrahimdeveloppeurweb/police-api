package jwt

import (
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	Matricule string `json:"matricule"`
	Role      string `json:"role"`
	SessionID string `json:"session_id,omitempty"` // Session ID for revocation check
	DeviceID  string `json:"device_id,omitempty"`  // Device ID for device binding
	jwt.RegisteredClaims
}

// Service defines JWT service interface
type Service interface {
	GenerateToken(userID, matricule, role string) (string, error)
	GenerateTokenWithSession(userID, matricule, role, sessionID, deviceID string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ValidateTokenIgnoreExpiry(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, error)
}

// service implements JWT service
type service struct {
	config *config.JWTConfig
	logger *zap.Logger
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.Config, logger *zap.Logger) Service {
	return &service{
		config: &cfg.JWT,
		logger: logger,
	}
}

// GenerateToken generates a JWT token for a user (without session binding - for backwards compatibility)
func (s *service) GenerateToken(userID, matricule, role string) (string, error) {
	return s.GenerateTokenWithSession(userID, matricule, role, "", "")
}

// GenerateTokenWithSession generates a JWT token with session and device binding
func (s *service) GenerateTokenWithSession(userID, matricule, role, sessionID, deviceID string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Matricule: matricule,
		Role:      role,
		SessionID: sessionID,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "police-traffic-api",
			Subject:   userID,
			Audience:  []string{"police-traffic-frontend"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Debug("Token generated successfully",
		zap.String("user_id", userID),
		zap.String("matricule", matricule),
		zap.String("role", role),
		zap.String("session_id", sessionID),
		zap.String("device_id", deviceID),
	)

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		s.logger.Warn("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		s.logger.Warn("Token claims invalid")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("Token expired", 
			zap.Time("expired_at", claims.ExpiresAt.Time),
			zap.String("user_id", claims.UserID),
		)
		return nil, fmt.Errorf("token expired")
	}

	s.logger.Debug("Token validated successfully", 
		zap.String("user_id", claims.UserID),
		zap.String("matricule", claims.Matricule),
		zap.String("role", claims.Role),
	)

	return claims, nil
}

// ValidateTokenIgnoreExpiry validates a JWT token but ignores expiration
// This is useful for refresh token scenarios where the token may be expired
func (s *service) ValidateTokenIgnoreExpiry(tokenString string) (*Claims, error) {
	// Parse with options to skip expiration check
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		s.logger.Warn("Token parsing failed (ignore expiry)", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		s.logger.Warn("Token claims invalid (ignore expiry)")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify signature is valid (token.Valid checks signature)
	// We need to re-verify the signature since we skipped validation
	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Secret), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		s.logger.Warn("Token signature invalid", zap.Error(err))
		return nil, fmt.Errorf("invalid token signature: %w", err)
	}

	s.logger.Debug("Token validated (ignoring expiry)",
		zap.String("user_id", claims.UserID),
		zap.String("matricule", claims.Matricule),
		zap.String("role", claims.Role),
	)

	return claims, nil
}

// RefreshToken generates a new token from an existing token (can be expired)
func (s *service) RefreshToken(tokenString string) (string, error) {
	// Use ValidateTokenIgnoreExpiry to allow refreshing expired tokens
	claims, err := s.ValidateTokenIgnoreExpiry(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate new token with same user info but new expiration
	return s.GenerateToken(claims.UserID, claims.Matricule, claims.Role)
}