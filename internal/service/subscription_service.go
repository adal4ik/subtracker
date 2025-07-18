package service

import (
	"context"
	"subtracker/internal/domain"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/repository"
	"subtracker/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionServiceInterface interface {
	CreateSubscription(ctx context.Context, subDomain domain.Subscription) error
	ListSubscriptions(ctx context.Context, filter dto.SubscriptionFilter) ([]domain.Subscription, error)
	GetSubscription(ctx context.Context, id string) (domain.Subscription, error)
	UpdateSubscription(ctx context.Context, subDomain domain.Subscription) error
	DeleteSubscription(ctx context.Context, id string) error
}

type SubscriptionService struct {
	repo   repository.SubscriptionRepositoryInterface
	logger logger.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepositoryInterface, logger logger.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, subDomain domain.Subscription) error {
	s.logger.Debug("Creating subscription", zap.String("service_name", subDomain.ServiceName),
		zap.Int("price", subDomain.Price),
		zap.String("user_id", subDomain.UserID.String()),
		zap.Time("start_date", subDomain.StartDate),
		zap.Any("end_date", subDomain.EndDate),
	)
	if subDomain.ID == uuid.Nil {
		subDomain.ID = uuid.New()
	}
	subDao := mapper.ToDAOFromDomain(subDomain)
	return s.repo.CreateSubscription(ctx, subDao)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, filter dto.SubscriptionFilter) ([]domain.Subscription, error) {
	s.logger.Debug("Filtering subscriptions", zap.String("user_id", filter.UserID),
		zap.String("service_name", filter.ServiceName),
		zap.String("start_date", filter.StartDate),
		zap.String("end_date", filter.EndDate),
		zap.Int("limit", filter.Limit),
		zap.Int("offset", filter.Offset),
	)
	subscriptions, err := s.repo.ListSubscriptions(ctx, filter)
	if err != nil {
		return nil, err
	}
	subDomainList := make([]domain.Subscription, len(subscriptions))
	for i, sub := range subscriptions {
		subDomainList[i] = mapper.ToDomainFromDAO(sub)
	}
	return subDomainList, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id string) (domain.Subscription, error) {
	subDao, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return domain.Subscription{}, err
	}
	return mapper.ToDomainFromDAO(subDao), nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, subToUpdate domain.Subscription) error {
	s.logger.Debug("Attempting to update subscription", zap.String("id", subToUpdate.ID.String()))

	existingSubDAO, err := s.repo.GetSubscription(ctx, subToUpdate.ID.String())
	if err != nil {
		return err
	}

	finalSubDAO := dao.SubscriptionRow{
		ID:          existingSubDAO.ID,
		UserID:      existingSubDAO.UserID,
		ServiceName: subToUpdate.ServiceName,
		Price:       subToUpdate.Price,
		StartDate:   subToUpdate.StartDate,
		EndDate:     subToUpdate.EndDate,
	}

	return s.repo.UpdateSubscription(ctx, finalSubDAO)
}
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id string) error {
	s.logger.Debug("Deleting subscription", zap.String("id", id))
	return s.repo.DeleteSubscription(ctx, id)
}
