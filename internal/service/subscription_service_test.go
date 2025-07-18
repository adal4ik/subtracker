package service

import (
	"context"
	"errors"
	"subtracker/internal/domain"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/repository/mocks" // <-- Путь к вашему сгенерированному моку
	"subtracker/pkg/logger"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubscriptionService_CreateSubscription(t *testing.T) {
	t.Run("Success - Generates ID", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())

		// Создаем подписку без ID
		subDomain := domain.Subscription{UserID: uuid.New(), ServiceName: "Yandex Plus"}

		// mock.MatchedBy позволяет нам проверить аргумент с помощью функции.
		// Мы убедимся, что ID был сгенерирован и он не пустой (uuid.Nil).
		mockRepo.On("CreateSubscription", mock.Anything, mock.MatchedBy(func(d dao.SubscriptionRow) bool {
			return d.ID != uuid.Nil && d.UserID == subDomain.UserID
		})).Return(nil).Once()

		err := service.CreateSubscription(context.Background(), subDomain)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		dbError := errors.New("repository error")

		mockRepo.On("CreateSubscription", mock.Anything, mock.AnythingOfType("dao.SubscriptionRow")).
			Return(dbError).Once()

		err := service.CreateSubscription(context.Background(), domain.Subscription{})

		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

// --- Тесты для ListSubscriptions ---

func TestSubscriptionService_ListSubscriptions(t *testing.T) {
	t.Run("Success - With Results", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())

		filter := dto.SubscriptionFilter{Limit: 10, Offset: 0}
		mockDAOList := []dao.SubscriptionRow{
			{ID: uuid.New(), ServiceName: "Netflix"},
			{ID: uuid.New(), ServiceName: "Spotify"},
		}
		expectedDomainList := []domain.Subscription{
			mapper.ToDomainFromDAO(mockDAOList[0]),
			mapper.ToDomainFromDAO(mockDAOList[1]),
		}

		mockRepo.On("ListSubscriptions", mock.Anything, filter).Return(mockDAOList, nil).Once()

		result, err := service.ListSubscriptions(context.Background(), filter)

		assert.NoError(t, err)
		assert.Equal(t, expectedDomainList, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - No Results", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		filter := dto.SubscriptionFilter{}

		// Репозиторий возвращает пустой срез и nil ошибку
		mockRepo.On("ListSubscriptions", mock.Anything, filter).Return([]dao.SubscriptionRow{}, nil).Once()

		result, err := service.ListSubscriptions(context.Background(), filter)

		assert.NoError(t, err)
		assert.NotNil(t, result) // Убеждаемся, что срез не nil
		assert.Len(t, result, 0) // А именно пустой
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		dbError := errors.New("db connection failed")

		mockRepo.On("ListSubscriptions", mock.Anything, mock.AnythingOfType("dto.SubscriptionFilter")).
			Return(nil, dbError).Once()

		result, err := service.ListSubscriptions(context.Background(), dto.SubscriptionFilter{})

		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}
