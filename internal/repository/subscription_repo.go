package repository

import (
	"context"
	"database/sql"
	"subtracker/internal/domain/dao"
	"subtracker/pkg/logger"

	"go.uber.org/zap"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error
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
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error {
	r.logger.Debug("Creating subscription in repository", zap.String("service_name", subDao.ServiceName),
		zap.Int("price", subDao.Price),
		zap.String("user_id", subDao.UserID.String()),
		zap.Time("start_date", subDao.StartDate),
		zap.Any("end_date", subDao.EndDate),
	)
	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
            VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, subDao.ID, subDao.UserID, subDao.ServiceName, subDao.Price, subDao.StartDate, subDao.EndDate)
	if err != nil {
		r.logger.Error("Failed to create subscription in database", zap.Error(err),
			zap.String("service_name", subDao.ServiceName),
			zap.Int("price", subDao.Price),
			zap.String("user_id", subDao.UserID.String()),
		)
		return err
	}
	return nil
}
