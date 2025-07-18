package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
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
	repo, mock := newTestRepo(t)

	subToCreate := dao.SubscriptionRow{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ServiceName: "Netflix",
		Price:       1000,
		StartDate:   time.Now(),
		EndDate:     nil,
	}

	query := regexp.QuoteMeta(`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`)
	mock.ExpectExec(query).
		WithArgs(subToCreate.ID, subToCreate.UserID, subToCreate.ServiceName, subToCreate.Price, subToCreate.StartDate, subToCreate.EndDate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateSubscription(context.Background(), subToCreate)

	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListSubscriptions(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
		AddRow(uuid.New(), userID, "Netflix", 1000, time.Now(), nil)

	filter := dto.SubscriptionFilter{
		UserID: userID.String(),
		Limit:  10,
		Offset: 0,
	}

	expectedQuery := `^SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE 1=1 .* ORDER BY start_date DESC LIMIT .*`

	mock.ExpectQuery(expectedQuery).
		WithArgs(filter.UserID, filter.Limit, filter.Offset).
		WillReturnRows(rows)

	result, err := repo.ListSubscriptions(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
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
		assert.True(t, errors.As(err, &appErr), "error should be of type AppError")

		assert.Equal(t, 404, appErr.Code)

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
		assert.True(t, errors.As(err, &appErr), "error should be of type AppError")

		assert.Equal(t, 500, appErr.Code)

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
