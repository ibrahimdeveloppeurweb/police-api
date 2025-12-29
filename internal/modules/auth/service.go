package auth

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/internal/infrastructure/crypto"
	"police-trafic-api-frontend-aligned/internal/infrastructure/jwt"
	"police-trafic-api-frontend-aligned/internal/infrastructure/repository"
	"police-trafic-api-frontend-aligned/internal/infrastructure/session"
	"police-trafic-api-frontend-aligned/internal/shared/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service interface defines auth service methods
type Service interface {
	Login(req LoginRequest, ipAddress string) (*LoginResponse, error)
	GetCurrentUser(token string) (*User, error)
	Logout(req LogoutRequest, token string) error
	RefreshToken(req RefreshTokenRequest, ipAddress string) (*LoginResponse, error)
	RefreshTokenLegacy(token string) (*LoginResponse, error) // Backwards compatible
	Register(req RegisterRequest) (*User, error)
	GetUserSessions(token string) ([]SessionDTO, error)
	RevokeSession(token string, sessionID string) error
}

type service struct {
	logger         *zap.Logger
	userRepo       repository.UserRepository
	jwtService     jwt.Service
	cryptoService  crypto.Service
	sessionService session.Service
}

// NewService creates a new auth service
func NewService(
	logger *zap.Logger,
	userRepo repository.UserRepository,
	jwtService jwt.Service,
	cryptoService crypto.Service,
	sessionService session.Service,
) Service {
	return &service{
		logger:         logger,
		userRepo:       userRepo,
		jwtService:     jwtService,
		cryptoService:  cryptoService,
		sessionService: sessionService,
	}
}

func (s *service) Login(req LoginRequest, ipAddress string) (*LoginResponse, error) {
	identifier := req.GetIdentifier()
	s.logger.Info("Login attempt", zap.String("matricule", identifier))

	ctx := context.Background()
	user, err := s.userRepo.GetByMatricule(ctx, identifier)
	if err != nil {
		s.logger.Warn("User not found", zap.String("matricule", identifier), zap.Error(err))
		// Fallback to mock data for testing
		return s.loginWithMockData(identifier, req.Password, req.Device, ipAddress)
	}

	// Verify password with bcrypt
	if err := s.cryptoService.CheckPassword(req.Password, user.Password); err != nil {
		s.logger.Warn("Password verification failed",
			zap.String("matricule", identifier),
			zap.Error(err),
		)
		return nil, errors.ErrInvalidCredentials
	}

	if !user.Active {
		s.logger.Warn("User account inactive", zap.String("matricule", identifier))
		return nil, errors.ErrUnauthorized
	}

	// Build user response
	userResp := User{
		ID:        user.ID.String(),
		Matricule: user.Matricule,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Email:     user.Email,
		Role:      user.Role,
		Grade:     user.Grade,
		Telephone: user.Telephone,
		Statut:    user.StatutService,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Add commissariat info if available
	if user.Edges.Commissariat != nil {
		userResp.CommissariatID = user.Edges.Commissariat.ID.String()
		userResp.Commissariat = user.Edges.Commissariat.Nom
	}

	// If device info provided, create a session (mobile app)
	if req.Device != nil && req.Device.DeviceID != "" {
		return s.loginWithSession(ctx, user.ID, userResp, req.Device, ipAddress)
	}

	// Standard login without session (web app - backwards compatible)
	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Matricule, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  userResp,
	}, nil
}

func (s *service) loginWithSession(ctx context.Context, userID uuid.UUID, userResp User, device *DeviceInfo, ipAddress string) (*LoginResponse, error) {
	// Create session
	deviceInfo := session.DeviceInfo{
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		DeviceType: device.DeviceType,
		DeviceOS:   device.DeviceOS,
		AppVersion: device.AppVersion,
	}

	sessionInfo, err := s.sessionService.CreateSession(ctx, userID, deviceInfo, ipAddress)
	if err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate JWT with session binding
	token, err := s.jwtService.GenerateTokenWithSession(
		userID.String(),
		userResp.Matricule,
		userResp.Role,
		sessionInfo.SessionID,
		device.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: sessionInfo.RefreshToken,
		SessionID:    sessionInfo.SessionID,
		ExpiresAt:    &sessionInfo.ExpiresAt,
		User:         userResp,
	}, nil
}

