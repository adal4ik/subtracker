package repository

import (
	"database/sql"

	"subtracker/pkg/logger"
)

type Repository struct {
	SubscriptionRepository *SubscriptionRepository
}

func NewRepository(db *sql.DB, logger logger.Logger) *Repository {
	return &Repository{
		NewSubscriptionRepository(db, logger),
	}
}
