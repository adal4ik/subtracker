package service

import (
	"context"
	"database/sql"
	"errors"
	"subtracker/internal/domain"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/repository/mocks"
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

		subDomain := domain.Subscription{UserID: uuid.New(), ServiceName: "Yandex Plus"}
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

		mockRepo.On("ListSubscriptions", mock.Anything, filter).Return([]dao.SubscriptionRow{}, nil).Once()

		result, err := service.ListSubscriptions(context.Background(), filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
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

func TestSubscriptionService_GetSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())

		testID := uuid.New().String()
		mockDAO := dao.SubscriptionRow{
			ID:          uuid.MustParse(testID),
			ServiceName: "Netflix",
		}
		expectedDomain := mapper.ToDomainFromDAO(mockDAO)

		mockRepo.On("GetSubscription", mock.Anything, testID).Return(mockDAO, nil).Once()

		result, err := service.GetSubscription(context.Background(), testID)

		assert.NoError(t, err)
		assert.Equal(t, expectedDomain, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found in Repo", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		testID := uuid.New().String()

		mockRepo.On("GetSubscription", mock.Anything, testID).
			Return(dao.SubscriptionRow{}, sql.ErrNoRows).Once()

		_, err := service.GetSubscription(context.Background(), testID)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Other Repo Error", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		testID := uuid.New().String()
		repoErr := errors.New("some other db error")

		mockRepo.On("GetSubscription", mock.Anything, testID).
			Return(dao.SubscriptionRow{}, repoErr).Once()

		_, err := service.GetSubscription(context.Background(), testID)

		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		mockRepo.AssertExpectations(t)
	})
}
