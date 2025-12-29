package database

import (
	"context"
	"fmt"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// EntClientProvider provides *ent.Client for repositories
func EntClientProvider(cfg *config.Config, logger *zap.Logger) (*ent.Client, error) {
	// Try to connect to PostgreSQL
	db, err := NewDB(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db.Client, nil
}

// DatabaseProvider provides database wrapper for lifecycle management
func DatabaseProvider(cfg *config.Config, logger *zap.Logger) (*DB, error) {
	return NewDB(cfg, logger)
}

// Module provides database dependency
var Module = fx.Module("database",
	fx.Provide(DatabaseProvider),
	fx.Provide(func(db *DB) *ent.Client {
		return db.Client
	}),
	fx.Invoke(func(lc fx.Lifecycle, db *DB) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return db.Ping()
			},
			OnStop: func(ctx context.Context) error {
				return db.Close()
			},
		})
	}),
)


