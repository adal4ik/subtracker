package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"subtracker/internal/config"
	"subtracker/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func ConnectDB(ctx context.Context, cfg config.PostgresConfig, logger logger.Logger) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout: failed to connect to DB within the deadline: %w", ctx.Err())
		case <-ticker.C:
			logger.Debug("Attempting to connect to the database")
			if err := db.PingContext(ctx); err == nil {
				logger.Info("Connected to the database successfully", zap.String("dsn", connStr))
				return db, nil
			}
		}
	}
}
