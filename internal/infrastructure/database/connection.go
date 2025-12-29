package database

import (
	"context"
	"fmt"
	"time"

	"police-trafic-api-frontend-aligned/ent"
	"police-trafic-api-frontend-aligned/internal/infrastructure/config"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// DB represents a database connection using Ent
type DB struct {
	Client *ent.Client
	logger *zap.Logger
	config *config.DatabaseConfig
}

// NewDB creates a new Ent database connection
func NewDB(cfg *config.Config, logger *zap.Logger) (*DB, error) {
	logger.Info("Initializing PostgreSQL Database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("dbname", cfg.Database.DBName),
	)

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.DBName,
	)

	// Add password if provided
	if cfg.Database.Password != "" {
		dsn += fmt.Sprintf(" password=%s", cfg.Database.Password)
	}

	logger.Debug("Connecting to database", zap.String("dsn", dsn))

	// Open database connection
	drv, err := sql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	// Configure connection pool
	db := drv.DB()
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Create Ent client
	client := ent.NewClient(ent.Driver(drv))

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Schema.Create(ctx); err != nil {
		logger.Warn("Failed to create schema", zap.Error(err))
		client.Close()
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	logger.Info("PostgreSQL Database connected successfully")

	return &DB{
		Client: client,
		logger: logger,
		config: &cfg.Database,
	}, nil
}

// MockDB represents a mock database connection for development
type MockDB struct {
	logger *zap.Logger
	config *config.DatabaseConfig
}

// NewMockDB creates a new mock database connection (for fallback)
func NewMockDB(cfg *config.Config, logger *zap.Logger) (*MockDB, error) {
	logger.Info("Initializing Mock Database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("dbname", cfg.Database.DBName),
	)

	db := &MockDB{
		logger: logger,
		config: &cfg.Database,
	}

	logger.Info("Mock Database connected successfully")
	return db, nil
}

// Ping tests the database connection
func (db *DB) Ping() error {
	ctx := context.Background()
	return db.Client.Schema.Create(ctx)
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.Client.Close()
}

// GetConnectionInfo returns connection information
func (db *DB) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"driver":        "postgres",
		"host":          db.config.Host,
		"port":          db.config.Port,
		"database":      db.config.DBName,
		"max_open":      db.config.MaxOpenConns,
		"max_idle":      db.config.MaxIdleConns,
		"status":        "connected",
	}
}

// Ping simulates a database ping for mock
func (db *MockDB) Ping() error {
	db.logger.Debug("Database ping successful")
	return nil
}

// Close simulates closing the database connection for mock
func (db *MockDB) Close() error {
	db.logger.Info("Closing mock database connection")
	return nil
}

// GetConnectionInfo returns mock connection information
func (db *MockDB) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"driver":        "mock",
		"host":          db.config.Host,
		"port":          db.config.Port,
		"database":      db.config.DBName,
		"max_open":      db.config.MaxOpenConns,
		"max_idle":      db.config.MaxIdleConns,
		"status":        "connected",
	}
}


