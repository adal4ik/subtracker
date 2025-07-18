package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error
	ListSubscriptions(ctx context.Context, subFilter dto.SubscriptionFilter) ([]dao.SubscriptionRow, error)
	GetSubscription(ctx context.Context, id string) (dao.SubscriptionRow, error)
	UpdateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error
	DeleteSubscription(ctx context.Context, id string) error
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
	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, subDao.ID, subDao.UserID, subDao.ServiceName, subDao.Price, subDao.StartDate, subDao.EndDate)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return apperrors.New(http.StatusConflict, "subscription with this ID already exists", err)
		}
		r.logger.Error("Failed to create subscription in database", zap.Error(err))
		return apperrors.NewInternalServerError("database error on create", err)
	}
	return nil
}
func (r *SubscriptionRepository) ListSubscriptions(ctx context.Context, f dto.SubscriptionFilter) ([]dao.SubscriptionRow, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if f.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, f.UserID)
		argIdx++
	}
	if f.ServiceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, f.ServiceName)
		argIdx++
	}
	if f.MinPrice > 0 {
		query += fmt.Sprintf(" AND price >= $%d", argIdx)
		args = append(args, f.MinPrice)
		argIdx++
	}
	if f.MaxPrice > 0 {
		query += fmt.Sprintf(" AND price <= $%d", argIdx)
		args = append(args, f.MaxPrice)
		argIdx++
	}
	if f.StartDate != "" {
		query += fmt.Sprintf(" AND start_date >= $%d", argIdx)
		args = append(args, f.StartDate)
		argIdx++
	}
	if f.EndDate != "" {
		query += fmt.Sprintf(" AND end_date <= $%d", argIdx)
		args = append(args, f.EndDate)
		argIdx++
	}
	if f.HasEndDate != nil {
		if *f.HasEndDate {
			query += " AND end_date IS NOT NULL"
		} else {
			query += " AND end_date IS NULL"
		}
	}

	query += fmt.Sprintf(" ORDER BY start_date DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list subscriptions", zap.Error(err))
		return nil, apperrors.NewInternalServerError("database error on list", err)
	}
	defer rows.Close()

	var result []dao.SubscriptionRow
	for rows.Next() {
		var sub dao.SubscriptionRow
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate); err != nil {
			r.logger.Error("Failed to scan subscription row", zap.Error(err))
			return nil, apperrors.NewInternalServerError("database error on scan", err)
		}
		result = append(result, sub)
	}
	return result, nil
}

func (r *SubscriptionRepository) GetSubscription(ctx context.Context, id string) (dao.SubscriptionRow, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var sub dao.SubscriptionRow
	if err := row.Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Subscription not found in DB", zap.String("id", id))
			return dao.SubscriptionRow{}, apperrors.NewNotFound("subscription not found", err)
		}
		r.logger.Error("Failed to get subscription from DB", zap.Error(err), zap.String("id", id))
		return dao.SubscriptionRow{}, apperrors.NewInternalServerError("database error", err)
	}

	return sub, nil
}
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error {
	// Implementation for updating a subscription
	return nil
}
func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, id string) error {
	// Implementation for deleting a subscription
	return nil
}