func (s *service) loginWithMockData(matricule, password string, device *DeviceInfo, ipAddress string) (*LoginResponse, error) {
	users := s.getMockUsers()
	for _, user := range users {
		if user.Matricule == matricule {
			if password != "" {
				// If device info provided, create session
				if device != nil && device.DeviceID != "" {
					// For mock users, generate a mock session
					ctx := context.Background()
					mockUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
					return s.loginWithSession(ctx, mockUserID, user, device, ipAddress)
				}

				token, err := s.jwtService.GenerateToken(user.ID, user.Matricule, user.Role)
				if err != nil {
					return nil, fmt.Errorf("failed to generate token: %w", err)
				}
				return &LoginResponse{Token: token, User: user}, nil
			}
			return nil, errors.ErrInvalidCredentials
		}
	}
	return nil, errors.ErrInvalidCredentials
}

func (s *service) GetCurrentUser(token string) (*User, error) {
	s.logger.Info("Getting current user from token")

	// Validate JWT token
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	// If session is bound, validate it
	if claims.SessionID != "" {
		ctx := context.Background()
		sessionID, err := uuid.Parse(claims.SessionID)
		if err == nil {
			valid, _ := s.sessionService.IsSessionValid(ctx, sessionID, claims.DeviceID)
			if !valid {
				s.logger.Warn("Session invalid or revoked",
					zap.String("session_id", claims.SessionID),
					zap.String("user_id", claims.UserID),
				)
				return nil, errors.ErrUnauthorized
			}
		}
	}

	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		s.logger.Warn("User not found in database", zap.String("user_id", claims.UserID))
		// Fallback to mock data
		users := s.getMockUsers()
		for _, u := range users {
			if u.ID == claims.UserID {
				return &u, nil
			}
		}
		return nil, errors.ErrNotFound
	}

	result := &User{
		ID:        user.ID.String(),
		Matricule: user.Matricule,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Email:     user.Email,
		Role:      user.Role,
		Grade:     user.Grade,
		Telephone: user.Telephone,
		Statut:    user.StatutService,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Add commissariat info if available
	if user.Edges.Commissariat != nil {
		result.CommissariatID = user.Edges.Commissariat.ID.String()
		result.Commissariat = user.Edges.Commissariat.Nom
	}

	return result, nil
}

func (s *service) Logout(req LogoutRequest, token string) error {
	s.logger.Info("User logout")

	// Validate token
	claims, err := s.jwtService.ValidateTokenIgnoreExpiry(token)
	if err != nil {
		return errors.ErrInvalidToken
	}

	ctx := context.Background()
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return errors.ErrInvalidToken
	}

	// Logout from all devices
	if req.AllDevices {
		if err := s.sessionService.RevokeAllUserSessions(ctx, userID, "logout_all_devices"); err != nil {
			s.logger.Error("Failed to revoke all sessions", zap.Error(err))
		}
		s.logger.Info("User logged out from all devices", zap.String("user_id", claims.UserID))
		return nil
	}

	// Logout specific session
	if req.SessionID != "" {
		sessionID, err := uuid.Parse(req.SessionID)
		if err != nil {
			return fmt.Errorf("invalid session ID")
		}
		if err := s.sessionService.RevokeSession(ctx, sessionID, "user_logout"); err != nil {
			s.logger.Error("Failed to revoke session", zap.Error(err))
		}
		s.logger.Info("Session revoked", zap.String("session_id", req.SessionID))
		return nil
	}

	// If session bound to token, revoke it
	if claims.SessionID != "" {
		sessionID, err := uuid.Parse(claims.SessionID)
		if err == nil {
			if err := s.sessionService.RevokeSession(ctx, sessionID, "user_logout"); err != nil {
				s.logger.Error("Failed to revoke session", zap.Error(err))
			}
		}
	}

	s.logger.Info("User logged out successfully")
	return nil
}

