package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type SubscriptionRepositoryInterface interface {
	CreateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error
	ListSubscriptions(ctx context.Context, subFilter dto.SubscriptionFilter) ([]dao.SubscriptionRow, error)
	GetSubscription(ctx context.Context, id string) (dao.SubscriptionRow, error)
	UpdateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error
	DeleteSubscription(ctx context.Context, id string) error
	ListForCostCalculation(ctx context.Context, filter dto.CostFilter) ([]dao.SubscriptionRow, error)
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
	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`
	r.logger.Debug("Executing CreateSubscription query",
		zap.String("sql", query),
		zap.String("subscription_id", subDao.ID.String()),
		zap.String("user_id", subDao.UserID.String()),
	)
	_, err := r.db.ExecContext(ctx, query, subDao.ID, subDao.UserID, subDao.ServiceName, subDao.Price, subDao.StartDate, subDao.EndDate)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.logger.Warn("Create subscription conflict: unique constraint violation",
				zap.String("subscription_id", subDao.ID.String()),
				zap.Error(err),
			)
			return apperrors.New(http.StatusConflict, "subscription with this ID already exists", err)
		}
		r.logger.Error("Failed to create subscription in database", zap.Error(err))
		return apperrors.NewInternalServerError("database error on create", err)
	}
	return nil
}

func (r *SubscriptionRepository) ListSubscriptions(ctx context.Context, f dto.SubscriptionFilter) ([]dao.SubscriptionRow, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryBuilder := psql.Select("id", "user_id", "service_name", "price", "start_date", "end_date").
		From("subscriptions")

	if f.UserID != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"user_id": f.UserID})
	}
	if f.ServiceName != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"service_name": f.ServiceName})
	}
	if f.MinPrice > 0 {
		queryBuilder = queryBuilder.Where(sq.GtOrEq{"price": f.MinPrice})
	}
	if f.MaxPrice > 0 {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{"price": f.MaxPrice})
	}
	if f.StartDate != "" {

		queryBuilder = queryBuilder.Where(sq.GtOrEq{"start_date": f.StartDate})
	}
	if f.EndDate != "" {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{"end_date": f.EndDate})
	}
	if f.HasEndDate != nil {
		if *f.HasEndDate {
			queryBuilder = queryBuilder.Where(sq.NotEq{"end_date": nil})
		} else {
			queryBuilder = queryBuilder.Where(sq.Eq{"end_date": nil})
		}
	}
	queryBuilder = queryBuilder.OrderBy("start_date DESC").
		Limit(uint64(f.Limit)).
		Offset(uint64(f.Offset))

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		r.logger.Error("Failed to build SQL query for ListSubscriptions", zap.Error(err))
		return nil, apperrors.NewInternalServerError("failed to build list query", err)
	}

	r.logger.Debug("Executing ListSubscriptions", zap.String("sql", sql), zap.Any("args", args))

	rows, err := r.db.QueryContext(ctx, sql, args...)
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
	r.logger.Debug("Executing GetSubscription query",
		zap.String("sql", query),
		zap.String("id", id),
	)
	var sub dao.SubscriptionRow
	if err := row.Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate); err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Subscription not found in DB", zap.String("id", id))
			return dao.SubscriptionRow{}, apperrors.NewNotFound("subscription not found", err)
		}

		r.logger.Error("Failed to scan/get subscription from DB", zap.Error(err), zap.String("id", id))
		return dao.SubscriptionRow{}, apperrors.NewInternalServerError("database error on get", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, subDao dao.SubscriptionRow) error {
	query := `UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 WHERE id = $5`

	r.logger.Debug("Executing UpdateSubscription query",
		zap.String("sql", query),
		zap.String("id", subDao.ID.String()),
	)

	result, err := r.db.ExecContext(ctx, query, subDao.ServiceName, subDao.Price, subDao.StartDate, subDao.EndDate, subDao.ID)
	if err != nil {
		r.logger.Error("Failed to execute update query", zap.Error(err), zap.String("id", subDao.ID.String()))
		return apperrors.NewInternalServerError("database error on update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected after update", zap.Error(err), zap.String("id", subDao.ID.String()))
		return apperrors.NewInternalServerError("database error on update result", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Update attempt on non-existent subscription", zap.String("id", subDao.ID.String()))
		return apperrors.NewNotFound("subscription to update not found", nil)
	}

	return nil
}

func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	r.logger.Debug("Executing DeleteSubscription query",
		zap.String("sql", query),
		zap.String("id", id),
	)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to execute delete query", zap.Error(err), zap.String("id", id))
		return apperrors.NewInternalServerError("database error on delete", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected after delete", zap.Error(err), zap.String("id", id))
		return apperrors.NewInternalServerError("database error on delete result", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Delete attempt on non-existent subscription", zap.String("id", id))
		return apperrors.NewNotFound("subscription to delete not found", nil)
	}

	return nil
}

func (r *SubscriptionRepository) ListForCostCalculation(ctx context.Context, filter dto.CostFilter) ([]dao.SubscriptionRow, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	queryBuilder := psql.Select("id", "user_id", "service_name", "price", "start_date", "end_date").
		From("subscriptions")

	queryBuilder = queryBuilder.Where(sq.Eq{"user_id": filter.UserID})
	if filter.ServiceName != "" {
		queryBuilder = queryBuilder.Where(sq.Eq{"service_name": filter.ServiceName})
	}
	queryBuilder = queryBuilder.Where(sq.LtOrEq{"start_date": filter.PeriodEnd}).
		Where(sq.Or{
			sq.Eq{"end_date": nil},
			sq.GtOrEq{"end_date": filter.PeriodStart},
		})

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		r.logger.Error("Failed to build SQL for ListForCostCalculation", zap.Error(err))
		return nil, apperrors.NewInternalServerError("failed to build cost query", err)
	}

	r.logger.Debug("Executing ListForCostCalculation query", zap.String("sql", sql), zap.Any("args", args))

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to execute cost calculation query", zap.Error(err))
		return nil, apperrors.NewInternalServerError("database error on cost calculation", err)
	}
	defer rows.Close()

	var result []dao.SubscriptionRow
	for rows.Next() {
		var sub dao.SubscriptionRow
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate); err != nil {
			r.logger.Error("Failed to scan subscription row for cost", zap.Error(err))
			return nil, apperrors.NewInternalServerError("database error on scan for cost", err)
		}
		result = append(result, sub)
	}
	return result, nil
}
