package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func newTestRepo(t *testing.T) (*SubscriptionRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	repo := NewSubscriptionRepository(db, logger.NewNopLogger())
	return repo, mock
}

func TestCreateSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		subToCreate := dao.SubscriptionRow{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			ServiceName: "Netflix",
		}
		query := regexp.QuoteMeta(`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`)
		mock.ExpectExec(query).
			WithArgs(subToCreate.ID, subToCreate.UserID, subToCreate.ServiceName, subToCreate.Price, subToCreate.StartDate, subToCreate.EndDate).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateSubscription(context.Background(), subToCreate)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Conflict on Duplicate ID", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		pgErr := &pgconn.PgError{Code: "23505"}
		query := regexp.QuoteMeta(`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`)
		mock.ExpectExec(query).WillReturnError(pgErr)

		err := repo.CreateSubscription(context.Background(), dao.SubscriptionRow{})
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusConflict, appErr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestListSubscriptions(t *testing.T) {
	t.Run("Success with UserID filter", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		userID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
			AddRow(uuid.New(), userID, "Netflix", 1000, time.Now(), nil)
		filter := dto.SubscriptionFilter{
			UserID: userID.String(),
			Limit:  10,
			Offset: 0,
		}
		expectedQuery := regexp.QuoteMeta("SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE user_id = $1 ORDER BY start_date DESC LIMIT 10 OFFSET 0")
		mock.ExpectQuery(expectedQuery).
			WithArgs(filter.UserID).
			WillReturnRows(rows)

		result, err := repo.ListSubscriptions(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with Multiple filters", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		userID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
			AddRow(uuid.New(), userID, "Yandex Plus", 500, time.Now(), nil)
		filter := dto.SubscriptionFilter{
			UserID:      userID.String(),
			ServiceName: "Yandex Plus",
			MinPrice:    300,
			Limit:       5,
			Offset:      0,
		}
		expectedQuery := regexp.QuoteMeta("SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE user_id = $1 AND service_name = $2 AND price >= $3 ORDER BY start_date DESC LIMIT 5 OFFSET 0")
		mock.ExpectQuery(expectedQuery).
			WithArgs(filter.UserID, filter.ServiceName, filter.MinPrice).
			WillReturnRows(rows)

		result, err := repo.ListSubscriptions(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with No Filters (Pagination only)", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"})
		filter := dto.SubscriptionFilter{Limit: 20, Offset: 10}
		expectedQuery := regexp.QuoteMeta("SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions ORDER BY start_date DESC LIMIT 20 OFFSET 10")
		mock.ExpectQuery(expectedQuery).
			WithArgs(). // Аргументов нет
			WillReturnRows(rows)

		result, err := repo.ListSubscriptions(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		expectedID := uuid.New()
		expectedRow := dao.SubscriptionRow{ID: expectedID}
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
			AddRow(expectedRow.ID, uuid.New(), "Netflix", 100, time.Now(), nil)
		query := regexp.QuoteMeta(`SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE id = $1`)
		mock.ExpectQuery(query).WithArgs(expectedID.String()).WillReturnRows(rows)
		result, err := repo.GetSubscription(context.Background(), expectedID.String())
		assert.NoError(t, err)
		assert.Equal(t, expectedRow.ID, result.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Not Found", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		testID := uuid.New().String()
		query := regexp.QuoteMeta(`SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE id = $1`)
		mock.ExpectQuery(query).WithArgs(testID).WillReturnError(sql.ErrNoRows)
		_, err := repo.GetSubscription(context.Background(), testID)
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusNotFound, appErr.Code)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Other DB Error", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		testID := uuid.New().String()
		dbErr := errors.New("connection failed")
		query := regexp.QuoteMeta(`SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE id = $1`)
		mock.ExpectQuery(query).WithArgs(testID).WillReturnError(dbErr)
		_, err := repo.GetSubscription(context.Background(), testID)
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusInternalServerError, appErr.Code)
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateSubscription(t *testing.T) {
	ctx := context.Background()
	t.Run("Success", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		subToUpdate := dao.SubscriptionRow{
			ID:          uuid.New(),
			ServiceName: "Updated Service",
			Price:       999,
		}
		query := regexp.QuoteMeta(`UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 WHERE id = $5`)
		mock.ExpectExec(query).
			WithArgs(subToUpdate.ServiceName, subToUpdate.Price, subToUpdate.StartDate, subToUpdate.EndDate, subToUpdate.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.UpdateSubscription(ctx, subToUpdate)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Not Found", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		subToUpdate := dao.SubscriptionRow{ID: uuid.New()}
		query := regexp.QuoteMeta(`UPDATE subscriptions SET service_name = $1, price = $2, start_date = $3, end_date = $4 WHERE id = $5`)
		mock.ExpectExec(query).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), subToUpdate.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err := repo.UpdateSubscription(ctx, subToUpdate)
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusNotFound, appErr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		testID := uuid.New().String()
		query := regexp.QuoteMeta(`DELETE FROM subscriptions WHERE id = $1`)
		mock.ExpectExec(query).WithArgs(testID).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.DeleteSubscription(context.Background(), testID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Not Found", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		testID := uuid.New().String()
		query := regexp.QuoteMeta(`DELETE FROM subscriptions WHERE id = $1`)
		mock.ExpectExec(query).WithArgs(testID).WillReturnResult(sqlmock.NewResult(0, 0))
		err := repo.DeleteSubscription(context.Background(), testID)
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusNotFound, appErr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("DB Error", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		testID := uuid.New().String()
		dbErr := errors.New("connection broken")
		query := regexp.QuoteMeta(`DELETE FROM subscriptions WHERE id = $1`)
		mock.ExpectExec(query).WithArgs(testID).WillReturnError(dbErr)
		err := repo.DeleteSubscription(context.Background(), testID)
		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusInternalServerError, appErr.Code)
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
func TestListForCostCalculation(t *testing.T) {
	t.Run("Success with Full Filter", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		userID := uuid.New()
		filter := dto.CostFilter{
			UserID:      userID.String(),
			ServiceName: "Netflix",
			PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		}
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
			AddRow(uuid.New(), userID, "Netflix", 100, time.Now(), nil)

		expectedQuery := regexp.QuoteMeta("SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE user_id = $1 AND service_name = $2 AND start_date <= $3 AND (end_date IS NULL OR end_date >= $4)")

		mock.ExpectQuery(expectedQuery).
			WithArgs(filter.UserID, filter.ServiceName, filter.PeriodEnd, filter.PeriodStart).
			WillReturnRows(rows)

		result, err := repo.ListForCostCalculation(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with UserID only", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		userID := uuid.New()
		filter := dto.CostFilter{
			UserID:      userID.String(),
			PeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		}
		rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
			AddRow(uuid.New(), userID, "Netflix", 100, time.Now(), nil).
			AddRow(uuid.New(), userID, "Spotify", 200, time.Now(), nil)

		expectedQuery := regexp.QuoteMeta("SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE user_id = $1 AND start_date <= $2 AND (end_date IS NULL OR end_date >= $3)")

		mock.ExpectQuery(expectedQuery).
			WithArgs(filter.UserID, filter.PeriodEnd, filter.PeriodStart).
			WillReturnRows(rows)

		result, err := repo.ListForCostCalculation(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error on Query", func(t *testing.T) {
		repo, mock := newTestRepo(t)
		dbErr := errors.New("something went wrong")
		filter := dto.CostFilter{
			UserID:      uuid.New().String(),
			PeriodStart: time.Now(),
			PeriodEnd:   time.Now(),
		}

		mock.ExpectQuery(".*").WillReturnError(dbErr)

		_, err := repo.ListForCostCalculation(context.Background(), filter)

		assert.Error(t, err)
		var appErr *apperrors.AppError
		assert.True(t, errors.As(err, &appErr))
		assert.Equal(t, http.StatusInternalServerError, appErr.Code)
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
