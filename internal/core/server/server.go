package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	coremiddleware "police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// CustomValidator wraps go-playground/validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

type Server struct {
	echo           *echo.Echo
	config         *config.Config
	logger         *zap.Logger
	controllers    []interfaces.Controller
	authMiddleware *coremiddleware.AuthMiddleware
}

func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	authMiddleware *coremiddleware.AuthMiddleware,
	controllers ...interfaces.Controller,
) *Server {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	// Set up validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Add middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":    "healthy",
			"service":   cfg.App.Name,
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// Swagger documentation
	if cfg.App.Environment == "development" {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	return &Server{
		echo:           e,
		config:         cfg,
		logger:         logger,
		controllers:    controllers,
		authMiddleware: authMiddleware,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Register all controller routes under /api/v1 prefix
	// Apply JWT authentication middleware to /api/v1 group (except /api/v1/auth)
	api := s.echo.Group("/api/v1", s.authMiddleware.RequireAuthWithSkipper(func(path string) bool {
		// Skip authentication for all auth routes
		skip := strings.HasPrefix(path, "/api/v1/auth")
		s.logger.Debug("Auth middleware check", 
			zap.String("path", path),
			zap.Bool("skip", skip))
		return skip
	}))
	
	s.logger.Info("Registering controllers", zap.Int("count", len(s.controllers)))
	for _, controller := range s.controllers {
		controller.RegisterRoutes(api)
	}

	// Log registered routes
	s.logger.Info("All routes registered:")
	for _, route := range s.echo.Routes() {
		s.logger.Info("Registered route",
			zap.String("method", route.Method),
			zap.String("path", route.Path),
			zap.String("name", route.Name),
		)
	}

	address := fmt.Sprintf(":%s", s.config.Server.Port)
	s.logger.Info("🚀 Server starting",
		zap.String("address", address),
		zap.String("environment", s.config.App.Environment),
		zap.Bool("debug", s.config.App.Debug),
	)

	s.logger.Info("📋 Health check available at: http://localhost" + address + "/health")
	if s.config.App.Environment == "development" {
		s.logger.Info("📚 Swagger docs available at: http://localhost" + address + "/swagger/index.html")
	}

	// Start server
	return s.echo.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("🛑 Shutting down server...")
	
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.Server.ShutdownTimeout)
	defer cancel()
	
	return s.echo.Shutdown(shutdownCtx)
}