package router

import (
	"police-trafic-api-frontend-aligned/internal/core/interfaces"
	"police-trafic-api-frontend-aligned/internal/core/middleware"
	"police-trafic-api-frontend-aligned/internal/core/server"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"go.uber.org/zap"
)

// NewServer creates a new server with all controllers
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	authMiddleware *middleware.AuthMiddleware,
	controllers []interfaces.Controller,
) *server.Server {
	return server.NewServer(cfg, logger, authMiddleware, controllers...)
}