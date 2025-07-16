package repository

import (
	"database/sql"
	"subtracker/pkg/logger"
)

type SubscriptionRepositoryInterface interface {
	// Define methods that SubscriptionRepository should implement
}

type SubscriptionRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewSubscriptionRepository(db *sql.DB, logger logger.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger,
	}
}
