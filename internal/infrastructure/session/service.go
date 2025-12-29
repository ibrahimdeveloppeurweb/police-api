package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/ent/usersession"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeviceInfo contains device information from the mobile app
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // ios, android, web
	DeviceOS   string `json:"device_os,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

// SessionInfo contains session data returned to the client
type SessionInfo struct {
	SessionID    string    `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Service defines the session management interface
type Service interface {
	// CreateSession creates a new session for a user on a device
	CreateSession(ctx context.Context, userID uuid.UUID, device DeviceInfo, ipAddress string) (*SessionInfo, error)

	// ValidateRefreshToken validates a refresh token and returns the session
	ValidateRefreshToken(ctx context.Context, refreshToken string, deviceID string) (*ent.UserSession, error)

	// RefreshSession refreshes a session (updates activity, generates new refresh token)
	RefreshSession(ctx context.Context, sessionID uuid.UUID) (*SessionInfo, error)

	// RevokeSession revokes a specific session
	RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error

	// RevokeAllUserSessions revokes all sessions for a user
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error

	// RevokeOtherSessions revokes all sessions except the current one
	RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID, reason string) error

	// GetUserSessions returns all active sessions for a user
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*ent.UserSession, error)

	// CleanupExpiredSessions removes expired sessions from the database
	CleanupExpiredSessions(ctx context.Context) (int, error)

	// IsSessionValid checks if a session is valid (not revoked, not expired, within max duration)
	IsSessionValid(ctx context.Context, sessionID uuid.UUID, deviceID string) (bool, error)

	// UpdateLastActivity updates the last activity timestamp for a session
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID, ipAddress string) error
}

type service struct {
	client *ent.Client
	config *config.JWTConfig
	logger *zap.Logger
}

// NewService creates a new session service
func NewService(client *ent.Client, cfg *config.Config, logger *zap.Logger) Service {
	// Set defaults if not configured
	if cfg.JWT.MaxSessionDuration == 0 {
		cfg.JWT.MaxSessionDuration = 30 * 24 * time.Hour // 30 days default
	}
	if cfg.JWT.MaxDevicesPerUser == 0 {
		cfg.JWT.MaxDevicesPerUser = 5 // 5 devices default
	}
	if cfg.JWT.InactivityTimeout == 0 {
		cfg.JWT.InactivityTimeout = 7 * 24 * time.Hour // 7 days default
	}
	if cfg.JWT.RefreshExpiration == 0 {
		cfg.JWT.RefreshExpiration = 7 * 24 * time.Hour // 7 days default
	}

	return &service{
		client: client,
		config: &cfg.JWT,
		logger: logger,
	}
}

