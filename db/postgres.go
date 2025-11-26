package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"ridash/utils/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// InitDatabase initialize the database connection pool and return the pool and also migrate the database
func InitDatabase(logger echo.Logger) (*pgxpool.Pool, error) {
	ctx := context.Background()

	// Configure connection pool to handle concurrent operations better
	config, err := pgxpool.ParseConfig(getDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Increase pool size to handle more concurrent connections
	config.MaxConns = 25
	config.MinConns = 5

	// Reduce prepared statement cache to prevent "conn busy" errors
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info("Database initialized")

	Migrator(logger)

	return pool, nil
}

// getDatabaseURL return a pgsql connection uri by the environment variables
func getDatabaseURL() string {
	dbHost := config.Env().DBHost
	dbPort := config.Env().DBPort
	dbUser := config.Env().DBUser
	dbPassword := config.Env().DBPassword
	dbName := config.Env().DBName
	dbSSLMode := config.Env().DBSSLMode
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode,
	)
}

// Migrator the database
func Migrator(logger echo.Logger) {
	logger.Info("Migrating database")

	wd, _ := os.Getwd()

	databaseURL := getDatabaseURL()
	migrationsPath := "file://" + wd + "/migrations"

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		logger.Fatal("failed to create migrator:", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("failed to migrate database:", err)
	}

	logger.Info("Database migrated")
}
