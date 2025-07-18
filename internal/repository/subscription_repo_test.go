package repository

import (
	"context"
	"regexp"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/pkg/logger"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// newTestRepo - хелпер для создания репозитория с моком базы данных
func newTestRepo(t *testing.T) (*SubscriptionRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	// Используем логгер, который ничего не пишет в консоль во время тестов
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

	// Ожидаем точный SQL-запрос. QuoteMeta экранирует спецсимволы.
	query := regexp.QuoteMeta(`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`)
	mock.ExpectExec(query).
		WithArgs(subToCreate.ID, subToCreate.UserID, subToCreate.ServiceName, subToCreate.Price, subToCreate.StartDate, subToCreate.EndDate).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateSubscription(context.Background(), subToCreate)

	assert.NoError(t, err)
	// Убеждаемся, что все ожидания от мока были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListSubscriptions(t *testing.T) {
	repo, mock := newTestRepo(t)
	userID := uuid.New()

	// Готовим строки, которые "вернет" база данных
	rows := sqlmock.NewRows([]string{"id", "user_id", "service_name", "price", "start_date", "end_date"}).
		AddRow(uuid.New(), userID, "Netflix", 1000, time.Now(), nil)

	filter := dto.SubscriptionFilter{
		UserID: userID.String(),
		Limit:  10,
		Offset: 0,
	}

	// Так как SQL-запрос строится динамически, используем гибкий regex,
	// который проверяет только начало и конец запроса.
	expectedQuery := `^SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions WHERE 1=1 .* ORDER BY start_date DESC LIMIT .*`

	mock.ExpectQuery(expectedQuery).
		WithArgs(filter.UserID, filter.Limit, filter.Offset).
		WillReturnRows(rows)

	result, err := repo.ListSubscriptions(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