// CreateSession creates a new session for a user
func (s *service) CreateSession(ctx context.Context, userID uuid.UUID, device DeviceInfo, ipAddress string) (*SessionInfo, error) {
	// Check if user already has a session on this device
	existingSession, err := s.client.UserSession.Query().
		Where(
			usersession.HasUserWith(),
			usersession.DeviceID(device.DeviceID),
			usersession.IsActive(true),
			usersession.IsRevoked(false),
		).
		First(ctx)

	if err == nil && existingSession != nil {
		// Revoke the existing session on this device
		s.logger.Info("Revoking existing session on device",
			zap.String("device_id", device.DeviceID),
			zap.String("session_id", existingSession.ID.String()),
		)
		if err := s.RevokeSession(ctx, existingSession.ID, "new_login_same_device"); err != nil {
			s.logger.Warn("Failed to revoke existing session", zap.Error(err))
		}
	}

	// Check max devices limit
	activeSessions, err := s.client.UserSession.Query().
		Where(
			usersession.HasUserWith(),
			usersession.IsActive(true),
			usersession.IsRevoked(false),
		).
		Count(ctx)

	if err != nil {
		s.logger.Error("Failed to count active sessions", zap.Error(err))
	} else if activeSessions >= s.config.MaxDevicesPerUser {
		// Revoke oldest session
		oldestSession, err := s.client.UserSession.Query().
			Where(
				usersession.HasUserWith(),
				usersession.IsActive(true),
				usersession.IsRevoked(false),
			).
			Order(ent.Asc(usersession.FieldLastActivityAt)).
			First(ctx)

		if err == nil && oldestSession != nil {
			s.logger.Info("Revoking oldest session due to max devices limit",
				zap.String("session_id", oldestSession.ID.String()),
			)
			if err := s.RevokeSession(ctx, oldestSession.ID, "max_sessions"); err != nil {
				s.logger.Warn("Failed to revoke oldest session", zap.Error(err))
			}
		}
	}

	// Generate refresh token
	refreshToken, err := generateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshTokenHash := hashToken(refreshToken)
	expiresAt := time.Now().Add(s.config.RefreshExpiration)

	// Create session
	session, err := s.client.UserSession.Create().
		SetUserID(userID).
		SetDeviceID(device.DeviceID).
		SetNillableDeviceName(nilIfEmpty(device.DeviceName)).
		SetNillableDeviceType(nilIfEmpty(device.DeviceType)).
		SetNillableDeviceOs(nilIfEmpty(device.DeviceOS)).
		SetNillableAppVersion(nilIfEmpty(device.AppVersion)).
		SetRefreshTokenHash(refreshTokenHash).
		SetRefreshTokenExpiresAt(expiresAt).
		SetSessionStartedAt(time.Now()).
		SetLastActivityAt(time.Now()).
		SetNillableLastIPAddress(nilIfEmpty(ipAddress)).
		SetIsActive(true).
		SetIsRevoked(false).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	s.logger.Info("Session created",
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("device_id", device.DeviceID),
	)

	return &SessionInfo{
		SessionID:    session.ID.String(),
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateRefreshToken validates a refresh token
func (s *service) ValidateRefreshToken(ctx context.Context, refreshToken string, deviceID string) (*ent.UserSession, error) {
	tokenHash := hashToken(refreshToken)

	session, err := s.client.UserSession.Query().
		Where(
			usersession.RefreshTokenHash(tokenHash),
			usersession.IsActive(true),
			usersession.IsRevoked(false),
		).
		WithUser().
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("invalid refresh token")
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Check device binding
	if session.DeviceID != deviceID {
		s.logger.Warn("Device ID mismatch for refresh token",
			zap.String("expected", session.DeviceID),
			zap.String("received", deviceID),
		)
		// Revoke the session - potential token theft
		_ = s.RevokeSession(ctx, session.ID, "device_mismatch")
		return nil, fmt.Errorf("device mismatch - session revoked for security")
	}

	// Check if refresh token is expired
	if time.Now().After(session.RefreshTokenExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// Check max session duration
	if time.Now().After(session.SessionStartedAt.Add(s.config.MaxSessionDuration)) {
		_ = s.RevokeSession(ctx, session.ID, "max_session_duration")
		return nil, fmt.Errorf("session expired - please login again")
	}

	// Check inactivity timeout
	if time.Now().After(session.LastActivityAt.Add(s.config.InactivityTimeout)) {
		_ = s.RevokeSession(ctx, session.ID, "inactivity")
		return nil, fmt.Errorf("session expired due to inactivity - please login again")
	}

	return session, nil
}

// RefreshSession refreshes a session with a new refresh token
func (s *service) RefreshSession(ctx context.Context, sessionID uuid.UUID) (*SessionInfo, error) {
	// Generate new refresh token
	refreshToken, err := generateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshTokenHash := hashToken(refreshToken)
	expiresAt := time.Now().Add(s.config.RefreshExpiration)

	// Update session
	session, err := s.client.UserSession.UpdateOneID(sessionID).
		SetRefreshTokenHash(refreshTokenHash).
		SetRefreshTokenExpiresAt(expiresAt).
		SetLastActivityAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	return &SessionInfo{
		SessionID:    session.ID.String(),
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// RevokeSession revokes a specific session
func (s *service) RevokeSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	now := time.Now()
	_, err := s.client.UserSession.UpdateOneID(sessionID).
		SetIsActive(false).
		SetIsRevoked(true).
		SetRevokedAt(now).
		SetRevokedReason(reason).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	s.logger.Info("Session revoked",
		zap.String("session_id", sessionID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *service) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	now := time.Now()
	_, err := s.client.UserSession.Update().
		Where(
			usersession.HasUserWith(),
			usersession.IsActive(true),
		).
		SetIsActive(false).
		SetIsRevoked(true).
		SetRevokedAt(now).
		SetRevokedReason(reason).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}

	s.logger.Info("All user sessions revoked",
		zap.String("user_id", userID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// RevokeOtherSessions revokes all sessions except the current one
func (s *service) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID, reason string) error {
	now := time.Now()
	_, err := s.client.UserSession.Update().
		Where(
			usersession.HasUserWith(),
			usersession.IsActive(true),
			usersession.IDNEQ(currentSessionID),
		).
		SetIsActive(false).
		SetIsRevoked(true).
		SetRevokedAt(now).
		SetRevokedReason(reason).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke other sessions: %w", err)
	}

	s.logger.Info("Other user sessions revoked",
		zap.String("user_id", userID.String()),
		zap.String("current_session_id", currentSessionID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *service) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*ent.UserSession, error) {
	sessions, err := s.client.UserSession.Query().
		Where(
			usersession.HasUserWith(),
			usersession.IsActive(true),
			usersession.IsRevoked(false),
		).
		Order(ent.Desc(usersession.FieldLastActivityAt)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *service) CleanupExpiredSessions(ctx context.Context) (int, error) {
	now := time.Now()
	maxAge := now.Add(-s.config.MaxSessionDuration)

	deleted, err := s.client.UserSession.Delete().
		Where(
			usersession.Or(
				usersession.RefreshTokenExpiresAtLT(now),
				usersession.SessionStartedAtLT(maxAge),
				usersession.IsRevoked(true),
			),
		).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	if deleted > 0 {
		s.logger.Info("Cleaned up expired sessions", zap.Int("count", deleted))
	}

	return deleted, nil
}

// IsSessionValid checks if a session is valid
func (s *service) IsSessionValid(ctx context.Context, sessionID uuid.UUID, deviceID string) (bool, error) {
	session, err := s.client.UserSession.Query().
		Where(
			usersession.ID(sessionID),
			usersession.IsActive(true),
			usersession.IsRevoked(false),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to query session: %w", err)
	}

	// Check device binding
	if session.DeviceID != deviceID {
		return false, nil
	}

	// Check max session duration
	if time.Now().After(session.SessionStartedAt.Add(s.config.MaxSessionDuration)) {
		return false, nil
	}

	// Check inactivity
	if time.Now().After(session.LastActivityAt.Add(s.config.InactivityTimeout)) {
		return false, nil
	}

	return true, nil
}

// UpdateLastActivity updates the last activity timestamp
func (s *service) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID, ipAddress string) error {
	update := s.client.UserSession.UpdateOneID(sessionID).
		SetLastActivityAt(time.Now())

	if ipAddress != "" {
		update = update.SetLastIPAddress(ipAddress)
	}

	_, err := update.Save(ctx)
	return err
}

// Helper functions

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