func (s *service) RefreshToken(req RefreshTokenRequest, ipAddress string) (*LoginResponse, error) {
	s.logger.Info("Refreshing token with refresh token")

	ctx := context.Background()

	// Validate refresh token
	entSession, err := s.sessionService.ValidateRefreshToken(ctx, req.RefreshToken, req.DeviceID)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user
	user := entSession.Edges.User
	if user == nil {
		return nil, errors.ErrNotFound
	}

	// Refresh the session (generates new refresh token)
	newSession, err := s.sessionService.RefreshSession(ctx, entSession.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	// Update last activity
	_ = s.sessionService.UpdateLastActivity(ctx, entSession.ID, ipAddress)

	// Generate new JWT with session binding
	token, err := s.jwtService.GenerateTokenWithSession(
		user.ID.String(),
		user.Matricule,
		user.Role,
		newSession.SessionID,
		req.DeviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:        token,
		RefreshToken: newSession.RefreshToken,
		SessionID:    newSession.SessionID,
		ExpiresAt:    &newSession.ExpiresAt,
		User: User{
			ID:        user.ID.String(),
			Matricule: user.Matricule,
			Nom:       user.Nom,
			Prenom:    user.Prenom,
			Email:     user.Email,
			Role:      user.Role,
			Grade:     user.Grade,
			Telephone: user.Telephone,
			Statut:    user.StatutService,
			Active:    user.Active,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RefreshTokenLegacy refreshes using the access token (backwards compatible for web)
func (s *service) RefreshTokenLegacy(token string) (*LoginResponse, error) {
	s.logger.Info("Refreshing token (legacy)")

	// Use JWT service to refresh token
	newToken, err := s.jwtService.RefreshToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user info from new token
	user, err := s.GetCurrentUser(newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &LoginResponse{Token: newToken, User: *user}, nil
}

func (s *service) GetUserSessions(token string) ([]SessionDTO, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	ctx := context.Background()
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	sessions, err := s.sessionService.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	result := make([]SessionDTO, len(sessions))
	for i, sess := range sessions {
		result[i] = SessionDTO{
			ID:             sess.ID.String(),
			DeviceID:       sess.DeviceID,
			DeviceName:     sess.DeviceName,
			DeviceType:     sess.DeviceType,
			DeviceOS:       sess.DeviceOs,
			AppVersion:     sess.AppVersion,
			LastActivityAt: sess.LastActivityAt,
			LastIPAddress:  sess.LastIPAddress,
			IsCurrent:      claims.SessionID == sess.ID.String(),
			CreatedAt:      sess.SessionStartedAt,
		}
	}

	return result, nil
}

func (s *service) RevokeSession(token string, sessionID string) error {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return errors.ErrInvalidToken
	}

	ctx := context.Background()
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return errors.ErrInvalidToken
	}

	targetSessionID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID")
	}

	// Verify the session belongs to this user
	sessions, err := s.sessionService.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	found := false
	for _, sess := range sessions {
		if sess.ID == targetSessionID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("session not found")
	}

	return s.sessionService.RevokeSession(ctx, targetSessionID, "user_revoked")
}

func (s *service) Register(req RegisterRequest) (*User, error) {
	s.logger.Info("User registration attempt", zap.String("matricule", req.Matricule))

	// Hash password
	hashedPassword, err := s.cryptoService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	ctx := context.Background()

	// Check if user already exists
	if _, err := s.userRepo.GetByMatricule(ctx, req.Matricule); err == nil {
		return nil, fmt.Errorf("user with matricule %s already exists", req.Matricule)
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Create user
	userInput := &repository.CreateUserInput{
		ID:        fmt.Sprintf("%d", time.Now().Unix()), // Simple ID generation
		Matricule: req.Matricule,
		Nom:       req.Nom,
		Prenom:    req.Prenom,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
	}

	user, err := s.userRepo.Create(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:        user.ID.String(),
		Matricule: user.Matricule,
		Nom:       user.Nom,
		Prenom:    user.Prenom,
		Email:     user.Email,
		Role:      user.Role,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// getMockUsers returns mock users for testing
func (s *service) getMockUsers() []User {
	// Utiliser le commissariat du 7ème Arrondissement qui a des alertes
	commissariatID := "566f69ab-8146-44ed-bea2-2fb251523a24"
	commissariatNom := "Commissariat du 7ème Arrondissement"
	
	return []User{
		{
			ID:             "1",
			Matricule:      "12345",
			Nom:            "Dupont",
			Prenom:         "Jean",
			Email:          "j.dupont@police.gouv.fr",
			Role:           "agent",
			Active:         true,
			CommissariatID: commissariatID,
			Commissariat:   commissariatNom,
			CreatedAt:      time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "2",
			Matricule:      "67890",
			Nom:            "Martin",
			Prenom:         "Marie",
			Email:          "m.martin@police.gouv.fr",
			Role:           "admin",
			Active:         true,
			CommissariatID: commissariatID,
			Commissariat:   commissariatNom,
			CreatedAt:      time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "3",
			Matricule:      "11111",
			Nom:            "Durand",
			Prenom:         "Pierre",
			Email:          "p.durand@police.gouv.fr",
			Role:           "supervisor",
			CommissariatID: commissariatID,
			Commissariat:   commissariatNom,
			Active:    true,
			CreatedAt: time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}
}
