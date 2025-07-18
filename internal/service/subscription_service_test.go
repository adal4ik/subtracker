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
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"testing"
	"time"

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

func TestSubscriptionService_UpdateSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())

		subID := uuid.New()
		userID := uuid.New()
		now := time.Now().Truncate(time.Second)

		subFromHandler := domain.Subscription{
			ID:          subID,
			ServiceName: "New Service Name",
			Price:       999,
			StartDate:   now,
			EndDate:     nil,
		}

		subFromDB := dao.SubscriptionRow{
			ID:          subID,
			UserID:      userID,
			ServiceName: "Old Service Name",
			Price:       500,
			StartDate:   now.AddDate(0, -1, 0),
			EndDate:     &now,
		}

		expectedDAOForUpdate := dao.SubscriptionRow{
			ID:          subID,
			UserID:      userID,
			ServiceName: subFromHandler.ServiceName,
			Price:       subFromHandler.Price,
			StartDate:   subFromHandler.StartDate,
			EndDate:     subFromHandler.EndDate,
		}

		mockRepo.On("GetSubscription", mock.Anything, subID.String()).Return(subFromDB, nil).Once()

		mockRepo.On("UpdateSubscription", mock.Anything, expectedDAOForUpdate).Return(nil).Once()

		err := service.UpdateSubscription(context.Background(), subFromHandler)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetSubscription Fails (Not Found)", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		subID := uuid.New()

		repoErr := apperrors.NewNotFound("not found", nil)
		mockRepo.On("GetSubscription", mock.Anything, subID.String()).Return(dao.SubscriptionRow{}, repoErr).Once()

		err := service.UpdateSubscription(context.Background(), domain.Subscription{ID: subID})

		assert.Error(t, err)
		assert.Equal(t, repoErr, err)

		mockRepo.AssertNotCalled(t, "UpdateSubscription", mock.Anything, mock.Anything)
		mockRepo.AssertExpectations(t)
	})
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		testID := uuid.New().String()

		mockRepo.On("DeleteSubscription", mock.Anything, testID).Return(nil).Once()

		err := service.DeleteSubscription(context.Background(), testID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository Returns Error", func(t *testing.T) {
		mockRepo := new(mocks.SubscriptionRepositoryInterface)
		service := NewSubscriptionService(mockRepo, logger.NewNopLogger())
		testID := uuid.New().String()

		repoErr := apperrors.NewNotFound("not found in repo", nil)
		mockRepo.On("DeleteSubscription", mock.Anything, testID).Return(repoErr).Once()

		err := service.DeleteSubscription(context.Background(), testID)

		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		mockRepo.AssertExpectations(t)
	})
}
